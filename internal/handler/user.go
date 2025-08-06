package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

    "github.com/iliyamo/Library-Management-System/internal/repository"
    "github.com/iliyamo/Library-Management-System/internal/utils"
	"github.com/labstack/echo/v4"
)

// SearchUsers ➔ GET /api/v1/users
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

	// تشکیل کلید کش منحصر به فرد براساس گزینهها
	cacheKey := fmt.Sprintf("users:query=%s:cursor=%d:limit=%d", query, cursor, limit)
	if cached, err := utils.GetCache(cacheKey); err == nil {
		var response map[string]interface{}
		if err := json.Unmarshal([]byte(cached), &response); err == nil {
			return c.JSON(http.StatusOK, response)
		}
	}

	// اجرای جستجو
	users, total, err := repo.SearchUsers(query, cursor, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "خطا در جستجو"})
	}

	nextCursor := 0
	if len(users) > 0 {
		nextCursor = users[len(users)-1].ID
	}

	response := echo.Map{
		"data":        users,
		"total":       total,
		"limit":       limit,
		"next_cursor": nextCursor,
	}

	// ذخیره نتیجه در Redis به مدت 30 ثانیه
	if data, err := json.Marshal(response); err == nil {
		_ = utils.SetCache(cacheKey, string(data), 30*time.Second)
	}

	return c.JSON(http.StatusOK, response)
}
