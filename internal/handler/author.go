// internal/handler/author.go
package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

    "github.com/iliyamo/Library-Management-System/internal/model"
    "github.com/iliyamo/Library-Management-System/internal/repository"
    "github.com/iliyamo/Library-Management-System/internal/utils"
	"github.com/labstack/echo/v4"
)

// AuthorRequest ساختار ورودی برای ایجاد یا ویرایش نویسنده است
// این داده‌ها از سمت کاربر دریافت می‌شود
// و شامل نام، بیوگرافی و تاریخ تولد نویسنده است.
type AuthorRequest struct {
	Name      string `json:"name"`
	Biography string `json:"biography"`
	BirthDate string `json:"birth_date"`
}

// CreateAuthor ایجاد نویسنده جدید
func CreateAuthor(c echo.Context) error {
	repo := c.Get("author_repo").(*repository.AuthorRepository)
	req := new(AuthorRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "درخواست نامعتبر است"})
	}

	birthDate, err := time.Parse("2006-01-02", req.BirthDate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "فرمت تاریخ تولد نادرست است، از YYYY-MM-DD استفاده کنید"})
	}

	exists, err := repo.Exists(req.Name, birthDate)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "خطا در بررسی نویسنده"})
	}
	if exists {
		return c.JSON(http.StatusConflict, echo.Map{"error": "نویسنده قبلاً ثبت شده است"})
	}

	author := &model.Author{
		Name:      req.Name,
		Biography: req.Biography,
		BirthDate: birthDate,
		CreatedAt: time.Now(),
	}

	if err := repo.CreateAuthor(author); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "ثبت نویسنده با خطا مواجه شد"})
	}

	return c.JSON(http.StatusCreated, author)
}

// GetAllAuthors دریافت لیست همه نویسنده‌ها
func GetAllAuthors(c echo.Context) error {
	repo := c.Get("author_repo").(*repository.AuthorRepository)

	q := c.QueryParam("query")
	cursorStr := c.QueryParam("cursor_id")
	limitStr := c.QueryParam("limit")

	if q != "" {
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

		// ✅ کش بر اساس query + cursor + limit
		cacheKey := fmt.Sprintf("authors:query=%s:cursor=%d:limit=%d", q, cursor, limit)
		if cached, err := utils.GetCache(cacheKey); err == nil {
			var response map[string]interface{}
			if err := json.Unmarshal([]byte(cached), &response); err == nil {
				return c.JSON(http.StatusOK, response)
			}
		}

		params := &model.AuthorSearchParams{
			Query:    q,
			CursorID: cursor,
			Limit:    limit,
		}

		authors, total, err := repo.SearchAuthors(params)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "خطا در جستجو"})
		}

		nextCursor := 0
		if len(authors) > 0 {
			nextCursor = authors[len(authors)-1].ID
		}

		response := echo.Map{
			"data":        authors,
			"next_cursor": nextCursor,
			"limit":       limit,
			"total":       total,
		}
		if data, err := json.Marshal(response); err == nil {
			_ = utils.SetCache(cacheKey, string(data), 30*time.Second)
		}
		return c.JSON(http.StatusOK, response)
	}

	authors, err := repo.GetAllAuthors()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "دریافت نویسنده‌ها با خطا مواجه شد"})
	}
	return c.JSON(http.StatusOK, authors)
}

// GetAuthorByID دریافت اطلاعات نویسنده با شناسه مشخص
func GetAuthorByID(c echo.Context) error {
	repo := c.Get("author_repo").(*repository.AuthorRepository)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "شناسه نامعتبر است"})
	}
	author, err := repo.GetAuthorByID(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "خطا در دریافت نویسنده"})
	}
	if author == nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "نویسنده یافت نشد"})
	}
	return c.JSON(http.StatusOK, author)
}

// UpdateAuthor بروزرسانی اطلاعات نویسنده
func UpdateAuthor(c echo.Context) error {
	repo := c.Get("author_repo").(*repository.AuthorRepository)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "شناسه نامعتبر است"})
	}

	req := new(AuthorRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "درخواست نامعتبر است"})
	}

	birthDate, err := time.Parse("2006-01-02", req.BirthDate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "فرمت تاریخ تولد نادرست است"})
	}

	author := &model.Author{
		ID:        id,
		Name:      req.Name,
		Biography: req.Biography,
		BirthDate: birthDate,
	}

	updated, err := repo.UpdateAuthor(author)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "خطا در بروزرسانی نویسنده"})
	}
	if !updated {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "نویسنده یافت نشد"})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "نویسنده با موفقیت بروزرسانی شد"})
}

// DeleteAuthor حذف نویسنده بر اساس شناسه
func DeleteAuthor(c echo.Context) error {
	repo := c.Get("author_repo").(*repository.AuthorRepository)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "شناسه نامعتبر است"})
	}

	deleted, err := repo.DeleteAuthor(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "خطا در حذف نویسنده"})
	}
	if !deleted {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "نویسنده یافت نشد"})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "نویسنده با موفقیت حذف شد"})
}
