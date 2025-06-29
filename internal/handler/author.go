// internal/handler/author.go
package handler

import (
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
	repo := c.Get("author_repo").(*repository.AuthorRepository) // گرفتن ریپازیتوری از context
	req := new(AuthorRequest)                                   // ساختار برای گرفتن دادهٔ ورودی

	// تبدیل JSON ورودی به ساختار Go
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}

	// تبدیل تاریخ تولد از رشته به time.Time
	birthDate, err := time.Parse("2006-01-02", req.BirthDate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid birth_date format, use YYYY-MM-DD"})
	}

	// ساخت نویسنده جدید بر اساس دادهٔ ورودی
	author := &model.Author{
		Name:      req.Name,
		Biography: req.Biography,
		BirthDate: birthDate,
		CreatedAt: time.Now(),
	}

	// ذخیره در دیتابیس
	if err := repo.CreateAuthor(author); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to create author"})
	}

	return c.JSON(http.StatusCreated, author) // نویسنده با موفقیت ایجاد شد
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

	// گرفتن مقدار id از پارامتر URL
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid id"})
	}

	// گرفتن نویسنده از دیتابیس
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

	// دریافت id از URL
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid id"})
	}

	req := new(AuthorRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}

	// تبدیل رشته تاریخ تولد به time.Time
	birthDate, err := time.Parse("2006-01-02", req.BirthDate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid birth_date format, use YYYY-MM-DD"})
	}

	// آماده‌سازی داده برای بروزرسانی
	author := &model.Author{
		ID:        id,
		Name:      req.Name,
		Biography: req.Biography,
		BirthDate: birthDate,
	}

	// بروزرسانی در دیتابیس
	if err := repo.UpdateAuthor(author); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to update author"})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "author updated"})
}

// 🔹 حذف نویسنده با شناسه خاص
func DeleteAuthor(c echo.Context) error {
	repo := c.Get("author_repo").(*repository.AuthorRepository)

	// دریافت id از URL
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid id"})
	}

	// حذف نویسنده از دیتابیس
	if err := repo.DeleteAuthor(id); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to delete author"})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "author deleted"})
}
