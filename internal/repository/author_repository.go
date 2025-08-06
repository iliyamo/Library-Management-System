// internal/repository/author_repository.go
package repository

import (
	"database/sql"
	"log"
	"time"

    "github.com/iliyamo/Library-Management-System/internal/model"
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
	author.ID = int(id)
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

// SearchAuthors جستجوی full-text با پشتیبانی از cursor-based pagination
func (r *AuthorRepository) SearchAuthors(params *model.AuthorSearchParams) ([]model.Author, int, error) {
	log.Printf("[REPO] SearchAuthors: query='%s', cursor=%d, limit=%d", params.Query, params.CursorID, params.Limit)

	query := `
		SELECT id, name, biography, birth_date, created_at
		FROM authors
		WHERE MATCH(name, biography) AGAINST (? IN BOOLEAN MODE) AND id > ?
		ORDER BY id ASC
		LIMIT ?`

	rows, err := r.DB.Query(query, params.Query, params.CursorID, params.Limit)
	if err != nil {
		log.Printf("[REPO] SearchAuthors query error: %v", err)
		return nil, 0, err
	}
	defer rows.Close()

	var authors []model.Author
	for rows.Next() {
		var a model.Author
		if err := rows.Scan(&a.ID, &a.Name, &a.Biography, &a.BirthDate, &a.CreatedAt); err != nil {
			return nil, 0, err
		}
		authors = append(authors, a)
	}

	countQuery := `SELECT COUNT(*) FROM authors WHERE MATCH(name, biography) AGAINST (? IN BOOLEAN MODE)`
	var total int
	err = r.DB.QueryRow(countQuery, params.Query).Scan(&total)
	if err != nil {
		log.Printf("[REPO] Count query error: %v", err)
		return nil, 0, err
	}

	return authors, total, nil
}
