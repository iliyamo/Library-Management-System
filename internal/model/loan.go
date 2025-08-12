// internal/model/loan.go
package model

import "time"

// Loan نمایان‌گر یک رکورد امانت کتاب است.
// تگ‌های gorm:type:date برای سازگاری با ستون‌های DATE در MySQL.
type Loan struct {
	ID         uint       `json:"id" gorm:"primaryKey"`
	UserID     uint       `json:"user_id"`                                // FK به users.id
	BookID     uint       `json:"book_id"`                                // FK به books.id
	LoanDate   time.Time  `json:"loan_date"  gorm:"type:date"`            // DATE
	DueDate    time.Time  `json:"due_date"   gorm:"type:date"`            // DATE
	ReturnDate *time.Time `json:"return_date,omitempty" gorm:"type:date"` // DATE/NULL
	Status     string     `json:"status"`                                 // borrowed | returned | late
}

// LoanRequest ورودی کلاینت برای شروع امانت.
type LoanRequest struct {
	BookID uint `json:"book_id"`
	// Days اگر صفر یا منفی باشد، در هندلر/سرویس شما ۷ روز پیش‌فرض اعمال می‌شود.
	Days int `json:"days,omitempty"`
}

// مقادیر معتبر وضعیت وام (هم‌راستا با ENUM دیتابیس).
const (
	StatusBorrowed = "borrowed"
	StatusReturned = "returned"
	StatusLate     = "late"
)
