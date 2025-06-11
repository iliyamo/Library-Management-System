package handlers

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	models "github.com/iliyamo/go-learning/internal/model" // توجه: اینجا models هست نه model
	"github.com/iliyamo/go-learning/internal/repository"   // اطمینان از import صحیح
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

// ساختار داده ورودی ثبت نام و ورود
type AuthRequest struct {
	FullName string `json:"full_name"` // فقط برای ثبت نام لازم است
	Email    string `json:"email"`
	Password string `json:"password"`
}

// کلید مخفی برای امضای JWT
var jwtSecret = []byte("your_secret_key")

// ایجاد توکن JWT با ادعای (claims) دلخواه
func createToken(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role_id": user.RoleID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// هندلر ثبت نام
func Register(c echo.Context) error {
	db := c.Get("db").(*repository.UserRepository)

	req := new(AuthRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	// هش کردن پسورد
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to hash password"})
	}

	user := &models.User{
		FullName:     req.FullName,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		RoleID:       2, // نقش پیش فرض
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := db.CreateUser(user); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create user"})
	}

	return c.JSON(http.StatusCreated, map[string]string{"message": "user registered successfully"})
}

// هندلر ورود
func Login(c echo.Context) error {
	db := c.Get("db").(*repository.UserRepository)

	req := new(AuthRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	user, err := db.GetUserByEmail(req.Email)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "database error"})
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
	}

	token, err := createToken(user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create token"})
	}

	return c.JSON(http.StatusOK, map[string]string{"access_token": token})
}
