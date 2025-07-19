// internal/handler/user.go
package handler

import (
	"net/http"
	"strconv"

	"github.com/iliyamo/go-learning/internal/repository"
	"github.com/labstack/echo/v4"
)

// SearchUsers ➜ GET /api/v1/users
// جستجوی کاربران بر اساس full_name و email با پشتیبانی از cursor-based pagination
func SearchUsers(c echo.Context) error {
	repo := c.Get("user_repo").(*repository.UserRepository)

	// استخراج پارامترهای کوئری از URL
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

	// اجرای جستجو
	users, total, err := repo.SearchUsers(query, cursor, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "خطا در جستجو"})
	}

	var nextCursor int
	if len(users) > 0 {
		nextCursor = users[len(users)-1].ID
	}

	return c.JSON(http.StatusOK, echo.Map{
		"data":        users,
		"total":       total,
		"limit":       limit,
		"next_cursor": nextCursor,
	})
}
