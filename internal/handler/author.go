// internal/handler/author.go

package handler

import (
	"database/sql"
	"errors" // برای بررسی خطاهایی مثل sql.ErrNoRows
	"net/http"
	"strconv"
	"time"

	"github.com/iliyamo/go-learning/internal/model"
	"github.com/iliyamo/go-learning/internal/repository"
	"github.com/labstack/echo/v4"
)

// ساختار ورودی برای ایجاد یا ویرایش نویسنده
// زمانی که کاربر یک نویسنده جدید ایجاد می‌کند یا اطلاعاتش را ویرایش می‌کند،
// این ساختار داده از بدنهٔ درخواست گرفته می‌شود.
type AuthorRequest struct {
	Name      string `json:"name"`       // نام نویسنده
	Biography string `json:"biography"`  // زندگی‌نامه نویسنده
	BirthDate string `json:"birth_date"` // تاریخ تولد نویسنده (به صورت رشته)
}

// 🔹 ایجاد نویسنده جدید
func CreateAuthor(c echo.Context) error {
	repo := c.Get("author_repo").(*repository.AuthorRepository)
	req := new(AuthorRequest)

	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}

	birthDate, err := time.Parse("2006-01-02", req.BirthDate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid birth_date format, use YYYY-MM-DD"})
	}

	author := &model.Author{
		Name:      req.Name,
		Biography: req.Biography,
		BirthDate: birthDate,
		CreatedAt: time.Now(),
	}

	if err := repo.CreateAuthor(author); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to create author"})
	}

	return c.JSON(http.StatusCreated, author)
}

// 🔹 دریافت لیست همه نویسنده‌ها
func GetAllAuthors(c echo.Context) error {
	repo := c.Get("author_repo").(*repository.AuthorRepository)

	authors, err := repo.GetAllAuthors()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to fetch authors"})
	}

	return c.JSON(http.StatusOK, authors)
}

// 🔹 دریافت نویسنده بر اساس شناسه (id)
func GetAuthorByID(c echo.Context) error {
	repo := c.Get("author_repo").(*repository.AuthorRepository)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid id"})
	}

	author, err := repo.GetAuthorByID(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to fetch author"})
	}
	if author == nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "author not found"})
	}

	return c.JSON(http.StatusOK, author)
}

// 🔹 بروزرسانی اطلاعات نویسنده
func UpdateAuthor(c echo.Context) error {
	repo := c.Get("author_repo").(*repository.AuthorRepository)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid id"})
	}

	req := new(AuthorRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}

	birthDate, err := time.Parse("2006-01-02", req.BirthDate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid birth_date format, use YYYY-MM-DD"})
	}

	author := &model.Author{
		ID:        id,
		Name:      req.Name,
		Biography: req.Biography,
		BirthDate: birthDate,
	}

	if err := repo.UpdateAuthor(author); err != nil {
		// بررسی اینکه آیا خطا به‌خاطر نبودن نویسنده است
		if errors.Is(err, sql.ErrNoRows) {
			return c.JSON(http.StatusNotFound, echo.Map{"error": "author not found"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to update author"})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "author updated"})
}

// 🔹 حذف نویسنده با شناسه خاص
func DeleteAuthor(c echo.Context) error {
	repo := c.Get("author_repo").(*repository.AuthorRepository)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid id"})
	}

	if err := repo.DeleteAuthor(id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.JSON(http.StatusNotFound, echo.Map{"error": "author not found"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to delete author"})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "author deleted"})
}
