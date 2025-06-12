package handlers

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	models "github.com/iliyamo/go-learning/internal/model" // مدل‌های دیتابیس (User)
	"github.com/iliyamo/go-learning/internal/repository"   // ریپازیتوری برای کاربران
	"github.com/labstack/echo/v4"                          // فریم‌ورک HTTP
	"golang.org/x/crypto/bcrypt"                           // کتابخانه هش کردن رمز عبور
)

// AuthRequest ساختار دریافت اطلاعات از کلاینت برای ثبت‌نام و ورود
type AuthRequest struct {
	FullName string `json:"full_name"` // فقط در ثبت‌نام استفاده می‌شود
	Email    string `json:"email"`     // ایمیل کاربر
	Password string `json:"password"`  // رمز عبور کاربر
}

// jwtSecret کلید محرمانه برای امضای JWT (در محیط واقعی بهتره از .env بخونی)
var jwtSecret = []byte("your_secret_key")

// createToken یک JWT شامل اطلاعات کاربر (claims) تولید می‌کند
func createToken(user *models.User) (string, error) {
	// ساخت claims دلخواه
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role_id": user.RoleID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(), // تاریخ انقضا: ۲۴ ساعت بعد
	}

	// ساخت توکن با امضای HMAC
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// امضای نهایی توکن با کلید مخفی
	return token.SignedString(jwtSecret)
}

// Register هندلری برای ثبت‌نام کاربر جدید است
func Register(c echo.Context) error {
	// گرفتن ریپازیتوری کاربر از context
	db := c.Get("db").(*repository.UserRepository)

	// گرفتن داده‌های درخواست و bind به ساختار
	req := new(AuthRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	// هش کردن رمز عبور با bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to hash password"})
	}

	// ساخت ساختار کاربر برای ذخیره در دیتابیس
	user := &models.User{
		FullName:     req.FullName,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		RoleID:       2, // نقش پیش‌فرض (مثلاً: کاربر معمولی)
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// ذخیره کاربر در دیتابیس
	if err := db.CreateUser(user); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create user"})
	}

	// پاسخ موفق
	return c.JSON(http.StatusCreated, map[string]string{"message": "user registered successfully"})
}

// Login هندلری برای ورود کاربر و تولید JWT است
func Login(c echo.Context) error {
	// گرفتن ریپازیتوری از context
	db := c.Get("db").(*repository.UserRepository)

	// گرفتن داده‌های ورودی از کلاینت
	req := new(AuthRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	// جستجوی کاربر با ایمیل
	user, err := db.GetUserByEmail(req.Email)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "database error"})
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
	}

	// بررسی صحت رمز عبور وارد شده
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
	}

	// ایجاد توکن JWT در صورت اعتبارسنجی موفق
	token, err := createToken(user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create token"})
	}

	// ارسال توکن به عنوان پاسخ
	return c.JSON(http.StatusOK, map[string]string{"access_token": token})
}
