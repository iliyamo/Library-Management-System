package repository

import (
	"database/sql" // پکیج استاندارد برای عملیات SQL
	"errors"       // برای مقایسه خطاها

	"github.com/iliyamo/go-learning/internal/model" // مدل‌های داده‌ای پروژه
)

// UserRepository ساختاری است برای مدیریت عملیات CRUD بر روی جدول users.
// این ریپازیتوری اتصال به دیتابیس را در فیلد DB نگه می‌دارد.
type UserRepository struct {
	DB *sql.DB // نشان‌دهندهٔ اتصال فعال به دیتابیس
}

// GetUserByID یک کاربر را بر اساس شناسه‌ی یکتا بازیابی می‌کند.
// اگر کاربر پیدا نشود، مقدار nil برگردانده و خطا ندارد.
func (r *UserRepository) GetUserByID(id int) (*model.User, error) {
	// تعریف کوئری SQL برای انتخاب ستون‌های مورد نیاز
	query := `
		SELECT id, full_name, email, password_hash, role_id, created_at, updated_at
		FROM users
		WHERE id = ?
	`

	// متغیر برای نگهداری نتیجه
	var user model.User

	// اجرای کوئری و نگاشت نتایج به فیلدهای مدل
	err := r.DB.QueryRow(query, id).Scan(
		&user.ID,
		&user.FullName,
		&user.Email,
		&user.PasswordHash,
		&user.RoleID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	// اگر خطایی رخ داد، بررسی می‌کنیم آیا خطا از نوع "بدون سطر" است
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// یعنی کاربری با این id وجود ندارد
			return nil, nil
		}
		// هر خطای دیگر را بازگردان
		return nil, err
	}

	// در صورتی که موفق بودیم، آدرس user را برمی‌گردانیم
	return &user, nil
}

// NewUserRepository یک نمونهٔ جدید از UserRepository می‌سازد
// و اتصال DB را در آن می‌نویسد.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{DB: db}
}

// CreateUser یک کاربر جدید را در جدول users درج می‌کند.
// فیلدهای full_name، email، password_hash و role_id را مقداردهی می‌کند.
func (r *UserRepository) CreateUser(user *model.User) error {
	// کوئری درج داده به همراه پارامترهای مورد نیاز
	query := `
		INSERT INTO users (full_name, email, password_hash, role_id)
		VALUES (?, ?, ?, ?)
	`

	// اجرای کوئری و ارسال مقادیر
	_, err := r.DB.Exec(
		query,
		user.FullName,
		user.Email,
		user.PasswordHash,
		user.RoleID,
	)
	// خطای ممکن را به بالا پاس می‌دهیم
	return err
}

// GetUserByEmail کاربر را براساس ایمیل جستجو می‌کند.
// اگر کاربری با ایمیل داده‌شده وجود نداشت، nil و بدون خطا برمی‌گردد.
func (r *UserRepository) GetUserByEmail(email string) (*model.User, error) {
	// کوئری انتخاب بر اساس ایمیل
	query := `
		SELECT id, full_name, email, password_hash, role_id, created_at, updated_at
		FROM users
		WHERE email = ?
	`

	var user model.User

	// اجرای کوئری و نگاشت نتایج
	err := r.DB.QueryRow(query, email).Scan(
		&user.ID,
		&user.FullName,
		&user.Email,
		&user.PasswordHash,
		&user.RoleID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	// کنترل خطاها
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// کاربر پیدا نشد
			return nil, nil
		}
		// خطای دیگر را بازگردان
		return nil, err
	}

	// در صورت موفقیت، آدرس user را بازمی‌گردانیم
	return &user, nil
}
