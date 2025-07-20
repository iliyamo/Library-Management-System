// internal/repository/book_repository.go
package repository

import (
	"database/sql"
	"errors"

	"github.com/iliyamo/go-learning/internal/model"
)

type BookRepository struct {
	DB *sql.DB
}

// سازندهٔ ریپازیتوری
func NewBookRepository(db *sql.DB) *BookRepository {
	return &BookRepository{DB: db}
}

// ExistsByISBN بررسی می‌کند ISBN تکراری نباشد
func (r *BookRepository) ExistsByISBN(isbn string) (bool, error) {
	var cnt int
	if err := r.DB.QueryRow(`SELECT COUNT(*) FROM books WHERE isbn = ?`, isbn).Scan(&cnt); err != nil {
		return false, err
	}
	return cnt > 0, nil
}

// CreateBook درج کتاب تازه
func (r *BookRepository) CreateBook(b *model.Book) error {
	_, err := r.DB.Exec(`
		INSERT INTO books
		    (title, isbn, author_id, category_id, description,
		     published_year, total_copies, available_copies, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		b.Title, b.ISBN, b.AuthorID, b.CategoryID, b.Description,
		b.PublishedYear, b.TotalCopies, b.AvailableCopies, b.CreatedAt,
	)
	return err
}

// GetAllBooks برگرداندن همه کتاب‌ها
func (r *BookRepository) GetAllBooks() ([]model.Book, error) {
	rows, err := r.DB.Query(`
		SELECT id, title, isbn, author_id, category_id, description,
		       published_year, total_copies, available_copies, created_at
		FROM books`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []model.Book
	for rows.Next() {
		var b model.Book
		if err := rows.Scan(
			&b.ID, &b.Title, &b.ISBN, &b.AuthorID, &b.CategoryID,
			&b.Description, &b.PublishedYear, &b.TotalCopies,
			&b.AvailableCopies, &b.CreatedAt,
		); err != nil {
			return nil, err
		}
		books = append(books, b)
	}
	return books, nil
}

// GetBookByID واکشی کتاب با ID
func (r *BookRepository) GetBookByID(id int) (*model.Book, error) {
	var b model.Book
	err := r.DB.QueryRow(`
		SELECT id, title, isbn, author_id, category_id, description,
		       published_year, total_copies, available_copies, created_at
		FROM books WHERE id = ?`, id).
		Scan(&b.ID, &b.Title, &b.ISBN, &b.AuthorID, &b.CategoryID,
			&b.Description, &b.PublishedYear, &b.TotalCopies,
			&b.AvailableCopies, &b.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil // پیدا نشد
	}
	return &b, err
}

// UpdateBook بروزرسانی؛ مقدار بولی می‌گوید سطری تغییر کرده یا نه
func (r *BookRepository) UpdateBook(b *model.Book) (bool, error) {
	res, err := r.DB.Exec(`
		UPDATE books
		SET title = ?, isbn = ?, author_id = ?, category_id = ?, description = ?,
		    published_year = ?, total_copies = ?, available_copies = ?
		WHERE id = ?`,
		b.Title, b.ISBN, b.AuthorID, b.CategoryID, b.Description,
		b.PublishedYear, b.TotalCopies, b.AvailableCopies, b.ID)
	if err != nil {
		return false, err
	}
	aff, _ := res.RowsAffected()
	return aff > 0, nil
}

// DeleteBook حذف با ID
func (r *BookRepository) DeleteBook(id int) (bool, error) {
	res, err := r.DB.Exec(`DELETE FROM books WHERE id = ?`, id)
	if err != nil {
		return false, err
	}
	aff, _ := res.RowsAffected()
	return aff > 0, nil
}

// SearchBooks ➜ پشتیبانی از full-text و گرفتن همه کتاب‌ها در صورت نبود query
func (r *BookRepository) SearchBooks(params *model.BookSearchParams) ([]model.Book, int, error) {
	var (
		rows       *sql.Rows
		query      string
		args       []interface{}
		totalCount int
		err        error
	)

	if params.Query == "" {
		// حالت بدون جستجو: فقط بر اساس cursor id
		query = `
			SELECT id, title, isbn, author_id, category_id, description,
			       published_year, total_copies, available_copies, created_at
			FROM books
			WHERE id > ?
			ORDER BY id ASC
			LIMIT ?`
		args = []interface{}{params.CursorID, params.Limit}
	} else {
		// حالت جستجوی متنی
		query = `
			SELECT id, title, isbn, author_id, category_id, description,
			       published_year, total_copies, available_copies, created_at
			FROM books
			WHERE MATCH(title, description) AGAINST (? IN BOOLEAN MODE)
			AND id > ?
			ORDER BY id ASC
			LIMIT ?`
		args = []interface{}{params.Query, params.CursorID, params.Limit}
	}

	rows, err = r.DB.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var books []model.Book
	for rows.Next() {
		var b model.Book
		if err := rows.Scan(
			&b.ID, &b.Title, &b.ISBN, &b.AuthorID, &b.CategoryID,
			&b.Description, &b.PublishedYear, &b.TotalCopies,
			&b.AvailableCopies, &b.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		books = append(books, b)
	}

	if params.Query == "" {
		// کل شمارش فقط وقتی query خالیه
		err = r.DB.QueryRow(`SELECT COUNT(*) FROM books`).Scan(&totalCount)
	} else {
		err = r.DB.QueryRow(`SELECT COUNT(*) FROM books WHERE MATCH(title, description) AGAINST (? IN BOOLEAN MODE)`, params.Query).Scan(&totalCount)
	}
	if err != nil {
		return nil, 0, err
	}

	return books, totalCount, nil
}
