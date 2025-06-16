package main

import (
	"log"

	"github.com/iliyamo/go-learning/internal/database"
	"github.com/iliyamo/go-learning/internal/repository"
	"github.com/iliyamo/go-learning/internal/router"
	"github.com/labstack/echo/v4"
)

func main() {
	db := database.InitDB()
	if db == nil {
		log.Fatal("Database connection failed")
	}

	userRepo := repository.NewUserRepository(db)
	refreshRepo := repository.NewRefreshTokenRepository(db)

	e := echo.New()

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("db", userRepo)
			c.Set("refresh_token_repo", refreshRepo)
			return next(c)
		}
	})

	router.RegisterRoutes(e)

	e.Logger.Fatal(e.Start(":8080"))
}
