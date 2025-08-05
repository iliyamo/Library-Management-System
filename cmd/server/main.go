// cmd/server/main.go
package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"

	"github.com/iliyamo/go-learning/internal/database"
	"github.com/iliyamo/go-learning/internal/repository"
	"github.com/iliyamo/go-learning/internal/router"
	"github.com/iliyamo/go-learning/internal/utils" // ✅ اضافه‌شده برای Redis
    "github.com/iliyamo/go-learning/internal/queue" // ✅ اضافه‌شده برای RabbitMQ
)

// App ساختار کلی برنامه شامل وابستگی‌ها
type App struct {
	Server      *echo.Echo
	UserRepo    *repository.UserRepository
	RefreshRepo *repository.RefreshTokenRepository
	AuthorRepo  *repository.AuthorRepository
	BookRepo    *repository.BookRepository // ✅ مدیریت کتاب‌ها
	LoanRepo    *repository.LoanRepository // ✅ مدیریت امانت‌ها
}

// NewApp مقداردهی اولیهٔ برنامه
func NewApp() (*App, error) {
	cwd, _ := os.Getwd()
	log.Println("Current working directory:", cwd)

	_ = godotenv.Load("../../.env")

	user := os.Getenv("DB_USER")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	name := os.Getenv("DB_NAME")
	if user == "" || host == "" || port == "" || name == "" {
		log.Println("\u274c لطفاً متغیرهای DB_USER, DB_HOST, DB_PORT و DB_NAME را تنظیم کنید.")
		return nil, errors.New("database connection failed")
	}

	db := database.InitDB()
	if db == nil {
		return nil, errors.New("database connection failed")
	}

    // ✅ اتصال به Redis
    utils.InitRedis()
    // ✅ مقداردهی RabbitMQ در صورت تنظیم متغیر محیطی
    queue.InitQueue()

	userRepo := repository.NewUserRepository(db)
	refreshRepo := repository.NewRefreshTokenRepository(db)
	authorRepo := repository.NewAuthorRepository(db)
	bookRepo := repository.NewBookRepository(db)
	loanRepo := repository.NewLoanRepository(db) // ✅ ساخت ریپازیتوری امانت‌ها

	e := echo.New()
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("user_repo", userRepo)
			c.Set("refresh_token_repo", refreshRepo)
			c.Set("author_repo", authorRepo)
			c.Set("book_repo", bookRepo)
			c.Set("loan_repo", loanRepo)
			return next(c)
		}
	})

	router.RegisterRoutes(e)

	return &App{
		Server:      e,
		UserRepo:    userRepo,
		RefreshRepo: refreshRepo,
		AuthorRepo:  authorRepo,
		BookRepo:    bookRepo,
		LoanRepo:    loanRepo,
	}, nil
}

// main نقطهٔ شروع برنامه
func main() {
	app, err := NewApp()
	if err != nil {
		log.Fatal(err)
	}

    // بستن RabbitMQ پس از اتمام اجرای سرور
    defer queue.CloseRabbitMQ()

    log.Println("Server running on http://localhost:8080")
    if err := app.Server.Start(":8080"); err != nil {
        log.Fatal(fmt.Errorf("server error: %w", err))
    }
}
