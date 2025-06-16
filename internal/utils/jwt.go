package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ✅ ساختار Claims اختصاصی ما برای JWT
type JWTClaims struct {
	UserID               uint   `json:"user_id"` // شناسه کاربر
	Email                string `json:"email"`   // ایمیل کاربر
	RoleID               uint   `json:"role_id"` // نقش کاربر
	jwt.RegisteredClaims        // شامل exp, iat و ...
}

// 🔐 کلید مخفی برای امضای JWT (پیشنهاد: از ENV بخون)
var jwtSecretKey = []byte("your_secret_key")

// ✅ GenerateAccessToken → تولید توکن دسترسی کوتاه‌مدت (مثلاً 15 دقیقه‌ای)
func GenerateAccessToken(userID uint, email string, roleID uint) (string, error) {
	claims := JWTClaims{
		UserID: userID,
		Email:  email,
		RoleID: roleID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecretKey)
}

// ✅ GenerateRefreshToken → تولید توکن بلندمدت برای تمدید (مثلاً 7 روزه)
func GenerateRefreshToken(userID uint, email string, roleID uint) (string, error) {
	claims := JWTClaims{
		UserID: userID,
		Email:  email,
		RoleID: roleID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecretKey)
}

// ✅ ValidateToken → اعتبارسنجی JWT و استخراج claims
func ValidateToken(tokenStr string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecretKey, nil
	})
	if err != nil {
		return nil, err
	}

	// بررسی معتبر بودن token و تبدیل به claims سفارشی خودمون
	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid or expired token")
	}
	return claims, nil
}
