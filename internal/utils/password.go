package utils

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword یک پسورد متنی ساده رو گرفته و با استفاده از الگوریتم bcrypt
// اون رو هش می‌کنه تا به‌صورت امن در دیتابیس ذخیره بشه.
// bcrypt.DefaultCost مقدار پیش‌فرض سختی هش‌کردن رو تعیین می‌کنه.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash وظیفه داره یک پسورد ساده (که یوزر وارد می‌کنه)
// رو با هش ذخیره‌شده در دیتابیس مقایسه کنه.
// اگر مقایسه موفق باشه (یعنی پسورد درست باشه)، خروجی true می‌ده.
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
