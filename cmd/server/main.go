package main

import (
	"log"

	"github.com/iliyamo/go-learning/internal/database"
	handlers "github.com/iliyamo/go-learning/internal/handler"
	"github.com/iliyamo/go-learning/internal/repository"
	"github.com/labstack/echo/v4"
)

// متغیر جهانی برای UserRepository
var userRepo *repository.UserRepository

func main() {
	// اتصال به دیتابیس و گرفتن شیء DB
	db := database.InitDB()
	if db == nil {
		log.Fatal("اتصال به دیتابیس موفق نبود!")
	}

	// ساخت UserRepository با DB
	userRepo = repository.NewUserRepository(db)

	// ساخت echo
	e := echo.New()

	// میدل‌ویر برای اتصال UserRepository به کانتکست echo
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// هر درخواست اینجا UserRepository را در context می‌گذارد
			c.Set("db", userRepo)
			return next(c)
		}
	})

	// ثبت مسیرها
	e.POST("/register", handlers.Register)
	e.POST("/login", handlers.Login)

	// اجرای سرور روی پورت 8080
	e.Logger.Fatal(e.Start(":8080"))
}
