package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/iliyamo/go-learning/internal/model"
	"github.com/iliyamo/go-learning/internal/repository"
	"github.com/iliyamo/go-learning/internal/utils"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

// AuthRequest ساختار داده‌های ثبت‌نام و ورود
type AuthRequest struct {
	FullName string `json:"full_name"` // فقط برای ثبت‌نام استفاده می‌شود
	Email    string `json:"email"`
	Password string `json:"password"`
}

// کلید مخفی JWT (پیشنهاد: از env بخون)
var jwtSecret = []byte("your_secret_key")

// ایجاد یک JWT ساده برای توکن
func createToken(user *model.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role_id": user.RoleID,
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ✅ Register: ثبت‌نام کاربر
func Register(c echo.Context) error {
	db := c.Get("db").(*repository.UserRepository)
	req := new(AuthRequest)

	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to hash password"})
	}

	user := &model.User{
		FullName:     req.FullName,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		RoleID:       2,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := db.CreateUser(user); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create user"})
	}

	return c.JSON(http.StatusCreated, map[string]string{"message": "user registered successfully"})
}

// ✅ Login: ورود کاربر و تولید access + refresh token
func Login(c echo.Context) error {
	userRepo := c.Get("db").(*repository.UserRepository)
	req := new(AuthRequest)

	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	user, err := userRepo.GetUserByEmail(req.Email)
	if err != nil || user == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
	}

	// Access Token
	accessToken, err := utils.GenerateAccessToken(uint(user.ID), user.Email, uint(user.RoleID))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to generate access token"})
	}

	// Refresh Token
	refreshToken, err := utils.GenerateRefreshToken(uint(user.ID), user.Email, uint(user.RoleID))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to generate refresh token"})
	}

	// ذخیره در پایگاه‌داده
	refreshRepo := repository.NewRefreshTokenRepository(userRepo.DB)
	if err := refreshRepo.Store(refreshToken, user.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to store refresh token"})
	}

	// خروجی به کلاینت
	return c.JSON(http.StatusOK, map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

// ✅ Profile: دریافت اطلاعات کاربر با JWT
// Profile handler اطلاعات کامل کاربر را برمی‌گرداند
func Profile(c echo.Context) error {
	// مرحله 1: گرفتن هدر Authorization
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Missing token"})
	}

	// مرحله 2: حذف "Bearer " از ابتدا
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// مرحله 3: اعتبارسنجی توکن
	claims, err := utils.ValidateToken(tokenString)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid token"})
	}

	// مرحله 4: گرفتن user_id از claims
	userID := int(claims.UserID)

	// مرحله 5: دسترسی به repo از context
	repo := c.Get("db").(*repository.UserRepository)

	// مرحله 6: گرفتن کاربر از دیتابیس
	user, err := repo.GetUserByID(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Database error"})
	}
	if user == nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "User not found"})
	}

	// مرحله 7: بازگرداندن اطلاعات کامل
	return c.JSON(http.StatusOK, user)
}

// ✅ Logout: حذف refresh token از دیتابیس (بی‌اثر کردن آن)
func Logout(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing or invalid token"})
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := utils.ValidateToken(tokenString)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid or expired token"})
	}

	userRepo := c.Get("db").(*repository.UserRepository)
	refreshRepo := repository.NewRefreshTokenRepository(userRepo.DB)

	if err := refreshRepo.DeleteAll(claims.UserID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to logout"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "logged out successfully"})
}
