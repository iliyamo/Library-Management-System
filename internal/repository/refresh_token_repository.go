package repository

import (
	"database/sql" // برای کار با پایگاه داده SQL
	"time"         // برای ثبت timestamp زمانی
)

// RefreshTokenRepository مدیریت توکن‌های تجدید دسترسی در دیتابیس را بر عهده دارد.
// این ریپازیتوری قابلیت ذخیره، حذف و اعتبارسنجی توکن‌ها را ارائه می‌کند.
type RefreshTokenRepository struct {
	DB *sql.DB // نشان‌دهندهٔ اتصال به دیتابیس
}

// NewRefreshTokenRepository یک نمونهٔ جدید از ریپازیتوری توکن‌ها می‌سازد.
func NewRefreshTokenRepository(db *sql.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{DB: db}
}

// Store یک توکن تجدید را برای کاربر مشخص ذخیره می‌کند.
// پارامترها:
// - token: رشتهٔ توکن JWT برای تجدید
// - userID: شناسهٔ کاربر صاحب توکن
func (r *RefreshTokenRepository) Store(token string, userID int) error {
	// کوئری SQL برای درج رکورد جدید در جدول refresh_tokens
	query := `
		INSERT INTO refresh_tokens (token, user_id, created_at)
		VALUES (?, ?, ?)
	`

	// اجرای کوئری با پارامترهای token، userID و زمان فعلی
	_, err := r.DB.Exec(query, token, userID, time.Now())
	// ارور اجرایی را مستقیماً بازمی‌گرداند
	return err
}

// DeleteAll تمام توکن‌های تجدید مربوط به یک userID را حذف می‌کند.
// این متد معمولاً هنگام خروج کاربر از سیستم فراخوانی می‌شود.
// پارامتر:
// - userID: شناسهٔ کاربر که توکن‌هایش باید حذف شود
func (r *RefreshTokenRepository) DeleteAll(userID uint) error {
	// کوئری SQL برای حذف تمام رکوردهای مربوط به userID
	query := `
		DELETE FROM refresh_tokens WHERE user_id = ?
	`

	// اجرای کوئری حذف
	_, err := r.DB.Exec(query, userID)
	return err
}

// Validate بررسی می‌کند که آیا یک توکن تجدید مشخص برای یک کاربر معتبر است یا خیر.
// این متد تعداد رکوردهای مطابق را شمرده و در صورت بزرگ‌تر از صفر معتبر می‌داند.
// پارامترها:
// - token: رشتهٔ توکن JWT که باید بررسی شود
// - userID: شناسهٔ کاربر صاحب توکن
// بازگشت:
// - bool: معتبر (true) یا نامعتبر (false)
// - error: خطای احتمالی در زمان اجرا
func (r *RefreshTokenRepository) Validate(token string, userID uint) (bool, error) {
	// کوئری SQL برای شمارش تعداد رکوردهای مطابق
	query := `
		SELECT COUNT(*) FROM refresh_tokens WHERE token = ? AND user_id = ?
	`

	var count int
	// اجرای کوئری و اسکن نتیجه در count
	err := r.DB.QueryRow(query, token, userID).Scan(&count)
	if err != nil {
		// در صورت خطا، false و خود خطا را بازمی‌گرداند
		return false, err
	}

	// اگر count بزرگ‌تر از صفر باشد، توکن معتبر است
	return count > 0, nil
}
