package database

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql" // درایور MySQL
)

func InitDB() *sql.DB {
	// تنظیم رشته اتصال (Data Source Name)
	dsn := "root:@tcp(localhost:3306)/digital_library?parseTime=true"

	// باز کردن اتصال به دیتابیس
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("خطا در باز کردن اتصال دیتابیس: %v", err)
	}

	// تست اتصال
	if err := db.Ping(); err != nil {
		log.Fatalf("خطا در برقراری اتصال دیتابیس: %v", err)
	}

	log.Println("اتصال به دیتابیس موفق بود")
	return db
}
