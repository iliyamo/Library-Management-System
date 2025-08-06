// internal/queue/consumer.go
package queue

import (
    "context"
    "encoding/json"
    "log"
    "time"

    "github.com/redis/go-redis/v9"

    "github.com/iliyamo/Library-Management-System/internal/model"
)

var (
	redisCtx      = context.Background()
	loanEventChan = "loan_events"
)

// StartLoanConsumer برای نشستن روی چنل Redis و پرست هر event است
// StartLoanConsumer روی کانال مشخص‌شده در Redis مشترک می‌شود و رویدادها را به handlerFunc هدایت می‌کند.
func StartLoanConsumer(client *redis.Client, handlerFunc func(event model.LoanEvent)) {
	go func() {
		sub := client.Subscribe(redisCtx, loanEventChan)
		ch := sub.Channel()

		log.Println("[Queue] Listening on channel:", loanEventChan)

        for msg := range ch {
            var evt model.LoanEvent
            if err := json.Unmarshal([]byte(msg.Payload), &evt); err != nil {
                log.Printf("[Queue] Invalid event received: %v\n", err)
                continue
            }
            log.Printf("[Queue] Received event: %+v\n", evt)
            handlerFunc(evt)
        }
	}()
}

// ExampleHandler یک هندلر ساده برای loan events
// ExampleHandler نمونه‌ای از پردازشگر رویدادهای امانت است که صرفاً پیام‌های لاگ ایجاد می‌کند.
func ExampleHandler(evt model.LoanEvent) {
    // Choose the appropriate logger.  If RabbitMQ logging is enabled use it,
    // otherwise fall back to the standard logger.
    logger := log.Printf
    if rabbitLogger != nil {
        // wrap rabbitLogger.Println into a printf-like function
        logger = func(format string, args ...interface{}) {
            rabbitLogger.Printf(format, args...)
        }
    }
    switch evt.EventType {
    case model.LoanRequested:
        // Calculate days until the due date.  If DueDate is zero the result will be 0.
        days := int(evt.DueDate.Sub(time.Now()).Hours() / 24)
        if evt.RemainingCopies <= 0 {
            logger("[Loan] User %d borrowed book %d; no copies left. Due in %d days.", evt.UserID, evt.BookID, days)
        } else {
            logger("[Loan] User %d borrowed book %d; remaining copies: %d. Due in %d days.", evt.UserID, evt.BookID, evt.RemainingCopies, days)
        }
    
        // Schedule a log six hours before due date and at due date
        if !evt.DueDate.IsZero() {
            sixHoursBefore := evt.DueDate.Add(-6 * time.Hour)
            go func(e model.LoanEvent) {
                d := time.Until(sixHoursBefore)
                if d > 0 {
                    time.Sleep(d)
                }
                logger("[Loan] Book %d for user %d is due in 6 hours.", e.BookID, e.UserID)
            }(evt)
            go func(e model.LoanEvent) {
                d := time.Until(e.DueDate)
                if d > 0 {
                    time.Sleep(d)
                }
                logger("[Loan] Book %d for user %d is overdue.", e.BookID, e.UserID)
            }(evt)
        }
case model.LoanReturned:
        if evt.RemainingCopies <= 0 {
            logger("[Loan] User %d returned book %d. No copies left.", evt.UserID, evt.BookID)
        } else {
            logger("[Loan] User %d returned book %d. Remaining copies: %d.", evt.UserID, evt.BookID, evt.RemainingCopies)
        }
    case model.LoanApproved:
        logger("[Loan] Loan %d approved.", evt.LoanID)
    case model.LoanRejected:
        logger("[Loan] Loan %d rejected.", evt.LoanID)
    default:
        logger("[Loan] Unknown event: %+v", evt)
    }
}
