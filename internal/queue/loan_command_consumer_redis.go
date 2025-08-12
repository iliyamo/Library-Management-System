package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	redis "github.com/redis/go-redis/v9"

	"github.com/iliyamo/Library-Management-System/internal/model"
	"github.com/iliyamo/Library-Management-System/internal/repository"
)

// StartLoanCommandConsumerRedis subscribes to the loan_commands channel via Redis Pub/Sub
// and dispatches commands. It reuses processBorrow/processReturn defined in this package
// (used by the Rabbit consumer), so there are no undefined helpers.
func StartLoanCommandConsumerRedis(ctx context.Context, rdb *redis.Client, loanRepo *repository.LoanRepository, bookRepo *repository.BookRepository) error {
	if rdb == nil {
		return fmt.Errorf("nil redis client")
	}

	sub := rdb.Subscribe(ctx, LoanCommandsQueue) // "loan_commands"

	// Ensure subscription is created
	if _, err := sub.Receive(ctx); err != nil {
		return fmt.Errorf("redis subscribe %s failed: %w", LoanCommandsQueue, err)
	}

	ch := sub.Channel()

	go func() {
		for msg := range ch {
			var cmd model.LoanCommand
			if err := json.Unmarshal([]byte(msg.Payload), &cmd); err != nil {
				log.Printf("[LoanCmd][redis] bad payload: %v", err)
				continue
			}

			var procErr error
			switch cmd.Type {
			case model.CmdBorrow:
				procErr = processBorrow(cmd, loanRepo, bookRepo)
			case model.CmdReturn:
				procErr = processReturn(cmd, loanRepo, bookRepo)
			default:
				log.Printf("[LoanCmd][redis] unknown type=%q (ignored)", cmd.Type)
			}

			if procErr != nil {
				log.Printf("[LoanCmd][redis] type=%s failed: %v", cmd.Type, procErr)
			}
		}
	}()

	log.Printf("[LoanCmd][redis] subscribed to %s", LoanCommandsQueue)
	return nil
}
