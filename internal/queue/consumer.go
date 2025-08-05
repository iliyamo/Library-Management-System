// internal/queue/consumer.go
package queue

import (
    "context"
    "encoding/json"
    "log"

    "github.com/redis/go-redis/v9"

    "github.com/iliyamo/go-learning/internal/model"
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
    switch evt.EventType {
    case model.LoanRequested:
        log.Printf("[Loan] User %d requested loan for book %d", evt.UserID, evt.BookID)
    case model.LoanApproved:
        log.Printf("[Loan] Loan %d approved", evt.LoanID)
    case model.LoanRejected:
        log.Printf("[Loan] Loan %d rejected", evt.LoanID)
    case model.LoanReturned:
        log.Printf("[Loan] Loan %d returned", evt.LoanID)
    default:
        log.Printf("[Loan] Unknown event: %+v", evt)
    }
}
