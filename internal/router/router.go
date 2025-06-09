package router

import (
	handlers "github.com/iliyamo/go-learning/internal/handler" // جایگزین کن با مسیر واقعی پکیج handlers
	"github.com/labstack/echo/v4"
)

// RegisterRoutes تمام مسیرهای API را در اینجا ثبت می‌کنیم
func RegisterRoutes(e *echo.Echo) {
	// مسیرهای مربوط به احراز هویت (Authentication)
	auth := e.Group("/auth")

	// مسیر POST برای ثبت‌نام کاربر جدید
	auth.POST("/register", handlers.Register)

	// مسیر POST برای ورود کاربر
	auth.POST("/login", handlers.Login)
}
