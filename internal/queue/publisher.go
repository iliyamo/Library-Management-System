// internal/queue/publisher.go
//
// Abstraction for publishing messages to RabbitMQ (preferred) or Redis fallback.
// Supports both domain Events (LoanEvent) and Commands (LoanCommand).
// This version ensures persistent message publishing and logs errors encountered during publishing.

package queue

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/streadway/amqp"

	"github.com/iliyamo/Library-Management-System/internal/model"
	"github.com/iliyamo/Library-Management-System/internal/utils"
)

var useRabbit bool

// InitQueue tries to init RabbitMQ if RABBITMQ_URL is set.
// Falls back silently to Redis when it can't.
func InitQueue() {
	amqpURL := os.Getenv("RABBITMQ_URL")
	if amqpURL == "" {
		return
	}
	if err := InitRabbitMQ(amqpURL); err != nil {
		log.Printf("[Queue] Failed to initialise RabbitMQ: %v", err)
	} else {
		useRabbit = true
		log.Printf("[Queue] RabbitMQ initialised")
	}
}

// UsingRabbit returns whether RabbitMQ is being used (exported for use in main).
func UsingRabbit() bool {
	return useRabbit
}

// Publish sends a raw string to a queue/channel.
// With RabbitMQ it publishes plain text; otherwise it uses Redis Pub/Sub.
func Publish(queueName string, message string) error {
	if useRabbit {
		if rabbitClient == nil || rabbitClient.Channel == nil {
			log.Printf("[Queue] RabbitMQ requested but client not initialised; falling back to Redis")
		} else {
			publishing := amqp.Publishing{
				ContentType:  "text/plain",
				DeliveryMode: amqp.Persistent, // ensure message survives broker restart
				Timestamp:    time.Now(),
				Body:         []byte(message),
			}
			if err := rabbitClient.Channel.Publish(
				"",        // default exchange
				queueName, // queue as routing key
				false,     // mandatory
				false,     // immediate (deprecated, kept false)
				publishing,
			); err != nil {
				log.Printf("[Queue] publish failed to %s via RabbitMQ: %v", queueName, err)
				// Try Redis as a soft fallback too
				return publishJSONRedis(queueName, []byte(message))
			}
			return nil
		}
	}
	// Fallback to Redis if available; if Redis is nil, it's a no-op.
	if utils.RedisClient == nil {
		return nil
	}
	return utils.RedisClient.Publish(context.Background(), queueName, message).Err()
}

// --- Helpers for JSON publishing ---

// publishJSONRabbit publishes a JSON payload to a RabbitMQ queue.
func publishJSONRabbit(queueName string, payload []byte) error {
	if rabbitClient == nil || rabbitClient.Channel == nil {
		// Fall back to the text path (which may go to Redis)
		return Publish(queueName, string(payload))
	}
	publishing := amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent, // durable message
		Timestamp:    time.Now(),
		Body:         payload,
	}
	if err := rabbitClient.Channel.Publish(
		"",
		queueName,
		false,
		false,
		publishing,
	); err != nil {
		log.Printf("[Queue] JSON publish failed to %s via RabbitMQ: %v", queueName, err)
		// Optionally mirror to Redis so we don't drop the message when broker is flaky
		return publishJSONRedis(queueName, payload)
	}
	return nil
}

// publishJSONRedis publishes a JSON payload to a Redis channel.
func publishJSONRedis(channel string, payload []byte) error {
	if utils.RedisClient == nil {
		return nil
	}
	return utils.RedisClient.
		Publish(context.Background(), channel, string(payload)).
		Err()
}

// --- Domain Event publishing ---

// PublishEvent marshals a LoanEvent and publishes it.
// RabbitMQ -> queue "loan_events" (JSON)
// Redis    -> channel "loan_events" (stringified JSON)
func PublishEvent(event model.LoanEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	if useRabbit {
		return publishJSONRabbit(LoanEventsQueue, data)
	}
	return publishJSONRedis(LoanEventsQueue, data)
}

// --- Command publishing ---

// PublishLoanCommand marshals and publishes a LoanCommand.
// RabbitMQ -> queue "loan_commands" (JSON)
// Redis    -> channel "loan_commands" (stringified JSON)
func PublishLoanCommand(cmd model.LoanCommand) error {
	data, err := json.Marshal(cmd)
	if err != nil {
		return err
	}
	if useRabbit {
		return publishJSONRabbit(LoanCommandsQueue, data)
	}
	return publishJSONRedis(LoanCommandsQueue, data)
}

// Backward-compatible alias (در صورتی که جایی هنوز این اسم را صدا می‌زند)
func PublishCommand(cmd model.LoanCommand) error {
	return PublishLoanCommand(cmd)
}

// Backward-compatible alias for older call-sites that used LoanRequestCommand.
// نوع پارامتر را هم به LoanCommand تغییر دادیم تا نیاز به مدل قدیمی نباشد.
func PublishLoanRequestCommand(cmd model.LoanCommand) error {
	return PublishLoanCommand(cmd)
}
