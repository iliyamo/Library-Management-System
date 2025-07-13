// internal/middleware/jwt.go
package middleware

import (
	"net/http" // برای ارسال پاسخ HTTP با وضعیت مناسب
	"strings"  // برای پردازش رشته هدر Authorization

	"github.com/iliyamo/go-learning/internal/utils" // توابع اعتبارسنجی و ساخت توکن JWT
	"github.com/labstack/echo/v4"                   // فریم‌ورک Echo برای هندل درخواست‌ها
)

// JWTAuth میدل‌ویری است برای محافظت از مسیر‌های نیازمند احراز هویت.
// این میدل‌ویر:
// 1. هدر Authorization را چک می‌کند
// 2. اگر وجود داشت و با "Bearer " شروع می‌شد، توکن را جدا می‌کند
// 3. توکن را با تابع ValidateToken اعتبارسنجی می‌کند
// 4. Claims (ادعاها) و userID را در context ذخیره می‌کند تا handler بتواند استفاده کند
func JWTAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// 1. خواندن هدر Authorization از درخواست
		authHeader := c.Request().Header.Get("Authorization")
		// اگر هدر خالی باشد یا با "Bearer " شروع نشود، پاسخ Unauthorized می‌دهیم
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "missing or invalid token",
			})
		}

		// 2. جدا کردن عبارت "Bearer " از ابتدای هدر
		// TrimSpace برای حذف فاصله‌های اضافی استفاده می‌شود
		tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer"))

		// 3. اعتبارسنجی توکن و بازگشت claims سفارشی
		claims, err := utils.ValidateToken(tokenString)
		if err != nil {
			// اگر اعتبارسنجی شکست خورد، دسترسی رد می‌شود
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "invalid or expired token",
			})
		}

		// 4. ذخیرهٔ claims و userID در context برای استفادهٔ بعدی
		// کلید "claims" شامل کل داده‌های توکن است
		c.Set("claims", claims)
		// کلید "userID" فقط شناسه کاربر را نگه می‌دارد (برای راحتی دسترسی)
		c.Set("userID", claims.UserID)

		// 5. اگر همه چیز اوکی باشد، به handler بعدی می‌رود
		return next(c)
	}
}
