// internal/model/loan.go
package model

import "time"

// Loan نمایان‌گر یک رکورد امانت کتاب است
// Loan represents a borrow record in the loans table.  The field
// names and types mirror the columns defined in the SQL schema
// (Docs/digital_library.sql).  See the loans table definition for
// reference: id, user_id, book_id, loan_date, due_date, return_date,
// status.
type Loan struct {
    ID        uint       `json:"id" gorm:"primaryKey"`
    UserID    uint       `json:"user_id"`   // foreign key to users.id
    BookID    uint       `json:"book_id"`   // foreign key to books.id
    LoanDate  time.Time  `json:"loan_date"` // when the book was borrowed
    DueDate   time.Time  `json:"due_date"`  // when it must be returned
    ReturnDate *time.Time `json:"return_date,omitempty"` // actual return time
    Status    string     `json:"status"`    // borrowed, returned or late
}

// LoanRequest captures the minimal information required from a
// client to initiate a borrow.  Only the book identifier is
// necessary; user_id is inferred from the authenticated context and
// other fields are populated by the server.
type LoanRequest struct {
    BookID uint `json:"book_id"`
}

// Valid values for the Status field.  These align with the
// underlying ENUM in the loans table: borrowed, returned and late.
const (
    StatusBorrowed = "borrowed" // the book is currently borrowed
    StatusReturned = "returned" // the book has been returned
    StatusLate     = "late"     // the book has not been returned by the due date
)
