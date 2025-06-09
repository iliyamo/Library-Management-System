package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// Register هندلری برای ثبت‌نام کاربران جدید
func Register(c echo.Context) error {
	// فعلاً فقط یک پیام ساده برمی‌گرداند
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Register endpoint hit!",
	})
}

// Login هندلری برای ورود کاربران
func Login(c echo.Context) error {
	// فعلاً فقط یک پیام ساده برمی‌گرداند
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Login endpoint hit!",
	})
}
