package repository

import (
	"database/sql"
	"errors"

	models "github.com/iliyamo/go-learning/internal/model"
)

// UserRepository ساختار مدیریت ارتباط با دیتابیس برای کاربران
type UserRepository struct {
	DB *sql.DB // اتصال به دیتابیس SQL
}

// NewUserRepository سازنده برای ایجاد یک نمونه UserRepository با اتصال دیتابیس
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{DB: db}
}

// CreateUser یک کاربر جدید را در جدول users درج می‌کند
func (r *UserRepository) CreateUser(user *models.User) error {
	query := `
		INSERT INTO users (email, password_hash, role_id)
		VALUES (?, ?, ?)
	`
	// اجرای کوئری درج داده‌ها
	_, err := r.DB.Exec(query, user.Email, user.PasswordHash, user.RoleID)
	return err // بازگرداندن خطا در صورت وجود
}

// GetUserByEmail کاربری را بر اساس ایمیل جستجو و برمی‌گرداند
func (r *UserRepository) GetUserByEmail(email string) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, role_id FROM users
		WHERE email = ?
	`

	var user models.User

	// اجرای کوئری و اسکن نتیجه در ساختار user
	err := r.DB.QueryRow(query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.RoleID,
	)

	// مدیریت خطاها
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// اگر کاربر پیدا نشد، مقدار nil و خطای nil برگردانده شود
			return nil, nil
		}
		// سایر خطاهای احتمالی برگردانده شوند
		return nil, err
	}

	return &user, nil
}
