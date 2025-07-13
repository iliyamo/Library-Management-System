package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

// InitDB اتصال به دیتابیس را برقرار می‌کند و در صورت موفق بودن، یک شیء *sql.DB بازمی‌گرداند
func InitDB() *sql.DB {
	// تلاش برای بارگذاری متغیرهای محیطی از فایل .env
	_ = godotenv.Load()

	// خواندن اطلاعات اتصال از محیط
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	name := os.Getenv("DB_NAME")

	// بررسی کامل بودن متغیرها
	if user == "" || host == "" || port == "" || name == "" {
		log.Println("❌ لطفاً متغیرهای DB_USER, DB_HOST, DB_PORT و DB_NAME را تنظیم کنید.")
		return nil
	}

	// ساخت رشته اتصال
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", user, pass, host, port, name)

	// تلاش برای اتصال
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Println("❌ اتصال به دیتابیس ناموفق:", err)
		return nil
	}

	// تست اتصال
	if err := db.Ping(); err != nil {
		log.Println("❌ دیتابیس در دسترس نیست:", err)
		return nil
	}

	log.Println("✅ اتصال موفق به دیتابیس")
	return db
}
