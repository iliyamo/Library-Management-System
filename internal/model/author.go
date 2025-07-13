package model

import "time"

// Author ساختار مربوط به نویسنده‌ها در سیستم
type Author struct {
	ID        int       `json:"id"`         // شناسه یکتا
	Name      string    `json:"name"`       // نام نویسنده
	Biography string    `json:"biography"`  // زندگینامه نویسنده
	BirthDate time.Time `json:"birth_date"` // تاریخ تولد نویسنده
	CreatedAt time.Time `json:"created_at"` // زمان ایجاد رکورد
	UpdatedAt time.Time `json:"updated_at"` // زمان آخرین به‌روزرسانی
}
