package repository

import (
	"database/sql"
	"time"
)

// RefreshTokenRepository مدیریت ذخیره و حذف رفرش‌توکن‌ها در پایگاه داده را بر عهده دارد.
type RefreshTokenRepository struct {
	DB *sql.DB // اتصال مستقیم به دیتابیس
}

// NewRefreshTokenRepository ساخت نمونه جدید از رفرش‌توکن ریپازیتوری
func NewRefreshTokenRepository(db *sql.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{DB: db}
}

// Store ذخیره یک refresh token برای یک کاربر خاص
func (r *RefreshTokenRepository) Store(token string, userID int) error {
	query := `
		INSERT INTO refresh_tokens (token, user_id, created_at)
		VALUES (?, ?, ?)
	`

	_, err := r.DB.Exec(query, token, userID, time.Now())
	return err
}

// DeleteAll حذف تمام رفرش‌توکن‌های مربوط به یک کاربر (مثلاً هنگام خروج از حساب)
func (r *RefreshTokenRepository) DeleteAll(userID uint) error {
	query := `
		DELETE FROM refresh_tokens WHERE user_id = ?
	`

	_, err := r.DB.Exec(query, userID)
	return err
}

// Validate بررسی می‌کند که آیا یک refresh token خاص برای کاربر معتبر است یا خیر
func (r *RefreshTokenRepository) Validate(token string, userID uint) (bool, error) {
	query := `
		SELECT COUNT(*) FROM refresh_tokens WHERE token = ? AND user_id = ?
	`

	var count int
	err := r.DB.QueryRow(query, token, userID).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
