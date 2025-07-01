// internal/repository/author_repository.go
package repository

import (
	"database/sql" // ارتباط با دیتابیس MySQL/MariaDB
	"errors"       // بررسی خطاها مثل sql.ErrNoRows
	"time"         // برای مقداردهی created_at

	"github.com/iliyamo/go-learning/internal/model" // ساختار دادهٔ Author
)

// AuthorRepository لایهٔ دسترسی به داده برای جدول authors
// متدهای CRUD را فراهم می‌کند.
type AuthorRepository struct {
	DB *sql.DB // هندلر اتصال به پایگاه داده
}

// NewAuthorRepository سازندهٔ ریپازیتوری
func NewAuthorRepository(db *sql.DB) *AuthorRepository {
	return &AuthorRepository{DB: db}
}

// CreateAuthor نویسندهٔ جدیدی را در جدول authors درج می‌کند.
func (r *AuthorRepository) CreateAuthor(author *model.Author) error {
	const query = `INSERT INTO authors (name, biography, birth_date, created_at) VALUES (?, ?, ?, ?)`
	_, err := r.DB.Exec(query, author.Name, author.Biography, author.BirthDate, time.Now())
	return err // در صورت خطا به لایه بالاتر ارسال می‌شود
}

// GetAllAuthors تمام نویسنده‌ها را برمی‌گرداند.
func (r *AuthorRepository) GetAllAuthors() ([]model.Author, error) {
	const query = `SELECT id, name, biography, birth_date, created_at FROM authors`

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

// GetAuthorByID نویسنده‌ای با شناسهٔ مشخص برمی‌گرداند.
// اگر نویسنده‌ای پیدا نشود، (nil, nil) برگشت داده می‌شود.
func (r *AuthorRepository) GetAuthorByID(id int) (*model.Author, error) {
	const query = `SELECT id, name, biography, birth_date, created_at FROM authors WHERE id = ?`

	var a model.Author
	err := r.DB.QueryRow(query, id).Scan(&a.ID, &a.Name, &a.Biography, &a.BirthDate, &a.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // نویسنده‌ای با این ID نیست
		}
		return nil, err
	}
	return &a, nil
}

// UpdateAuthor اطلاعات نویسنده را بروزرسانی می‌کند.
// اگر نویسنده وجود نداشته باشد، sql.ErrNoRows بازگردانده می‌شود تا هندلر پیام مناسب بدهد.
func (r *AuthorRepository) UpdateAuthor(a *model.Author) error {
	const query = `UPDATE authors SET name = ?, biography = ?, birth_date = ? WHERE id = ?`

	res, err := r.DB.Exec(query, a.Name, a.Biography, a.BirthDate, a.ID)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows // هیچ رکوردی آپدیت نشد ➜ نویسنده‌ای با این ID وجود ندارد
	}
	return nil
}

// DeleteAuthor نویسنده‌ای را حذف می‌کند.
// اگر ID موجود نباشد، sql.ErrNoRows بازمی‌گرداند.
func (r *AuthorRepository) DeleteAuthor(id int) error {
	const query = `DELETE FROM authors WHERE id = ?`

	res, err := r.DB.Exec(query, id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows // نویسنده‌ای با این ID پیدا نشد
	}
	return nil
}
