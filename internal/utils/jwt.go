package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ساختار اطلاعاتی که می‌خوایم در توکن ذخیره کنیم
type JWTClaims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	RoleID uint   `json:"role_id"`
	jwt.RegisteredClaims
}

// secret key برای امضای JWT
var jwtSecretKey = []byte("your-secret-key") // 🔐 حتماً در env نگه‌دار

// GenerateAccessToken توکن JWT کوتاه‌مدت (مثلاً ۱۵ دقیقه‌ای) برای احراز هویت تولید می‌کنه
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

// GenerateRefreshToken توکن بلندمدت (مثلاً ۷ روزه) برای تمدید توکن اصلی تولید می‌کنه
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

// ValidateToken بررسی می‌کنه که آیا توکن معتبر هست یا نه
func ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecretKey, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
