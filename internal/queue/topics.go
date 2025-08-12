// internal/queue/topics.go
package queue

// Centralized queue / topic names to avoid duplicate declarations across files.
const (
	// Used for publishing/consuming domain events (e.g., LoanRequested, LoanReturned)
	LoanEventsQueue = "loan_events"

	// Used for publishing/consuming commands (e.g., borrow/return commands)
	LoanCommandsQueue = "loan_commands"
)
