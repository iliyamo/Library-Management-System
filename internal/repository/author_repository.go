// internal/repository/author_repository.go
package repository

import (
	"database/sql" // برای کار با پایگاه داده
	"errors"       // برای بررسی خطاها مثل sql.ErrNoRows
	"time"         // برای کار با زمان و تاریخ

	"github.com/iliyamo/go-learning/internal/model" // مدل داده‌ها (Author)
)

// AuthorRepository ساختاری است برای مدیریت عملیات دیتابیس مربوط به نویسنده‌ها
// شامل متدهایی برای ایجاد، خواندن، بروزرسانی و حذف اطلاعات نویسنده
type AuthorRepository struct {
	DB *sql.DB // اتصال به پایگاه داده
}

// NewAuthorRepository تابع سازنده برای ایجاد نمونهٔ جدید از AuthorRepository
func NewAuthorRepository(db *sql.DB) *AuthorRepository {
	return &AuthorRepository{DB: db}
}

// ✅ ایجاد نویسنده جدید
// این تابع اطلاعات نویسنده را در جدول authors وارد می‌کند
func (r *AuthorRepository) CreateAuthor(author *model.Author) error {
	query := `
		INSERT INTO authors (name, biography, birth_date, created_at)
		VALUES (?, ?, ?, ?)
	`
	// اجرای کوئری و ارسال مقادیر ورودی به صورت امن (با استفاده از prepared statement)
	_, err := r.DB.Exec(query, author.Name, author.Biography, author.BirthDate, time.Now())
	return err
}

// ✅ دریافت همه نویسنده‌ها
// این تابع تمام نویسنده‌ها را از جدول authors واکشی کرده و در قالب slice بازمی‌گرداند
func (r *AuthorRepository) GetAllAuthors() ([]model.Author, error) {
	query := `SELECT id, name, biography, birth_date, created_at FROM authors`
	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close() // بستن اتصال پس از پایان خواندن داده‌ها

	var authors []model.Author
	for rows.Next() {
		var a model.Author
		err := rows.Scan(&a.ID, &a.Name, &a.Biography, &a.BirthDate, &a.CreatedAt)
		if err != nil {
			return nil, err
		}
		authors = append(authors, a)
	}

	return authors, nil
}

// ✅ دریافت نویسنده با شناسه
// این تابع اطلاعات یک نویسنده خاص را بر اساس ID دریافت می‌کند
func (r *AuthorRepository) GetAuthorByID(id int) (*model.Author, error) {
	query := `SELECT id, name, biography, birth_date, created_at FROM authors WHERE id = ?`
	var a model.Author
	// استفاده از QueryRow چون فقط یک نتیجه انتظار داریم
	err := r.DB.QueryRow(query, id).Scan(&a.ID, &a.Name, &a.Biography, &a.BirthDate, &a.CreatedAt)
	if err != nil {
		// اگر نویسنده‌ای پیدا نشد، nil برمی‌گردد نه خطا
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

// ✅ ویرایش اطلاعات نویسنده
// این تابع اطلاعات یک نویسنده را بر اساس ID بروزرسانی می‌کند
func (r *AuthorRepository) UpdateAuthor(a *model.Author) error {
	query := `
		UPDATE authors
		SET name = ?, biography = ?, birth_date = ?
		WHERE id = ?
	`
	_, err := r.DB.Exec(query, a.Name, a.Biography, a.BirthDate, a.ID)
	return err
}

// ✅ حذف نویسنده
// این تابع نویسنده‌ای را بر اساس ID حذف می‌کند
func (r *AuthorRepository) DeleteAuthor(id int) error {
	query := `DELETE FROM authors WHERE id = ?`
	_, err := r.DB.Exec(query, id)
	return err
}
