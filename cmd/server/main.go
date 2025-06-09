package main

import (
	"github.com/iliyamo/go-learning/internal/router" // مسیر درست رو وارد کن

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	// ثبت تمام مسیرها
	router.RegisterRoutes(e)

	// اجرا روی پورت 8080
	e.Start(":8080")
}
