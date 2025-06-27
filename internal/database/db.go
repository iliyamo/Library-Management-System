package database

import (
	"database/sql" // پکیج استاندارد برای کار با دیتابیس
	"fmt"          // برای ساخت رشته اتصال
	"log"          // برای لاگ کردن خطاها و پیام‌ها
	"os"           // برای خواندن متغیرهای محیطی

	// درایور MySQL
	_ "github.com/go-sql-driver/mysql"
	// بارگذاری فایل .env به محیط سیستم
	"github.com/joho/godotenv"
)

// InitDB تنظیمات اتصال به دیتابیس را از .env می‌خواند و یک *sql.DB بازمی‌گرداند.
// اگر متغیرهای مورد نیاز وجود نداشته باشد یا اتصال ناموفق باشد، برنامه را متوقف می‌کند.
func InitDB() *sql.DB {
	// تلاش برای بارگذاری متغیرهای محیطی از فایل .env
	// اگر فایل موجود نباشد، ادامه می‌دهد بدون خطا
	_ = godotenv.Load()

	// خواندن مقادیر مورد نیاز از متغیرهای محیطی
	user := os.Getenv("DB_USER")     // نام کاربری دیتابیس
	pass := os.Getenv("DB_PASSWORD") // رمز عبور دیتابیس
	host := os.Getenv("DB_HOST")     // آدرس میزبانی دیتابیس
	port := os.Getenv("DB_PORT")     // پورت دیتابیس
	name := os.Getenv("DB_NAME")     // نام دیتابیس

	// اگر هر یک از مقادیر حیاتی تهی بود، برنامه را با پیغام مناسب متوقف می‌کنیم
	if user == "" || host == "" || port == "" || name == "" {
		log.Fatal("❌ لطفاً متغیرهای DB_USER, DB_HOST, DB_PORT و DB_NAME را تنظیم کنید.")
	}

	// ساخت رشته اتصال به فرمت MySQL:
	// user:pass@tcp(host:port)/dbname?parseTime=true
	// parseTime=true برای تبدیل خودکار DATETIME به time.Time
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true",
		user, pass, host, port, name,
	)

	// باز کردن اتصال به دیتابیس؛ این مرحله فقط یک handle ایجاد می‌کند
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		// اگر ساخت handle هم خطا داد، لاگ و خروج
		log.Fatalf("❌ خطا در sql.Open: %v", err)
	}

	// Ping برای آزمایش واقعی اتصال به دیتابیس
	if err := db.Ping(); err != nil {
		// اگر دیتابیس در دسترس نیست یا credentials اشتباه است، خروج
		log.Fatalf("❌ اتصال به دیتابیس برقرار نشد: %v", err)
	}

	// اگر همه‌چیز اوکی باشد پیغام موفقیت‌آمیز چاپ می‌شود
	log.Println("✅ اتصال به دیتابیس موفق بود")
	return db // return handle برای استفاده در بقیه برنامه
}
