package service

import (
    "context"
    "errors"
    "time"

    "github.com/iliyamo/go-learning/internal/model"
    "github.com/iliyamo/go-learning/internal/queue"
    "gorm.io/gorm"
)

// LoanService ساختار اصلی برای مدیریت وام‌ها
// شامل اتصال به DB و logic لازم است.
type LoanService struct {
	DB *gorm.DB
}

// NewLoanService ایجاد سرویس جدید وام
func NewLoanService(db *gorm.DB) *LoanService {
	return &LoanService{DB: db}
}

// RequestLoan ➜ کاربر یک درخواست جدید ثبت می‌کند
func (s *LoanService) RequestLoan(ctx context.Context, userID, bookID uint) error {
    // بررسی موجود بودن کتاب
    var book model.Book
    if err := s.DB.First(&book, bookID).Error; err != nil {
        return errors.New("کتاب مورد نظر یافت نشد")
    }
    if book.AvailableCopies < 1 {
        return errors.New("هیچ نسخه‌ای از کتاب موجود نیست")
    }
    // TODO: بررسی وجود امانت فعال برای کاربر/کتاب در نسخه‌های آتی

    // ساخت و ذخیره امانت
    now := time.Now()
    loan := model.Loan{
        UserID:    userID,
        BookID:    bookID,
        LoanDate:  now,
        DueDate:   now.Add(7 * 24 * time.Hour),
        Status:    model.StatusBorrowed,
        ReturnDate: nil,
    }
    if err := s.DB.Create(&loan).Error; err != nil {
        return err
    }
    // کاهش موجودی کتاب
    book.AvailableCopies--
    _ = s.DB.Save(&book)
    // ارسال پیام به message broker برای ادامه async
    // ساخت و ارسال رویداد به صف پیام.  در صورت تنظیم RabbitMQ رویداد
    // به صورت JSON به صف ارسال می‌شود، در غیر این صورت از Redis استفاده می‌شود.
    event := model.LoanEvent{
        EventType: model.LoanRequested,
        LoanID:    loan.ID,
        UserID:    userID,
        BookID:    bookID,
        Time:      time.Now(),
    }
    _ = queue.PublishEvent(event)
    return nil
}
