package utils

import (
	"golang.org/x/crypto/bcrypt" // bcrypt برای تولید هش امن همراه با salt و تنظیم difficulty
)

// HashPassword پسورد متنی ساده را می‌گیرد و آن را با الگوریتم bcrypt هش می‌کند.
// bcrypt به‌طور خودکار salt تصادفی اضافه می‌کند که امنیت را افزایش می‌دهد.
// نتیجهٔ این تابع رشتهٔ هش شده است که باید در دیتابیس ذخیره شود.
func HashPassword(password string) (string, error) {
	// GenerateFromPassword هش را تولید می‌کند.
	// DefaultCost یک مقدار cost پیش‌فرض و متعادل برای امنیت مناسب است.
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		// اگر خطایی رخ دهد، رشتهٔ خالی و خطا برگردانده می‌شود.
		return "", err
	}
	// هش تولیدشده را به صورت string برمی‌گردانیم.
	return string(hashedBytes), nil
}

// CheckPasswordHash صحت پسورد ورودی را با هش ذخیره‌شده مقایسه می‌کند.
// ابتدا هش ذخیره‌شده دیکد می‌شود و سپس با پسورد ورودی تطبیق پیدا می‌کند.
// اگر مقایسه موفق باشد (پسورد درست باشد)، خروجی true خواهد بود.
func CheckPasswordHash(password, hash string) bool {
	// CompareHashAndPassword هش ذخیره و پسورد را مقایسه می‌کند.
	// در صورت تطابق، err برابر nil خواهد بود.
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	// اگر err nil باشد، پسورد ورودی صحیح است.
	return err == nil
}
