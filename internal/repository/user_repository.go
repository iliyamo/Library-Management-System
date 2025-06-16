package repository

import (
	"database/sql"
	"errors"

	"github.com/iliyamo/go-learning/internal/model"
)

// UserRepository ساختاری برای انجام عملیات روی جدول users
type UserRepository struct {
	DB *sql.DB // اتصال به پایگاه داده
}

func (r *UserRepository) GetUserByID(id int) (*model.User, error) {
	query := `
		SELECT id, full_name, email, password_hash, role_id, created_at, updated_at
		FROM users
		WHERE id = ?
	`

	var user model.User
	err := r.DB.QueryRow(query, id).Scan(
		&user.ID,
		&user.FullName,
		&user.Email,
		&user.PasswordHash,
		&user.RoleID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

// NewUserRepository یک نمونه جدید از UserRepository ایجاد می‌کند
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{DB: db}
}

// CreateUser یک کاربر جدید را در جدول users وارد می‌کند
func (r *UserRepository) CreateUser(user *model.User) error {
	query := `
		INSERT INTO users (full_name, email, password_hash, role_id)
		VALUES (?, ?, ?, ?)
	`

	_, err := r.DB.Exec(query, user.FullName, user.Email, user.PasswordHash, user.RoleID)
	return err
}

// GetUserByEmail کاربر را بر اساس ایمیل از پایگاه داده پیدا می‌کند
func (r *UserRepository) GetUserByEmail(email string) (*model.User, error) {
	query := `
		SELECT id, full_name, email, password_hash, role_id, created_at, updated_at
		FROM users
		WHERE email = ?
	`

	var user model.User
	err := r.DB.QueryRow(query, email).Scan(
		&user.ID,
		&user.FullName,
		&user.Email,
		&user.PasswordHash,
		&user.RoleID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// کاربر پیدا نشد
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}
