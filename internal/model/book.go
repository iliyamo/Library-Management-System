// internal/model/book.go
package model

import "time"

// Book نمایانگر رکوردی در جدول books است.
type Book struct {
	ID              int       `json:"id"`               // شناسه یکتا
	Title           string    `json:"title"`            // عنوان
	ISBN            string    `json:"isbn"`             // شماره استاندارد کتاب (باید یکتا باشد)
	AuthorID        int       `json:"author_id"`        // FK به authors
	CategoryID      *int      `json:"category_id"`      // FK به categories (می‌تواند NULL باشد)
	Description     *string   `json:"description"`      // توضیحات (NULLable)
	PublishedYear   *int      `json:"published_year"`   // سال انتشار (NULLable)
	TotalCopies     int       `json:"total_copies"`     // کل نسخه‌ها
	AvailableCopies int       `json:"available_copies"` // نسخه‌های در دسترس
	CreatedAt       time.Time `json:"created_at"`       // زمان ایجاد
}

// BookSearchParams ➜ پارامترهای لازم برای جستجو با cursor-based pagination
type BookSearchParams struct {
	Query    string `query:"query"`     // عبارت جستجو (برای full-text search)
	CursorID int    `query:"cursor_id"` // آخرین ID مشاهده‌شده
	Limit    int    `query:"limit"`     // تعداد نتایج در هر صفحه (مثلاً 10)
}
