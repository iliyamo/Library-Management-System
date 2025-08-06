// internal/queue/rabbit.go
//
// This file defines a minimal RabbitMQ client wrapper for the go-learning
// project.  It encapsulates the connection and channel to RabbitMQ and
// exposes helper functions to initialise and close these resources as well
// as publish loan events.  The implementation is kept simple to avoid
// unnecessary complexity for developers who are just getting started with Go.

package queue

import (
    "encoding/json"
    "fmt"
    "log"
    "os"
    "io"

    "github.com/streadway/amqp"

    "github.com/iliyamo/Library-Management-System/internal/model"
)

// RabbitMQClient holds the AMQP connection and channel used for publishing messages.
type RabbitMQClient struct {
    Conn    *amqp.Connection
    Channel *amqp.Channel
}

// rabbitClient is a package-level singleton.  It will be initialised on demand
// by InitRabbitMQ and closed via CloseRabbitMQ.
var rabbitClient *RabbitMQClient

// rabbitLogger writes RabbitMQ-specific logs to both standard output and a file.  It is
// initialised in InitRabbitMQ when a connection is successfully established.  All
// publishing and consuming operations should write to this logger rather than the
// global logger so that RabbitMQ activities are captured separately.  If
// rabbitLogger is nil, logging falls back to the standard log package.
var (
    rabbitLogger  *log.Logger
    rabbitLogFile *os.File
)

// InitRabbitMQ establishes a connection to RabbitMQ and declares a durable
// queue named "loan_events".  If a connection is already present, this
// function does nothing.  Any errors encountered are returned to the caller.
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
        conn.Close()
        return fmt.Errorf("failed to open channel: %w", err)
    }
    // Declare the queue to ensure it exists.  Durable queues survive broker restarts.
    if _, err := ch.QueueDeclare(
        "loan_events",
        true,  // durable
        false, // autoDelete
        false, // exclusive
        false, // noWait
        nil,   // args
    ); err != nil {
        ch.Close()
        conn.Close()
        return fmt.Errorf("failed to declare queue: %w", err)
    }
    rabbitClient = &RabbitMQClient{Conn: conn, Channel: ch}
    // Initialise a dedicated logger for RabbitMQ.  Logs are written to
    // rabbitmq.log in append mode as well as stdout.  Any failure to
    // open the log file does not prevent the connection from being
    // established but will fall back to the default logger.
    if rabbitLogger == nil {
        // Create or open the log file once per application run
        f, errf := os.OpenFile("rabbitmq.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
        if errf == nil {
            rabbitLogFile = f
            multi := io.MultiWriter(os.Stdout, f)
            rabbitLogger = log.New(multi, "[RabbitMQ] ", log.LstdFlags)
        } else {
            // Unable to open log file; fall back to stdout only
            rabbitLogger = log.New(os.Stdout, "[RabbitMQ] ", log.LstdFlags)
        }
    }
    if rabbitLogger != nil {
        rabbitLogger.Println("✅ RabbitMQ connected and queue declared")
    } else {
        log.Println("✅ RabbitMQ connected and queue declared")
    }
    return nil
}

// PublishLoanEvent publishes a loan event to the loan_events queue.  The event
// is marshalled to JSON before publishing.  If RabbitMQ is not initialised
// the function returns an error so callers can decide whether to fall back.
func PublishLoanEvent(event model.LoanEvent) error {
    if rabbitClient == nil || rabbitClient.Channel == nil {
        return fmt.Errorf("RabbitMQ is not initialised")
    }
    body, err := json.Marshal(event)
    if err != nil {
        return fmt.Errorf("failed to marshal event: %w", err)
    }
    // Log the outgoing event
    if rabbitLogger != nil {
        rabbitLogger.Printf("publishing event: %+v", event)
    }
    err = rabbitClient.Channel.Publish(
        "",
        "loan_events",
        false,
        false,
        amqp.Publishing{
            ContentType: "application/json",
            Body:        body,
        },
    )
    return err
}

// CloseRabbitMQ closes the channel and connection if they are open.  It is
// safe to call this function multiple times; subsequent calls have no effect.
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
    // Close the RabbitMQ log file if one was opened.  We track the file in
    // rabbitLogFile so it can be closed explicitly on shutdown.  After
    // closing the file we nil out both the logger and the file handle.
    if rabbitLogFile != nil {
        _ = rabbitLogFile.Close()
        rabbitLogFile = nil
    }
    rabbitLogger = nil
}