package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	redis "github.com/redis/go-redis/v9"
	amqp "github.com/streadway/amqp"

	"github.com/iliyamo/Library-Management-System/internal/database"
	"github.com/iliyamo/Library-Management-System/internal/model" // برای LoanEvent در handler consumer
	"github.com/iliyamo/Library-Management-System/internal/queue"
	"github.com/iliyamo/Library-Management-System/internal/repository"
	"github.com/iliyamo/Library-Management-System/internal/router"
	"github.com/iliyamo/Library-Management-System/internal/utils"
)

// تعریف ساختار اصلی اپلیکیشن که شامل سرویس‌های مختلف می‌شود
type App struct {
	Server      *echo.Echo                         // سرور Echo برای مدیریت درخواست‌ها
	UserRepo    *repository.UserRepository         // مخزن داده‌های کاربر
	RefreshRepo *repository.RefreshTokenRepository // مخزن توکن‌های ریفرش
	AuthorRepo  *repository.AuthorRepository       // مخزن داده‌های نویسنده
	BookRepo    *repository.BookRepository         // مخزن داده‌های کتاب
	LoanRepo    *repository.LoanRepository         // مخزن داده‌های وام
	RabbitConn  *amqp.Connection                   // اتصال به RabbitMQ
	RabbitChan  *amqp.Channel                      // کانال RabbitMQ
	Redis       *redis.Client                      // اتصال به Redis
}

// تابع NewApp برای راه‌اندازی و پیکربندی تمام اجزای سیستم
func NewApp() (*App, error) {
	// بارگذاری متغیرهای محیطی از فایل‌های .env
	_ = godotenv.Load()             // بارگذاری تنظیمات از فایل .env
	_ = godotenv.Load("../../.env") // بارگذاری تنظیمات از فایل .env در مسیر دیگر

	// دریافت مقادیر مورد نیاز برای اتصال به پایگاه داده از متغیرهای محیطی
	user := os.Getenv("DB_USER") // کاربر پایگاه داده
	host := os.Getenv("DB_HOST") // هاست پایگاه داده
	port := os.Getenv("DB_PORT") // پورت پایگاه داده
	name := os.Getenv("DB_NAME") // نام پایگاه داده

	// چک کردن اینکه آیا مقادیر مهم برای اتصال به پایگاه داده وجود دارند یا خیر
	if user == "" || host == "" || port == "" || name == "" {
		log.Println("❌ Missing required DB environment variables.")
		return nil, errors.New("database connection failed") // در صورت نبود مقادیر، ارور می‌دهیم
	}

	// اتصال به پایگاه داده
	db := database.InitDB()
	if db == nil {
		return nil, errors.New("database connection failed") // در صورت عدم اتصال، ارور می‌دهیم
	}

	// راه‌اندازی Redis
	utils.InitRedis()        // راه‌اندازی Redis
	rdb := utils.RedisClient // ذخیره اتصال Redis در متغیر

	// راه‌اندازی صف برای ارسال و دریافت پیام‌ها
	queue.InitQueue() // صف مورد استفاده برای ناشران پیام‌ها

	// ایجاد مخزن‌های داده برای انجام عملیات بر روی موجودیت‌ها
	userRepo := repository.NewUserRepository(db)            // مخزن کاربران
	refreshRepo := repository.NewRefreshTokenRepository(db) // مخزن توکن‌های ریفرش
	authorRepo := repository.NewAuthorRepository(db)        // مخزن نویسندگان
	bookRepo := repository.NewBookRepository(db)            // مخزن کتاب‌ها
	loanRepo := repository.NewLoanRepository(db)            // مخزن وام‌ها

	// دریافت اتصال و کانال RabbitMQ (اگر در InitQueue تنظیم شده باشد)
	var rabbitConn *amqp.Connection
	var rabbitChan *amqp.Channel
	if queue.UsingRabbit() {
		client := queue.GetRabbitClient() // استفاده از getter جدید
		if client != nil {
			rabbitConn = client.Conn
			rabbitChan = client.Channel
		}
	}

	// ایجاد سرور Echo
	e := echo.New()             // ایجاد نمونه Echo
	e.Use(middleware.Recover()) // میانه‌رو برای بازیابی از خطاها
	e.Use(middleware.Logger())  // استفاده از میانه‌رو برای لاگ کردن درخواست‌ها

	// اضافه کردن مخازن داده به کانتکست درخواست‌ها
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("user_repo", userRepo)             // اضافه کردن مخزن کاربران به کانتکست
			c.Set("refresh_token_repo", refreshRepo) // اضافه کردن مخزن توکن‌های ریفرش به کانتکست
			c.Set("author_repo", authorRepo)         // اضافه کردن مخزن نویسندگان به کانتکست
			c.Set("book_repo", bookRepo)             // اضافه کردن مخزن کتاب‌ها به کانتکست
			c.Set("loan_repo", loanRepo)             // اضافه کردن مخزن وام‌ها به کانتکست
			return next(c)                           // ادامه اجرای درخواست
		}
	})

	// ثبت مسیرها (Routes) برای سرویس‌دهی به درخواست‌ها
	router.RegisterRoutes(e)

	// بازگشت شیء App برای استفاده در مراحل بعدی
	return &App{
		Server:      e,
		UserRepo:    userRepo,
		RefreshRepo: refreshRepo,
		AuthorRepo:  authorRepo,
		BookRepo:    bookRepo,
		LoanRepo:    loanRepo,
		RabbitConn:  rabbitConn,
		RabbitChan:  rabbitChan,
		Redis:       rdb,
	}, nil
}

// تابع main نقطه شروع برنامه است
func main() {
	// ساخت اپلیکیشن جدید
	app, err := NewApp()
	if err != nil {
		log.Fatal(err) // در صورت بروز خطا در ساخت اپلیکیشن، برنامه متوقف می‌شود
	}
	defer queue.CloseRabbitMQ() // بستن RabbitMQ در پایان

	// بستن کانال و اتصال RabbitMQ
	if app.RabbitChan != nil {
		_ = app.RabbitChan.Close()
	}
	if app.RabbitConn != nil {
		_ = app.RabbitConn.Close()
	}

	// دریافت پورت از متغیر محیطی
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080" // پورت پیش‌فرض
	}

	// تنظیمات سرور HTTP
	srv := &http.Server{
		Addr:         ":" + port,       // آدرس سرور
		Handler:      app.Server,       // هندلر درخواست‌ها
		ReadTimeout:  15 * time.Second, // زمان‌تایم اوت خواندن
		WriteTimeout: 30 * time.Second, // زمان‌تایم اوت نوشتن
		IdleTimeout:  60 * time.Second, // زمان‌تایم اوت زمانی که سرور بیکار است
	}

	// *** جدید: شروع consumerها در goroutineها برای پردازش صف‌ها ***
	ctx := context.Background()
	if queue.UsingRabbit() { // اگر RabbitMQ فعال باشد
		// شروع consumer برای فرمان‌های وام (loan_commands)
		go func() {
			if err := queue.StartLoanCommandConsumerRabbit(app.RabbitChan, app.LoanRepo, app.BookRepo); err != nil {
				log.Printf("خطا در شروع consumer فرمان‌های RabbitMQ: %v", err)
			}
		}()

		// شروع consumer برای رویدادهای وام (loan_events) با هندلر ساده (لاگ یا نوتیفیکیشن)
		go func() {
			amqpURL := os.Getenv("RABBITMQ_URL")
			if err := queue.StartRabbitConsumer(amqpURL, func(evt model.LoanEvent) {
				// هندلر ساده: لاگ کردن یا ارسال نوتیفیکیشن (مثلاً به کاربر)
				log.Printf("رویداد پردازش شد: نوع=%s, شناسه وام=%d, شناسه کاربر=%d, شناسه کتاب=%d", evt.EventType, evt.LoanID, evt.UserID, evt.BookID)
				// TODO: اضافه کردن نوتیفیکیشن واقعی، مثل "کتاب با موفقیت امانت گرفته شد" به کاربر (از طریق ایمیل یا WebSocket)
			}); err != nil {
				log.Printf("خطا در شروع consumer رویدادهای RabbitMQ: %v", err)
			}
		}()
	} else if app.Redis != nil { // fallback به Redis اگر RabbitMQ نباشد
		// شروع consumer برای فرمان‌های Redis
		go func() {
			if err := queue.StartLoanCommandConsumerRedis(ctx, app.Redis, app.LoanRepo, app.BookRepo); err != nil {
				log.Printf("خطا در شروع consumer فرمان‌های Redis: %v", err)
			}
		}()
		// برای رویدادها، می‌توانید consumer مشابه Redis اضافه کنید اگر لازم باشد
	}

	// شروع سرور در یک گوروتین
	go func() {
		log.Printf("🚀 Server listening on http://localhost:%s", port)
		if err := app.Server.StartServer(srv); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// منتظر سیگنال‌های خاتمه برای خاموش کردن سرور
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM) // شنیدن سیگنال‌های خاتمه
	<-quit                                             // توقف برنامه در صورت دریافت سیگنال

	log.Println("⏳ Shutting down...")

	// خاموش کردن سرور با زمان‌تایم اوت
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = app.Server.Shutdown(ctx)
}
