// internal/handler/book.go
package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/iliyamo/go-learning/internal/model"
	"github.com/iliyamo/go-learning/internal/repository"
	"github.com/iliyamo/go-learning/internal/utils"
	"github.com/labstack/echo/v4"
)

// BookRequest داده‌ای که از کلاینت می‌گیریم
type BookRequest struct {
	Title         string  `json:"title"`
	ISBN          string  `json:"isbn"`
	AuthorID      int     `json:"author_id"`
	CategoryID    *int    `json:"category_id"`
	Description   *string `json:"description"`
	PublishedYear *int    `json:"published_year"`
	TotalCopies   int     `json:"total_copies"`
}

// CreateBook ➔ POST /books
func CreateBook(c echo.Context) error {
	repo := c.Get("book_repo").(*repository.BookRepository)
	var req BookRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "درخواست نامعتبر"})
	}
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
		AvailableCopies: req.TotalCopies,
		CreatedAt:       time.Now(),
	}
	if err := repo.CreateBook(book); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "ثبت کتاب ناموفق"})
	}
	return c.JSON(http.StatusCreated, book)
}

// GetAllBooks ➔ GET /books
func GetAllBooks(c echo.Context) error {
	repo := c.Get("book_repo").(*repository.BookRepository)

	query := c.QueryParam("query")
	cursorStr := c.QueryParam("cursor_id")
	limitStr := c.QueryParam("limit")

	cursor := 0
	if cursorStr != "" {
		if v, err := strconv.Atoi(cursorStr); err == nil {
			cursor = v
		}
	}

	limit := 10
	if limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}

	// ✅ کش‌کردن نتایج بر اساس پارامترهای جستجو
	cacheKey := fmt.Sprintf("books:query=%s:cursor=%d:limit=%d", query, cursor, limit)
	if cached, err := utils.GetCache(cacheKey); err == nil {
		var response map[string]interface{}
		if err := json.Unmarshal([]byte(cached), &response); err == nil {
			return c.JSON(http.StatusOK, response)
		}
	}

	params := &model.BookSearchParams{
		Query:    query,
		CursorID: cursor,
		Limit:    limit,
	}

	books, _, err := repo.SearchBooks(params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "خطا در جستجو"})
	}

	nextCursor := 0
	if len(books) > 0 {
		nextCursor = books[len(books)-1].ID
	}

	response := echo.Map{
		"data":        books,
		"next_cursor": nextCursor,
		"limit":       limit,
	}

	// ✅ ذخیره در Redis برای ۳۰ ثانیه
	if data, err := json.Marshal(response); err == nil {
		_ = utils.SetCache(cacheKey, string(data), 30*time.Second)
	}

	return c.JSON(http.StatusOK, response)
}

// GetBookByID ➔ GET /books/:id
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

// UpdateBook ➔ PUT /books/:id
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
		AvailableCopies: req.TotalCopies,
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

// DeleteBook ➔ DELETE /books/:id
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
