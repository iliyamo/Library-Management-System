// cmd/server/main.go
package main

import (
	"errors" // برای ساخت خطاهای سفارشی
	"fmt"    // قالب‌بندی متن لاگ
	"log"    // چاپ لاگ در کنسول
	"os"     // خواندن متغیرهای محیطی و دریافت مسیر کاری

	"github.com/joho/godotenv"    // بارگذاری متغیرهای محیطی از فایل .env
	"github.com/labstack/echo/v4" // فریم‌ورک وب Echo

	"github.com/iliyamo/go-learning/internal/database"   // اتصال به دیتابیس
	"github.com/iliyamo/go-learning/internal/repository" // لایهٔ دسترسی به داده
	"github.com/iliyamo/go-learning/internal/router"     // ثبت مسیرهای HTTP
)

// App تمام وابستگی‌های برنامه را نگه می‌دارد تا ماژولار و تست‌پذیر باشد.
type App struct {
	Server      *echo.Echo                         // سرور HTTP
	UserRepo    *repository.UserRepository         // لایهٔ دسترسی کاربران
	RefreshRepo *repository.RefreshTokenRepository // لایهٔ دسترسی رفرش‌توکن‌ها
	AuthorRepo  *repository.AuthorRepository       // لایهٔ دسترسی نویسنده‌ها
}

// NewApp برنامه را راه‌اندازی می‌کند:
// 1) بارگذاری .env
// 2) اتصال به دیتابیس
// 3) ساخت ریپازیتوری‌ها
// 4) تزریق آن‌ها در middleware
// 5) ثبت مسیرهای API
func NewApp() (*App, error) {
	// نمایش مسیر کاری فعلی برای دیباگ
	cwd, _ := os.Getwd()
	log.Println("Current working directory:", cwd)

	// بارگذاری متغیرهای محیطی از فایل .env
	_ = godotenv.Load("../../.env")

	// خواندن متغیرهای اتصال
	user := os.Getenv("DB_USER")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	name := os.Getenv("DB_NAME")
	if user == "" || host == "" || port == "" || name == "" {
		log.Println("❌ لطفاً متغیرهای DB_USER, DB_HOST, DB_PORT و DB_NAME را تنظیم کنید.")
		return nil, errors.New("database connection failed")
	}

	// اتصال به دیتابیس با استفاده از مقادیر محیطی
	db := database.InitDB()
	if db == nil {
		return nil, errors.New("database connection failed")
	}

	// ایجاد ریپازیتوری‌ها
	userRepo := repository.NewUserRepository(db)
	refreshRepo := repository.NewRefreshTokenRepository(db)
	authorRepo := repository.NewAuthorRepository(db)

	// ساخت سرور Echo
	e := echo.New()

	// middleware برای تزریق ریپازیتوری‌ها به Context هر درخواست
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("user_repo", userRepo)
			c.Set("refresh_token_repo", refreshRepo)
			c.Set("author_repo", authorRepo)
			return next(c)
		}
	})

	// ثبت مسیرهای HTTP
	router.RegisterRoutes(e)

	return &App{Server: e, UserRepo: userRepo, RefreshRepo: refreshRepo, AuthorRepo: authorRepo}, nil
}

// main نقطهٔ ورود برنامه است
func main() {
	app, err := NewApp()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Server running on http://localhost:8080")
	if err := app.Server.Start(":8080"); err != nil {
		log.Fatal(fmt.Errorf("server error: %w", err))
	}
}
