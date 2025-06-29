package router

import (
	"github.com/labstack/echo/v4"

	"github.com/iliyamo/go-learning/internal/handler"
	"github.com/iliyamo/go-learning/internal/middleware"
)

// RegisterRoutes همه مسیرهای مربوط به نسخه ۱ از API را ثبت می‌کند.
// این روش به ما اجازه می‌دهد تا در آینده نسخه‌های جدید را راحت‌تر مدیریت کنیم.
func RegisterRoutes(e *echo.Echo) {
	// ✅ مسیر پایه برای API نسخه ۱
	v1 := e.Group("/api/v1")

	// ================================
	// 📌 مسیرهای عمومی (بدون نیاز به JWT)
	// ================================

	auth := v1.Group("/auth")
	auth.POST("/register", handler.Register) // ثبت‌نام
	auth.POST("/login", handler.Login)       // ورود

	// ================================
	// 🔒 مسیرهای محافظت‌شده با JWT
	// ================================

	// اعمال middleware اعتبارسنجی JWT به مسیرهای auth محافظت‌شده
	auth.Use(middleware.JWTAuth)
	auth.GET("/profile", handler.Profile) // دریافت پروفایل کاربر
	auth.POST("/logout", handler.Logout)  // خروج کاربر و حذف refresh token

	// ================================
	// ✍ مسیرهای نویسنده (محافظت‌شده)
	// ================================

	authors := v1.Group("/authors")
	authors.Use(middleware.JWTAuth)              // همه مسیرهای نویسنده نیاز به احراز هویت دارند
	authors.POST("", handler.CreateAuthor)       // ایجاد نویسنده جدید
	authors.GET("", handler.GetAllAuthors)       // لیست همه نویسنده‌ها
	authors.GET("/:id", handler.GetAuthorByID)   // دریافت نویسنده خاص با شناسه
	authors.PUT("/:id", handler.UpdateAuthor)    // ویرایش اطلاعات نویسنده
	authors.DELETE("/:id", handler.DeleteAuthor) // حذف نویسنده
}
