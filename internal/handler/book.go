// internal/handler/book.go
package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/iliyamo/go-learning/internal/model"
	"github.com/iliyamo/go-learning/internal/repository"
	"github.com/labstack/echo/v4"
)

// BookRequest داده‎ای که از کلاینت می‌گیریم
type BookRequest struct {
	Title         string  `json:"title"`
	ISBN          string  `json:"isbn"`
	AuthorID      int     `json:"author_id"`
	CategoryID    *int    `json:"category_id"`
	Description   *string `json:"description"`
	PublishedYear *int    `json:"published_year"`
	TotalCopies   int     `json:"total_copies"`
}

// CreateBook ➜ POST /books
func CreateBook(c echo.Context) error {
	repo := c.Get("book_repo").(*repository.BookRepository)

	// تبدیل JSON ورودی
	var req BookRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "درخواست نامعتبر"})
	}

	// جلوگیری از ISBN تکراری
	if ok, err := repo.ExistsByISBN(req.ISBN); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "خطا در بررسی ISBN"})
	} else if ok {
		return c.JSON(http.StatusConflict, echo.Map{"error": "این ISBN قبلاً ثبت شده"})
	}

	book := &model.Book{
		Title:           req.Title,
		ISBN:            req.ISBN,
		AuthorID:        req.AuthorID,
		CategoryID:      req.CategoryID,
		Description:     req.Description,
		PublishedYear:   req.PublishedYear,
		TotalCopies:     req.TotalCopies,
		AvailableCopies: req.TotalCopies, // در ابتدا همه نسخه‌ها موجودند
		CreatedAt:       time.Now(),
	}

	if err := repo.CreateBook(book); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "ثبت کتاب ناموفق"})
	}
	return c.JSON(http.StatusCreated, book)
}

// GetAllBooks ➜ GET /books
func GetAllBooks(c echo.Context) error {
	repo := c.Get("book_repo").(*repository.BookRepository)
	books, err := repo.GetAllBooks()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "خطا در واکشی"})
	}
	return c.JSON(http.StatusOK, books)
}

// GetBookByID ➜ GET /books/:id
func GetBookByID(c echo.Context) error {
	repo := c.Get("book_repo").(*repository.BookRepository)
	id, _ := strconv.Atoi(c.Param("id"))

	book, err := repo.GetBookByID(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "خطا در واکشی"})
	}
	if book == nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "کتاب پیدا نشد"})
	}
	return c.JSON(http.StatusOK, book)
}

// UpdateBook ➜ PUT /books/:id
func UpdateBook(c echo.Context) error {
	repo := c.Get("book_repo").(*repository.BookRepository)
	id, _ := strconv.Atoi(c.Param("id"))

	var req BookRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "درخواست نامعتبر"})
	}

	book := &model.Book{
		ID:              id,
		ISBN:            req.ISBN,
		Title:           req.Title,
		AuthorID:        req.AuthorID,
		CategoryID:      req.CategoryID,
		Description:     req.Description,
		PublishedYear:   req.PublishedYear,
		TotalCopies:     req.TotalCopies,
		AvailableCopies: req.TotalCopies, // می‌توانید منطق پیچیده‌تری بگذارید
	}

	ok, err := repo.UpdateBook(book)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "خطا در بروزرسانی"})
	}
	if !ok {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "کتاب پیدا نشد"})
	}
	return c.JSON(http.StatusOK, echo.Map{"message": "کتاب بروزرسانی شد"})
}

// DeleteBook ➜ DELETE /books/:id
func DeleteBook(c echo.Context) error {
	repo := c.Get("book_repo").(*repository.BookRepository)
	id, _ := strconv.Atoi(c.Param("id"))

	ok, err := repo.DeleteBook(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "خطا در حذف"})
	}
	if !ok {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "کتاب پیدا نشد"})
	}
	return c.JSON(http.StatusOK, echo.Map{"message": "کتاب حذف شد"})
}
