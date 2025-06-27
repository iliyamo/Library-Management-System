package main

import (
	"errors" // برای ساخت خطاهای سفارشی
	"fmt"    // برای قالب‌بندی رشته‌ها
	"log"    // لاگ‌کردن پیام‌های خطا و شروع برنامه

	"github.com/joho/godotenv"    // بارگذاری متغیرهای محیطی از فایل .env
	"github.com/labstack/echo/v4" // فریم‌ورک وب Echo

	"github.com/iliyamo/go-learning/internal/database"   // پکیج اتصال به دیتابیس
	"github.com/iliyamo/go-learning/internal/repository" // لایهٔ دسترسی به داده
	"github.com/iliyamo/go-learning/internal/router"     // ثبت مسیرهای HTTP
)

// App ساختاری که تمام وابستگی‌های برنامه را نگه می‌دارد:
// - Server: نمونهٔ Echo برای مدیریت HTTP
// - UserRepo: ریپازیتوری کاربران
// - RefreshRepo: ریپازیتوری توکن‌های تجدید
// این ساختار کمک می‌کند تا برنامه ماژولار و تست‌پذیر باقی بماند.
type App struct {
	Server      *echo.Echo                         // سرور HTTP
	UserRepo    *repository.UserRepository         // عملیات مرتبط با کاربران
	RefreshRepo *repository.RefreshTokenRepository // عملیات مرتبط با رفرش‌توکن‌ها
}

// NewApp وظیفهٔ ساخت و پیکربندی کامل برنامه را دارد:
// 1. بارگذاری متغیرهای محیطی
// 2. اتصال به دیتابیس
// 3. ساخت ریپازیتوری‌ها
// 4. تنظیم middleware برای تزریق ریپازیتوری‌ها
// 5. ثبت تمام مسیرهای HTTP
func NewApp() (*App, error) {
	// اگر فایل .env وجود داشته باشد، مقادیرش را در محیط قرار می‌دهد
	_ = godotenv.Load()

	// اتصال به دیتابیس با استفاده از مقادیر محیطی
	db := database.InitDB()
	if db == nil {
		// اگر اتصال موفق نبود، خطا ایجاد و برمی‌گرداند
		return nil, errors.New("database connection failed")
	}

	// ساخت لایهٔ دسترسی به داده (ریپازیتوری‌ها)
	userRepo := repository.NewUserRepository(db)
	refreshRepo := repository.NewRefreshTokenRepository(db)

	// ساخت نمونهٔ Echo برای مدیریت درخواست‌های HTTP
	e := echo.New()

	// middleware که قبل از هر درخواست اجرا می‌شود:
	// ریپازیتوری‌ها را در context هر درخواست قرار می‌دهد
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("user_repo", userRepo)
			c.Set("refresh_token_repo", refreshRepo)
			return next(c) // ادامهٔ اجرای handler اصلی
		}
	})

	// ثبت مسیرها (روت‌ها) با فراخوانی تابع RegisterRoutes
	// این تابع مسیرهای مختلف API را به سرور اضافه می‌کند
	router.RegisterRoutes(e)

	// بازگرداندن ساختار App با تمام وابستگی‌ها
	return &App{
		Server:      e,
		UserRepo:    userRepo,
		RefreshRepo: refreshRepo,
	}, nil
}

// نقطهٔ ورود برنامه
func main() {
	// ساخت برنامه و دریافت ارور احتمالی
	app, err := NewApp()
	if err != nil {
		log.Fatal(err) // اگر خطا بود، لاگ و خروج
	}

	// نمایش پیام شروع سرویس
	log.Println("Server running on http://localhost:8080")

	// اجرای سرور روی پورت 8080
	// اگر خطایی رخ دهد، لاگ و خروج
	if err := app.Server.Start(":8080"); err != nil {
		log.Fatal(fmt.Errorf("server error: %w", err))
	}
}
