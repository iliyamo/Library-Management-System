// internal/router/router.go
package router

import (
	"github.com/labstack/echo/v4" // فریم‌ورک وب Echo برای مدیریت درخواست‌ها

	"github.com/iliyamo/go-learning/internal/handler"    // حاوی منطق پردازش درخواست‌ها (handlerها)
	"github.com/iliyamo/go-learning/internal/middleware" // میدل‌ویر سفارشی مثل اعتبارسنجی JWT
)

// RegisterRoutes تمام مسیرهای API را زیر /api/v1 ثبت می‌کند
// این ساختار به ما امکان می‌دهد نسخه‌های مختلف API را به سادگی مدیریت کنیم
func RegisterRoutes(e *echo.Echo) {
	// ساخت گروه برای نسخهٔ 1 از API: تمامی مسیرها با /api/v1 آغاز می‌شوند
	v1 := e.Group("/api/v1")

	// ===== مسیرهای عمومی (نیازی به ورود و توکن JWT ندارند) =====
	// گروه auth مربوط به عملیات ثبت‌نام و ورود کاربر است
	auth := v1.Group("/auth")

	// ثبت‌نام کاربر: ارسال اطلاعات کاربر جدید
	auth.POST("/register", handler.Register)
	// ورود کاربر: بررسی اعتبار و برگرداندن توکن‌های دسترسی
	auth.POST("/login", handler.Login)

	// ===== مسیرهای محافظت‌شده (نیاز به JWT دارند) =====
	// تمام مسیرهای بعد از این خط ابتدا باید توکن را بررسی کنند
	auth.Use(middleware.JWTAuth) // middleware.JWTAuth هدر Authorization را می‌خواند و توکن را چک می‌کند

	// دریافت پروفایل کاربر جاری: اطلاعات کاربر از Claims خوانده می‌شود
	auth.GET("/profile", handler.Profile)
	// خروج کاربر: لغو توکن‌های تجدید ذخیره‌شده در دیتابیس
	auth.POST("/logout", handler.Logout)

	// ===== افزودن گروه‌ها و مسیرهای دیگر =====
	// برای مثال می‌توانید گروه /authors یا /books را مشابه auth تعریف کنید
	// authors := v1.Group("/authors")
	// authors.Use(middleware.JWTAuth)
	// authors.GET("", handler.GetAllAuthors)
}
