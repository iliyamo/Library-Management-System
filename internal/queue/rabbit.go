// internal/queue/rabbit.go
//
// Minimal RabbitMQ wrapper: holds a connection/channel, declares the required
// queues, and provides helpers to publish JSON/text messages.

package queue

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/streadway/amqp"

	"github.com/iliyamo/Library-Management-System/internal/model"
)

// RabbitMQClient holds the AMQP connection and channel used for publishing messages.
type RabbitMQClient struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
}

// package-level singletons
var (
	rabbitClient  *RabbitMQClient
	rabbitLogger  *log.Logger
	rabbitLogFile *os.File
)

// InitRabbitMQ establishes a connection to RabbitMQ and declares durable queues.
// If already initialised, it’s a no-op.
func InitRabbitMQ(amqpURL string) error {
	if rabbitClient != nil {
		return nil
	}

	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare required queues (durable)
	if err := declareQueues(ch); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return err
	}

	rabbitClient = &RabbitMQClient{Conn: conn, Channel: ch}

	// Dedicated logger → stdout + file
	if rabbitLogger == nil {
		if f, errf := os.OpenFile("rabbitmq.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); errf == nil {
			rabbitLogFile = f
			multi := io.MultiWriter(os.Stdout, f)
			rabbitLogger = log.New(multi, "[RabbitMQ] ", log.LstdFlags)
		} else {
			rabbitLogger = log.New(os.Stdout, "[RabbitMQ] ", log.LstdFlags)
		}
	}
	if rabbitLogger != nil {
		rabbitLogger.Printf("✅ RabbitMQ connected; queues declared: %q, %q", LoanEventsQueue, LoanCommandsQueue)
	} else {
		log.Printf("✅ RabbitMQ connected; queues declared: %q, %q", LoanEventsQueue, LoanCommandsQueue)
	}
	return nil
}

// declareQueues ensures the needed queues exist (durable, non-autoDelete).
func declareQueues(ch *amqp.Channel) error {
	queues := []string{LoanEventsQueue, LoanCommandsQueue}
	for _, q := range queues {
		if _, err := ch.QueueDeclare(
			q,
			true,  // durable
			false, // autoDelete
			false, // exclusive
			false, // noWait
			nil,   // args
		); err != nil {
			return fmt.Errorf("failed to declare queue %q: %w", q, err)
		}
	}
	return nil
}

// PublishToRabbit publishes raw bytes to a specific queue (direct default exchange).
func PublishToRabbit(queue string, body []byte, contentType string) error {
	if rabbitClient == nil || rabbitClient.Channel == nil {
		return fmt.Errorf("RabbitMQ is not initialised")
	}
	if contentType == "" {
		contentType = "text/plain"
	}
	if rabbitLogger != nil {
		rabbitLogger.Printf("publishing to %s (%d bytes)", queue, len(body))
	}
	return rabbitClient.Channel.Publish(
		"",
		queue,
		false,
		false,
		amqp.Publishing{
			ContentType: contentType,
			Body:        body,
		},
	)
}

// PublishLoanEvent marshals and publishes a LoanEvent to the loan_events queue.
func PublishLoanEvent(event model.LoanEvent) error {
	b, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}
	if rabbitLogger != nil {
		rabbitLogger.Printf("publishing event to %s: %+v", LoanEventsQueue, event)
	}
	return PublishToRabbit(LoanEventsQueue, b, "application/json")
}

// CloseRabbitMQ closes channel/connection and the dedicated log file.
func CloseRabbitMQ() {
	if rabbitClient != nil {
		if rabbitClient.Channel != nil {
			_ = rabbitClient.Channel.Close()
		}
		if rabbitClient.Conn != nil {
			_ = rabbitClient.Conn.Close()
		}
		rabbitClient = nil
	}
	if rabbitLogFile != nil {
		_ = rabbitLogFile.Close()
		rabbitLogFile = nil
	}
	rabbitLogger = nil
}
