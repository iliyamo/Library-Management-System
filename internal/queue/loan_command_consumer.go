// internal/queue/loan_command_consumer.go

package queue

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/iliyamo/Library-Management-System/internal/model"
	"github.com/iliyamo/Library-Management-System/internal/repository"
	"github.com/streadway/amqp"
)

// StartLoanCommandConsumerRabbit consumes messages from loan_commands and dispatches them.
func StartLoanCommandConsumerRabbit(ch *amqp.Channel, loanRepo *repository.LoanRepository, bookRepo *repository.BookRepository) error {
	qName := LoanCommandsQueue // "loan_commands"

	// Subscribing to the loan_commands queue
	deliveries, err := ch.Consume(qName, "", false /* autoAck */, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume %s failed: %w", qName, err)
	}

	// Processing messages concurrently
	go func() {
		for d := range deliveries {
			var cmd model.LoanCommand
			if err := json.Unmarshal(d.Body, &cmd); err != nil {
				log.Printf("[LoanCmd] bad payload: %v", err)
				_ = d.Nack(false, false) // Drop poison (invalid message)
				continue
			}

			var procErr error
			switch cmd.Type {
			case model.CmdBorrow:
				procErr = processBorrow(cmd, loanRepo, bookRepo)
			case model.CmdReturn:
				procErr = processReturn(cmd, loanRepo, bookRepo)
			default:
				log.Printf("[LoanCmd] unknown type=%q -> ack", cmd.Type)
				_ = d.Ack(false)
				continue
			}

			// Handling processing errors
			if procErr != nil {
				log.Printf("[LoanCmd] type=%s failed: %v (will requeue)", cmd.Type, procErr)
				_ = d.Nack(false, true) // Requeue if it's a transient error
				continue
			}
			_ = d.Ack(false) // Acknowledge message after successful processing
		}
	}()

	log.Printf("[LoanCmd] subscribed to %s", qName)
	return nil
}

// processBorrow processes the borrow command and updates the loan and book status.
func processBorrow(cmd model.LoanCommand, loanRepo *repository.LoanRepository, bookRepo *repository.BookRepository) error {
	// 1) Check if the book is available
	book, err := bookRepo.GetBookByID(int(cmd.Payload.BookID))
	if err != nil {
		return fmt.Errorf("get book: %w", err)
	}
	if book == nil {
		return fmt.Errorf("book %d not found", cmd.Payload.BookID)
	}
	if book.AvailableCopies < 1 {
		return fmt.Errorf("no copies available for book %d", book.ID)
	}

	// 2) Check if the user already has an active loan for this book
	hasActive, err := loanRepo.CheckActiveLoan(int(cmd.Payload.UserID), int(cmd.Payload.BookID))
	if err != nil {
		return fmt.Errorf("check active loan: %w", err)
	}
	if hasActive {
		return fmt.Errorf("user %d already borrowed book %d", cmd.Payload.UserID, cmd.Payload.BookID)
	}

	// 3) Create loan record and update book inventory
	days := cmd.Payload.Days
	if days <= 0 {
		days = 7 // Default to 7 days if not specified
	}
	now := time.Now()
	due := now.Add(time.Duration(days) * 24 * time.Hour)

	loan := &model.Loan{
		UserID:   cmd.Payload.UserID,
		BookID:   cmd.Payload.BookID,
		LoanDate: now,
		DueDate:  due,
		Status:   model.StatusBorrowed,
	}
	if err := loanRepo.CreateLoan(loan); err != nil {
		return fmt.Errorf("create loan: %w", err)
	}

	// Reduce available copies of the book
	book.AvailableCopies--
	if _, err := bookRepo.UpdateBook(book); err != nil {
		return fmt.Errorf("update book: %w", err)
	}

	// Publish loan event
	_ = PublishEvent(model.LoanEvent{
		EventType:       model.LoanRequested,
		LoanID:          loan.ID,
		UserID:          loan.UserID,
		BookID:          loan.BookID,
		Time:            time.Now(),
		RemainingCopies: int(book.AvailableCopies),
		DueDate:         due,
	})
	return nil
}

// processReturn processes the return command and updates the loan and book status.
func processReturn(cmd model.LoanCommand, loanRepo *repository.LoanRepository, bookRepo *repository.BookRepository) error {
	// Ensure LoanID is provided in the payload
	if cmd.Payload.LoanID == 0 {
		return fmt.Errorf("return requires loan_id in payload")
	}
	loan, err := loanRepo.GetLoanByID(int(cmd.Payload.LoanID))
	if err != nil {
		return fmt.Errorf("get loan: %w", err)
	}
	if loan == nil {
		return fmt.Errorf("loan %d not found", cmd.Payload.LoanID)
	}

	// Check if the provided user ID matches the loan's user ID
	if cmd.Payload.UserID != 0 && loan.UserID != cmd.Payload.UserID {
		return fmt.Errorf("loan %d belongs to user %d not %d", loan.ID, loan.UserID, cmd.Payload.UserID)
	}

	// If the book has already been returned, return success
	if loan.Status != model.StatusBorrowed {
		return nil
	}

	// Mark loan as returned
	updated, err := loanRepo.MarkAsReturned(loan.ID, loan.UserID)
	if err != nil {
		return fmt.Errorf("mark returned: %w", err)
	}
	if !updated {
		return fmt.Errorf("no rows updated for loan %d", loan.ID)
	}

	// Increase available copies of the book
	book, err := bookRepo.GetBookByID(int(loan.BookID))
	if err == nil && book != nil {
		book.AvailableCopies++
		if _, err := bookRepo.UpdateBook(book); err != nil {
			return fmt.Errorf("update book: %w", err)
		}
	}

	remaining := 0
	if book != nil {
		remaining = int(book.AvailableCopies)
	}
	_ = PublishEvent(model.LoanEvent{
		EventType:       model.LoanReturned,
		LoanID:          loan.ID,
		UserID:          loan.UserID,
		BookID:          loan.BookID,
		Time:            time.Now(),
		RemainingCopies: remaining,
	})
	return nil
}
