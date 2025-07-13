package utils

import (
	"errors" // برای نمایش خطا در اعتبارسنجی توکن
	"time"   // برای زمان‌بندی انقضای توکن

	"github.com/golang-jwt/jwt/v5" // پکیج JWT برای ساخت و اعتبارسنجی توکن
)

// JWTClaims ساختار داده‌ای است که درون توکن ذخیره می‌شود:
// - UserID: شناسهٔ کاربر
// - Email: ایمیل کاربر
// - RoleID: نقش کاربر (مثلاً admin یا member)
// همچنین با استفاده از RegisteredClaims قادر به تنظیم exp و iat هستیم.
type JWTClaims struct {
	UserID               uint   `json:"user_id"` // شناسهٔ کاربر
	Email                string `json:"email"`   // ایمیل کاربر
	RoleID               uint   `json:"role_id"` // نقش کاربر
	jwt.RegisteredClaims        // فیلدهای استاندارد مثل exp, iat
}

// jwtSecretKey کلیدی است که با آن توکن‌ها را امضا می‌کنیم.
// پیشنهاد می‌شود این مقدار را در متغیر محیطی نگه دارید.
var jwtSecretKey = []byte("your_secret_key")

// GenerateAccessToken یک توکن کوتاه‌مدت می‌سازد که معمولاً ۱۵ دقیقه اعتبار دارد.
// این توکن برای دسترسی به مسیرهای محافظت‌شده استفاده می‌شود.
func GenerateAccessToken(userID uint, email string, roleID uint) (string, error) {
	// تنظیم ادعاها (Claims) در توکن
	claims := JWTClaims{
		UserID: userID,
		Email:  email,
		RoleID: roleID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)), // انقضا پس از 15 دقیقه
			IssuedAt:  jwt.NewNumericDate(time.Now()),                       // زمان صدور توکن
		},
	}
	// ساخت توکن با الگوریتم HS256 و ادعاهای تنظیم‌شده
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// امضای توکن با کلید مخفی
	return token.SignedString(jwtSecretKey)
}

// GenerateRefreshToken یک توکن بلندمدت (مثلاً ۷ روزه) می‌سازد
// برای درخواست توکن دسترسی جدید (Access Token) استفاده می‌شود.
func GenerateRefreshToken(userID uint, email string, roleID uint) (string, error) {
	claims := JWTClaims{
		UserID: userID,
		Email:  email,
		RoleID: roleID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)), // انقضا پس از 7 روز
			IssuedAt:  jwt.NewNumericDate(time.Now()),                         // زمان صدور توکن
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecretKey)
}

// ValidateToken توکنی را که ارسال شده بررسی می‌کند:
// 1. امضا معتبر باشد
// 2. توکن منقضی نشده باشد
// سپس ادعاها (Claims) را برمی‌گرداند.
func ValidateToken(tokenStr string) (*JWTClaims, error) {
	// پارس توکن با ساختار JWTClaims
	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecretKey, nil // بازگرداندن کلید برای بررسی امضا
	})
	if err != nil {
		// در صورت خطا در پارس یا امضا
		return nil, err
	}

	// تبدیل claims به نوع سفارشی JWTClaims
	claims, ok := token.Claims.(*JWTClaims)
	// بررسی صحت و معتبر بودن توکن
	if !ok || !token.Valid {
		return nil, errors.New("invalid or expired token")
	}
	return claims, nil
}
