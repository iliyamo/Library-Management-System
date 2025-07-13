package router

import (
	"github.com/labstack/echo/v4"

	"github.com/iliyamo/go-learning/internal/handler"
	"github.com/iliyamo/go-learning/internal/middleware"
)

// RegisterRoutes تمام مسیرهای مربوط به نسخه اول API را ثبت می‌کند.
func RegisterRoutes(e *echo.Echo) {
	// ✅ مسیر پایه برای API نسخه ۱
	v1 := e.Group("/api/v1")

	// ================================
	// 📌 مسیرهای عمومی (بدون نیاز به احراز هویت)
	// ================================
	auth := v1.Group("/auth")
	auth.POST("/register", handler.Register) // ثبت‌نام کاربر
	auth.POST("/login", handler.Login)       // ورود کاربر
	auth.POST("/refresh", handler.RefreshToken)

	// ================================
	// 🔒 مسیرهای محافظت‌شده با JWT
	// ================================
	auth.Use(middleware.JWTAuth)          // استفاده از middleware برای محافظت از مسیرها
	auth.GET("/profile", handler.Profile) // دریافت اطلاعات پروفایل
	auth.POST("/logout", handler.Logout)  // خروج از سیستم

	// ✍ مسیرهای نویسنده (محافظت‌شده)
	authors := v1.Group("/authors")
	authors.Use(middleware.JWTAuth)              // احراز هویت الزامی است
	authors.POST("", handler.CreateAuthor)       // ایجاد نویسنده
	authors.GET("", handler.GetAllAuthors)       // دریافت همه نویسندگان
	authors.GET("/:id", handler.GetAuthorByID)   // دریافت یک نویسنده خاص
	authors.PUT("/:id", handler.UpdateAuthor)    // بروزرسانی نویسنده
	authors.DELETE("/:id", handler.DeleteAuthor) // حذف نویسنده

	// 📚 مسیرهای کتاب‌ها (محافظت‌شده)
	books := v1.Group("/books")
	books.Use(middleware.JWTAuth)            // احراز هویت الزامی است
	books.POST("", handler.CreateBook)       // ایجاد کتاب جدید
	books.GET("", handler.GetAllBooks)       // دریافت لیست همه کتاب‌ها
	books.GET("/:id", handler.GetBookByID)   // دریافت اطلاعات یک کتاب خاص
	books.PUT("/:id", handler.UpdateBook)    // بروزرسانی اطلاعات کتاب
	books.DELETE("/:id", handler.DeleteBook) // حذف کتاب
}
