package model

import "time"

// RefreshToken نماینده یک رفرش توکن برای کاربر است.
// این توکن‌ها در دیتابیس نگهداری می‌شوند تا بتوان احراز هویت بلندمدت را مدیریت کرد.
type RefreshToken struct {
	ID        int       `json:"id"`         // شناسه یکتا در جدول
	UserID    int       `json:"user_id"`    // کلید خارجی به جدول users
	Token     string    `json:"token"`      // رشته‌ی توکن (JWT امضاشده)
	ExpiresAt time.Time `json:"expires_at"` // زمان انقضای توکن
	CreatedAt time.Time `json:"created_at"` // زمان ایجاد توکن
}
