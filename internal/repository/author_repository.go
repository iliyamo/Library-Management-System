// internal/repository/author_repository.go
package repository

import (
	"database/sql"
	"time"

	"github.com/iliyamo/go-learning/internal/model"
)

// AuthorRepository ساختار ریپازیتوری برای دسترسی به جدول نویسنده‌ها در دیتابیس است.
type AuthorRepository struct {
	DB *sql.DB // اتصال به دیتابیس
}

// NewAuthorRepository ریپازیتوری جدیدی می‌سازد.
func NewAuthorRepository(db *sql.DB) *AuthorRepository {
	return &AuthorRepository{DB: db}
}

// CreateAuthor نویسنده جدیدی را در دیتابیس ثبت می‌کند.
func (r *AuthorRepository) CreateAuthor(author *model.Author) error {
	query := `INSERT INTO authors (name, biography, birth_date, created_at) VALUES (?, ?, ?, ?)`
	res, err := r.DB.Exec(query, author.Name, author.Biography, author.BirthDate, author.CreatedAt)
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	author.ID = int(id) // 🆕 ست کردن ID دقیق برگشتی از دیتابیس
	return nil
}

// GetAllAuthors لیست همه نویسنده‌ها را بازمی‌گرداند.
func (r *AuthorRepository) GetAllAuthors() ([]model.Author, error) {
	query := `SELECT id, name, biography, birth_date, created_at FROM authors`
	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var authors []model.Author
	for rows.Next() {
		var a model.Author
		if err := rows.Scan(&a.ID, &a.Name, &a.Biography, &a.BirthDate, &a.CreatedAt); err != nil {
			return nil, err
		}
		authors = append(authors, a)
	}
	return authors, nil
}

// GetAuthorByID یک نویسنده خاص را بر اساس ID بازمی‌گرداند.
func (r *AuthorRepository) GetAuthorByID(id int) (*model.Author, error) {
	query := `SELECT id, name, biography, birth_date, created_at FROM authors WHERE id = ?`
	var a model.Author
	err := r.DB.QueryRow(query, id).Scan(&a.ID, &a.Name, &a.Biography, &a.BirthDate, &a.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return &a, nil
}

// UpdateAuthor اطلاعات نویسنده را بروزرسانی می‌کند و موفقیت عملیات را بازمی‌گرداند.
func (r *AuthorRepository) UpdateAuthor(a *model.Author) (bool, error) {
	query := `UPDATE authors SET name = ?, biography = ?, birth_date = ? WHERE id = ?`
	res, err := r.DB.Exec(query, a.Name, a.Biography, a.BirthDate, a.ID)
	if err != nil {
		return false, err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}

// DeleteAuthor نویسنده‌ای را حذف می‌کند و موفقیت عملیات را بازمی‌گرداند.
func (r *AuthorRepository) DeleteAuthor(id int) (bool, error) {
	query := `DELETE FROM authors WHERE id = ?`
	res, err := r.DB.Exec(query, id)
	if err != nil {
		return false, err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}

// Exists بررسی می‌کند آیا نویسنده‌ای با نام و تاریخ تولد مشخص وجود دارد یا نه.
func (r *AuthorRepository) Exists(name string, birthDate time.Time) (bool, error) {
	query := `SELECT COUNT(*) FROM authors WHERE name = ? AND birth_date = ?`
	var count int
	err := r.DB.QueryRow(query, name, birthDate).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
