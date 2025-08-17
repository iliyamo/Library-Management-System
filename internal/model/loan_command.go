// internal/model/loan_command.go
package model

import "time"

const (
	CmdBorrow = "borrow"
	CmdReturn = "return"
)

type LoanCommand struct {
	Type          string    `json:"type"` // ⟵ اینجا Type هست
	CorrelationID string    `json:"correlation_id,omitempty"`
	RequestedAt   time.Time `json:"requested_at,omitempty"`
	Payload       struct {
		UserID uint `json:"user_id"`
		BookID uint `json:"book_id,omitempty"`
		Days   int  `json:"days,omitempty"`
		LoanID uint `json:"loan_id,omitempty"`
	} `json:"payload"`
}
