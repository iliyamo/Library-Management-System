package router

import (
	"github.com/iliyamo/go-learning/internal/handler"
	"github.com/labstack/echo/v4"
)

// ✅ ثبت تمام روت‌های مربوط به احراز هویت
func RegisterRoutes(e *echo.Echo) {
	auth := e.Group("/auth")

	// 🟢 ثبت‌نام
	auth.POST("/register", handler.Register)

	// 🟢 ورود
	auth.POST("/login", handler.Login)

	// 🟢 دریافت پروفایل (با JWT)
	auth.GET("/profile", handler.Profile)

	// در مراحل بعد: logout و refresh هم اینجا اضافه می‌شن
}
