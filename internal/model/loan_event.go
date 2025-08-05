package model

import "time"

// LoanEvent تعریف یک رویداد مرتبط با سیستم امانت‌ها است.
// این ساختار برای ارسال پیام به message broker یا پردازش‌های asynchronous کاربرد دارد.
// فیلد EventType نوع رویداد را مشخص می‌کند و مطابق با ثابت‌های تعریف‌شده در ادامه است.
type LoanEvent struct {
    EventType string    `json:"event_type"` // نوع رویداد (LoanRequested، LoanApproved و ...)
    LoanID    uint      `json:"loan_id"`    // شناسه امانت
    UserID    uint      `json:"user_id"`    // شناسه کاربر
    BookID    uint      `json:"book_id"`    // شناسه کتاب
    Time      time.Time `json:"time"`       // زمان وقوع رویداد
}

// ثابت‌های رویدادها برای استفاده در سیستم صف
const (
    LoanRequested = "LoanRequested" // کاربر درخواست امانت داده است
    LoanApproved  = "LoanApproved"  // درخواست امانت پذیرفته شد (فعلاً استفاده نشده)
    LoanRejected  = "LoanRejected"  // درخواست امانت رد شد (فعلاً استفاده نشده)
    LoanReturned  = "LoanReturned"  // کتاب بازگردانده شده است
)