// internal/queue/rabbit_consumer.go
package queue

import (
	"encoding/json"
	"log"

	"github.com/iliyamo/go-learning/internal/model"
	"github.com/streadway/amqp"
)

// StartRabbitConsumer connects to RabbitMQ and consumes messages from the loan_events queue.
func StartRabbitConsumer(amqpURL string, handlerFunc func(evt model.LoanEvent)) error {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return err
	}
	ch, err := conn.Channel()
	if err != nil {
		return err
	}

	msgs, err := ch.Consume(
		"loan_events", // queue
		"",            // consumer
		true,          // auto-ack
		false,         // exclusive
		false,         // no-local
		false,         // no-wait
		nil,
	)
	if err != nil {
		return err
	}

	go func() {
		for d := range msgs {
			var evt model.LoanEvent
			if err := json.Unmarshal(d.Body, &evt); err != nil {
				log.Printf("[RabbitConsumer] Invalid event: %v", err)
				continue
			}
			handlerFunc(evt)
		}
	}()
	log.Println("✅ RabbitMQ consumer started for loan_events")
	return nil
}
