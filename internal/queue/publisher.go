// internal/queue/publisher.go
//
// This file provides a simple abstraction for publishing messages to either
// RabbitMQ or Redis.  If the environment variable RABBITMQ_URL is set and
// a connection to RabbitMQ is successfully initialised via InitQueue, then
// events will be published to RabbitMQ; otherwise the system falls back to
// Redis using the utils.RedisClient.  The goal is to give new Go learners
// a working example of how to integrate a message broker without breaking
// existing functionality.

package queue

import (
    "context"
    "encoding/json"
    "log"
    "os"

    "github.com/streadway/amqp"

    "github.com/iliyamo/Library-Management-System/internal/model"
    "github.com/iliyamo/Library-Management-System/internal/utils"
)

// useRabbit is set to true when InitQueue successfully initialises a RabbitMQ
// connection.  When false, the Publish functions fall back to Redis.
var useRabbit bool

// InitQueue inspects the RABBITMQ_URL environment variable and attempts to
// initialise a RabbitMQ connection.  If successful, useRabbit is set to true.
// Errors are logged but do not panic; this allows the application to run
// without RabbitMQ if desired.  This function should be called during
// application startup, e.g. in NewApp().
func InitQueue() {
    amqpURL := os.Getenv("RABBITMQ_URL")
    if amqpURL == "" {
        return
    }
    if err := InitRabbitMQ(amqpURL); err != nil {
        log.Printf("[Queue] Failed to initialise RabbitMQ: %v", err)
    } else {
        useRabbit = true
    }
}

// Publish sends a raw string message to the specified channel or queue.  When
// RabbitMQ is enabled, the message is published as plain text to the queue.
// Otherwise the message is sent to Redis on the given channel.  Errors are
// returned to the caller so they may be logged if desired.
func Publish(channel string, message string) error {
    if useRabbit {
        if rabbitClient == nil || rabbitClient.Channel == nil {
            log.Printf("[Queue] RabbitMQ requested but client not initialised; falling back to Redis")
        } else {
            return rabbitClient.Channel.Publish(
                "",
                channel,
                false,
                false,
                amqp.Publishing{ContentType: "text/plain", Body: []byte(message)},
            )
        }
    }
    // Fallback to Redis if available.  If RedisClient is nil, this is a no-op.
    if utils.RedisClient == nil {
        return nil
    }
    return utils.RedisClient.Publish(context.Background(), channel, message).Err()
}

// PublishEvent marshals a LoanEvent and publishes it via the appropriate
// transport.  When RabbitMQ is enabled the event is sent as JSON to the
// loan_events queue; otherwise it is marshalled and published as a string
// via Redis.
func PublishEvent(event model.LoanEvent) error {
    if useRabbit {
        return PublishLoanEvent(event)
    }
    data, err := json.Marshal(event)
    if err != nil {
        return err
    }
    return Publish("loan_events", string(data))
}