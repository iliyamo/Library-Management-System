package model

import "time"

// User نمایانگر ساختار اطلاعات کاربران در سیستم است.
// این ساختار برای نگهداری اطلاعات پایه هر کاربر استفاده می‌شود.
type User struct {
	ID           int       `json:"id"`         // شناسه یکتا برای هر کاربر (کلید اصلی)
	FullName     string    `json:"full_name"`  // نام کامل کاربر
	Email        string    `json:"email"`      // آدرس ایمیل کاربر (باید یکتا باشد)
	PasswordHash string    `json:"-"`          // هش رمز عبور (در خروجی JSON نمایش داده نمی‌شود برای امنیت)
	RoleID       int       `json:"role_id"`    // شناسه نقش کاربر (ارجاع به جدول roles برای تعیین سطح دسترسی)
	CreatedAt    time.Time `json:"created_at"` // زمان ایجاد حساب کاربری
	UpdatedAt    time.Time `json:"updated_at"` // زمان آخرین بروزرسانی اطلاعات حساب کاربری
}

// UserSearchParams برای جستجوی کاربران با پشتیبانی از cursor-based pagination
type UserSearchParams struct {
	Query    string `query:"query"`     // متن جستجو روی full_name و email
	CursorID int    `query:"cursor_id"` // آخرین شناسه‌ای که از سمت کلاینت گرفته شده
	Limit    int    `query:"limit"`     // تعداد نتایج برگشتی در هر درخواست
}
