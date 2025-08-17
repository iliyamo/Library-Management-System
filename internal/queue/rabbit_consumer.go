package queue

import (
	"encoding/json"
	"log"
	"time"

	"github.com/iliyamo/Library-Management-System/internal/model"
	"github.com/streadway/amqp"
)

// StartRabbitConsumer connects to RabbitMQ and consumes messages from the loan_events queue.
// Compatible with handler functions that do not return an error.
func StartRabbitConsumer(amqpURL string, handlerFunc func(evt model.LoanEvent)) error {
	conn, err := connectWithRetry(amqpURL, 5, 2*time.Second) // ✅ use helper
	if err != nil {
		return err
	}
	ch, err := conn.Channel()
	if err != nil {
		return err
	}

	if err := ch.Qos(10, 0, false); err != nil {
		return err
	}

	msgs, err := ch.Consume(
		"loan_events",
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	if rabbitLogger != nil {
		rabbitLogger.Println("✅ RabbitMQ consumer started for loan_events (manual ack mode)")
	} else {
		log.Println("✅ RabbitMQ consumer started for loan_events (manual ack mode)")
	}

	go func() {
		for d := range msgs {
			var evt model.LoanEvent
			if err := json.Unmarshal(d.Body, &evt); err != nil {
				if rabbitLogger != nil {
					rabbitLogger.Printf("[RabbitConsumer] invalid event: %v", err)
				} else {
					log.Printf("[RabbitConsumer] invalid event: %v", err)
				}
				_ = d.Ack(false)
				continue
			}

			if rabbitLogger != nil {
				rabbitLogger.Printf("[RabbitConsumer] received event: %+v", evt)
			}

			handlerFunc(evt)

			if err := d.Ack(false); err != nil {
				log.Printf("[RabbitConsumer] ack failed: %v", err)
			}
		}
	}()

	return nil
}

// StartRabbitConsumerWithErr connects to RabbitMQ and consumes messages, using a handler that can return an error.
// In case of error, the message is Nack'ed and requeued.
func StartRabbitConsumerWithErr(amqpURL string, handlerFunc func(evt model.LoanEvent) error) error {
	conn, err := connectWithRetry(amqpURL, 5, 2*time.Second) // ✅ use helper
	if err != nil {
		return err
	}
	ch, err := conn.Channel()
	if err != nil {
		return err
	}

	if err := ch.Qos(10, 0, false); err != nil {
		return err
	}

	msgs, err := ch.Consume(
		"loan_events",
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	if rabbitLogger != nil {
		rabbitLogger.Println("✅ RabbitMQ consumer started for loan_events (manual ack mode, error-aware)")
	} else {
		log.Println("✅ RabbitMQ consumer started for loan_events (manual ack mode, error-aware)")
	}

	go func() {
		for d := range msgs {
			var evt model.LoanEvent
			if err := json.Unmarshal(d.Body, &evt); err != nil {
				_ = d.Ack(false)
				continue
			}

			if err := handlerFunc(evt); err != nil {
				if rabbitLogger != nil {
					rabbitLogger.Printf("[RabbitConsumer] handler failed: %v", err)
				} else {
					log.Printf("[RabbitConsumer] handler failed: %v", err)
				}
				_ = d.Nack(false, true)
				continue
			}

			_ = d.Ack(false)
		}
	}()

	return nil
}

// connectWithRetry tries to connect to RabbitMQ with retry attempts.
func connectWithRetry(amqpURL string, maxRetries int, delay time.Duration) (*amqp.Connection, error) {
	var conn *amqp.Connection
	var err error
	for i := 0; i < maxRetries; i++ {
		conn, err = amqp.Dial(amqpURL)
		if err == nil {
			return conn, nil
		}
		log.Printf("[RabbitConsumer] connection failed (attempt %d/%d): %v", i+1, maxRetries, err)
		time.Sleep(delay)
	}
	return nil, err
}
