// internal/utils/jwt.go
package utils

import (
	"errors"
	"log"
	"os"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims ساختار اطلاعاتی برای ذخیره اطلاعات در توکن JWT است.
// این ساختار شامل شناسه کاربر، ایمیل، شناسه نقش کاربر و اطلاعات ثبت‌شده است.
type JWTClaims struct {
	UserID               uint   `json:"user_id"` // شناسه کاربر
	Email                string `json:"email"`   // ایمیل کاربر
	RoleID               uint   `json:"role_id"` // شناسه نقش کاربر
	jwt.RegisteredClaims        // اطلاعات ثبت‌شده مثل تاریخ انقضا و تاریخ صدور توکن
}

var (
	jwtSecretKey   []byte    // کلید سری برای امضای توکن JWT
	loadSecretOnce sync.Once // برای بارگذاری تنها یکبار کلید سری
)

// تابعی برای بارگذاری کلید سری از متغیر محیطی
func loadSecret() {
	secret := os.Getenv("JWT_SECRET") // دریافت کلید سری از متغیر محیطی
	if secret == "" {
		secret = "change_me_dev_only" // اگر متغیر محیطی وجود نداشت، از مقدار پیش‌فرض استفاده می‌کنیم
		log.Println("⚠️  JWT_SECRET در محیط تنظیم نشده است. از مقدار پیش‌فرض توسعه استفاده می‌شود.")
	}
	jwtSecretKey = []byte(secret) // تبدیل کلید سری به نوع بایت
}

// تابعی برای دریافت کلید سری با استفاده از sync.Once تا تنها یکبار بارگذاری شود
func getSecret() []byte {
	loadSecretOnce.Do(loadSecret) // تنها یکبار بارگذاری کلید سری
	return jwtSecretKey
}

// تابعی برای تولید توکن دسترسی (Access Token)
// این توکن برای دسترسی به منابع محافظت‌شده استفاده می‌شود و به مدت 15 دقیقه معتبر است
func GenerateAccessToken(userID uint, email string, roleID uint) (string, error) {
	// ساخت claims یا اطلاعات توکن
	claims := JWTClaims{
		UserID: userID,
		Email:  email,
		RoleID: roleID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)), // تاریخ انقضای توکن 15 دقیقه بعد از صدور
			IssuedAt:  jwt.NewNumericDate(time.Now()),                       // تاریخ صدور توکن
		},
	}
	// ایجاد توکن با استفاده از اطلاعات claims و امضای آن
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getSecret()) // توکن امضا شده را برمی‌گرداند
}

// تابعی برای تولید توکن refresh (تازه‌سازی)
// این توکن برای دریافت یک توکن جدید پس از انقضای توکن دسترسی استفاده می‌شود و به مدت 7 روز معتبر است
func GenerateRefreshToken(userID uint, email string, roleID uint) (string, error) {
	// ساخت claims یا اطلاعات توکن
	claims := JWTClaims{
		UserID: userID,
		Email:  email,
		RoleID: roleID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)), // تاریخ انقضای توکن 7 روز بعد از صدور
			IssuedAt:  jwt.NewNumericDate(time.Now()),                         // تاریخ صدور توکن
		},
	}
	// ایجاد توکن با استفاده از اطلاعات claims و امضای آن
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getSecret()) // توکن امضا شده را برمی‌گرداند
}

// تابعی برای اعتبارسنجی توکن JWT
// این تابع توکن را تجزیه می‌کند و بررسی می‌کند که آیا معتبر است یا خیر
func ValidateToken(tokenStr string) (*JWTClaims, error) {
	// تجزیه توکن و بررسی صحت آن
	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return getSecret(), nil // استفاده از کلید سری برای اعتبارسنجی توکن
	})
	if err != nil {
		return nil, err // در صورت بروز خطا توکن معتبر نیست
	}
	claims, ok := token.Claims.(*JWTClaims) // استخراج اطلاعات claims از توکن
	if !ok || !token.Valid {                // بررسی اعتبار توکن
		return nil, errors.New("invalid or expired token") // توکن نامعتبر یا منقضی شده است
	}
	return claims, nil // برگرداندن اطلاعات claims در صورت معتبر بودن توکن
}

// تابعی برای استخراج شناسه کاربر از توکن JWT
// این تابع شناسه کاربر را از اطلاعات claims استخراج می‌کند

