// internal/handler/auth.go
package handler

import (
	"net/http" // HTTP status codes and response handling
	"time"     // برای تنظیم تاریخ ایجاد و بروزرسانی

	"github.com/labstack/echo/v4" // وب فریم‌ورک Ech
	"golang.org/x/crypto/bcrypt"  // bcrypt برای هش‌کردن و بررسی رمز عبور

	"github.com/iliyamo/go-learning/internal/model"      // مدل‌های داده‌ای (User)
	"github.com/iliyamo/go-learning/internal/repository" // دسترسی به داده (UserRepo, RefreshTokenRepo)
	"github.com/iliyamo/go-learning/internal/utils"      // توليد و اعتبارسنجی JWT
)

// AuthRequest ساختار داده‌ای ورودی برای ثبت‌نام و ورود
type AuthRequest struct {
	FullName string `json:"full_name"` // فقط برای ثبت‌نام لازم است
	Email    string `json:"email"`     // ایمیل یکتا کاربر
	Password string `json:"password"`  // رمز عبور ساده‌ی کاربر
}

// Register کاربر جدید را ثبت می‌کند
func Register(c echo.Context) error {
	// 1. دریافت و تبدیل بدنه‌ی JSON به AuthRequest
	req := new(AuthRequest)
	if err := c.Bind(req); err != nil {
		// اگر JSON نامعتبر باشد، کد 400 ارسال می‌کنیم
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}

	// 2. هش‌کردن رمز عبور با bcrypt
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		// وقتی هش‌کردن با خطا مواجه شود، کد 500 ارسال می‌کنیم
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to hash password"})
	}

	// 3. ساخت شیء کاربر با مقادیر دریافتی و زمان‌های فعلی
	user := &model.User{
		FullName:     req.FullName,
		Email:        req.Email,
		PasswordHash: string(hashed),
		RoleID:       2, // نقش پیش‌فرض (مثلاً member)
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// 4. ذخیره‌ی کاربر در دیتابیس از طریق UserRepository
	userRepo := c.Get("user_repo").(*repository.UserRepository)
	if err := userRepo.CreateUser(user); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to create user"})
	}

	// 5. موفقیت ثبت‌نام، کد 201 ارسال شود
	return c.JSON(http.StatusCreated, echo.Map{"message": "user registered successfully"})
}

// Login اعتبارسنجی کاربر و تولید JWT
func Login(c echo.Context) error {
	// 1. دریافت و تبدیل ورودی
	req := new(AuthRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}

	// 2. واکشی کاربر با ایمیل
	userRepo := c.Get("user_repo").(*repository.UserRepository)
	user, err := userRepo.GetUserByEmail(req.Email)
	if err != nil || user == nil {
		// اگر کاربر یافت نشد یا خطا رخ داد، اعتبارسنجی ناموفق است
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid credentials"})
	}

	// 3. مقایسه‌ی رمز ورود با هش ذخیره‌شده
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid credentials"})
	}

	// 4. تولید Access Token کوتاه‌مدت
	accessToken, err := utils.GenerateAccessToken(uint(user.ID), user.Email, uint(user.RoleID))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to generate access token"})
	}

	// 5. تولید Refresh Token طولانی‌مدت
	refreshToken, err := utils.GenerateRefreshToken(uint(user.ID), user.Email, uint(user.RoleID))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to generate refresh token"})
	}

	// 6. ذخیره‌ی Refresh Token در دیتابیس (خطا نادیده گرفته می‌شود)
	refreshRepo := c.Get("refresh_token_repo").(*repository.RefreshTokenRepository)
	_ = refreshRepo.Store(refreshToken, user.ID)

	// 7. بازگشت توکن‌ها به کلاینت
	return c.JSON(http.StatusOK, echo.Map{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

// Profile اطلاعات کاربر جاری را بازمی‌گرداند (JWTAuth middleware الزامی است)
func Profile(c echo.Context) error {
	// 1. دریافت Claims از context که توسط middleware قرار داده شده
	claims := c.Get("claims").(*utils.JWTClaims)
	userID := int(claims.UserID) // تبدیل شناسه به int

	// 2. واکشی اطلاعات کاربری از دیتابیس
	userRepo := c.Get("user_repo").(*repository.UserRepository)
	user, err := userRepo.GetUserByID(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "database error"})
	}
	if user == nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "user not found"})
	}

	// 3. بازگشت اطلاعات کاربر
	return c.JSON(http.StatusOK, user)
}

// Logout حذف همه‌ی Refresh Token‌های کاربر فعلی (نیازی به خطا دادن نیست)
func Logout(c echo.Context) error {
	// 1. دریافت Claims از context
	claims := c.Get("claims").(*utils.JWTClaims)

	// 2. حذف توکن‌ها از دیتابیس
	refreshRepo := c.Get("refresh_token_repo").(*repository.RefreshTokenRepository)
	_ = refreshRepo.DeleteAll(claims.UserID) // خطا نادیده گرفته می‌شود

	// 3. همیشه پیام موفقیت‌آمیز بازگردانده شود
	return c.JSON(http.StatusOK, echo.Map{"message": "logged out successfully"})
}
