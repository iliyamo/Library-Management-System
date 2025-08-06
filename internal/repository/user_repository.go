package repository

import (
	"database/sql"
	"errors"

    "github.com/iliyamo/Library-Management-System/internal/model"
)

// UserRepository ساختاری است برای مدیریت عملیات روی جدول users
type UserRepository struct {
	DB *sql.DB
}

// NewUserRepository سازنده ریپازیتوری
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{DB: db}
}

// GetUserByID واکشی کاربر با شناسه یکتا
func (r *UserRepository) GetUserByID(id int) (*model.User, error) {
	query := `
		SELECT id, full_name, email, password_hash, role_id, created_at, updated_at
		FROM users
		WHERE id = ?`

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

// CreateUser درج کاربر جدید
func (r *UserRepository) CreateUser(user *model.User) error {
	query := `INSERT INTO users (full_name, email, password_hash, role_id) VALUES (?, ?, ?, ?)`
	_, err := r.DB.Exec(query, user.FullName, user.Email, user.PasswordHash, user.RoleID)
	return err
}

// GetUserByEmail واکشی کاربر با ایمیل
func (r *UserRepository) GetUserByEmail(email string) (*model.User, error) {
	query := `
		SELECT id, full_name, email, password_hash, role_id, created_at, updated_at
		FROM users
		WHERE email = ?`

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
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// SearchUsers ➜ جستجوی full-text بین نام و ایمیل با cursor-based pagination
func (r *UserRepository) SearchUsers(query string, cursorID, limit int) ([]model.User, int, error) {
	q := `
		SELECT id, full_name, email, password_hash, role_id, created_at, updated_at
		FROM users
		WHERE MATCH(full_name, email) AGAINST (? IN BOOLEAN MODE)
		  AND id > ?
		ORDER BY id ASC
		LIMIT ?`

	rows, err := r.DB.Query(q, query, cursorID, limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(
			&u.ID, &u.FullName, &u.Email, &u.PasswordHash,
			&u.RoleID, &u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}

	countQuery := `SELECT COUNT(*) FROM users WHERE MATCH(full_name, email) AGAINST (? IN BOOLEAN MODE)`
	var total int
	err = r.DB.QueryRow(countQuery, query).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}
