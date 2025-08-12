// internal/handler/loan.go

package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/iliyamo/Library-Management-System/internal/model"
	"github.com/iliyamo/Library-Management-System/internal/queue"
	"github.com/iliyamo/Library-Management-System/internal/repository"
	"github.com/iliyamo/Library-Management-System/internal/utils"
)

// این نسخه از هندلر، منطق ویرایش DB را به مصرف‌کننده‌های صف منتقل می‌کند.
// نتیجه: API فقط فرمان را به صف loan_commands می‌فرستد و 202 برمی‌گرداند.

// RequestLoan: درخواست امانت را صف می‌کند (CmdBorrow)
func RequestLoan(c echo.Context) error {
	var req model.LoanRequest
	if err := c.Bind(&req); err != nil || req.BookID == 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "درخواست نامعتبر"})
	}

	claims, ok := c.Get("claims").(*utils.JWTClaims)
	if !ok || claims == nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "توکن نامعتبر"})
	}

	// مقدار پیش‌فرض روزها (اگر ارسال نشده یا <=0)
	if req.Days <= 0 {
		req.Days = 7
	}

	// ساخت فرمان و ارسال به صف loan_commands
	cmd := model.LoanCommand{
		Type: model.CmdBorrow,
	}
	cmd.Payload.UserID = claims.UserID
	cmd.Payload.BookID = req.BookID
	cmd.Payload.Days = req.Days

	if err := queue.PublishLoanCommand(cmd); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "خطا در ارسال فرمان"})
	}

	return c.JSON(http.StatusAccepted, echo.Map{
		"status":  "queued",
		"message": "درخواست امانت در صف پردازش قرار گرفت",
	})
}

// ViewMyLoans: نمایش لیست امانت‌های کاربر (همچنان مستقیم از مخزن)
func ViewMyLoans(c echo.Context) error {
	claims, ok := c.Get("claims").(*utils.JWTClaims)
	if !ok || claims == nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "توکن نامعتبر"})
	}
	userID := claims.UserID

	loanRepo := c.Get("loan_repo").(*repository.LoanRepository)
	loans, err := loanRepo.GetLoansByUser(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "خطا در واکشی امانت‌ها"})
	}

	res := make([]map[string]interface{}, len(loans))
	for i, loan := range loans {
		m := map[string]interface{}{
			"id":        loan.ID,
			"user_id":   loan.UserID,
			"book_id":   loan.BookID,
			"loan_date": loan.LoanDate.Format("2006-01-02"),
			"due_date":  loan.DueDate.Format("2006-01-02"),
			"status":    loan.Status,
		}
		if loan.ReturnDate != nil {
			m["return_date"] = loan.ReturnDate.Format("2006-01-02")
		}
		res[i] = m
	}
	return c.JSON(http.StatusOK, res)
}

// ReturnBook: بازگرداندن کتاب را صف می‌کند (CmdReturn)
func ReturnBook(c echo.Context) error {
	idStr := c.Param("id")
	loanID, err := strconv.Atoi(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "شناسه نامعتبر است"})
	}

	claims, ok := c.Get("claims").(*utils.JWTClaims)
	if !ok || claims == nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "توکن نامعتبر"})
	}

	cmd := model.LoanCommand{Type: model.CmdReturn}
	cmd.Payload.UserID = claims.UserID
	cmd.Payload.LoanID = uint(loanID)

	if err := queue.PublishLoanCommand(cmd); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "خطا در ارسال فرمان"})
	}

	return c.JSON(http.StatusAccepted, echo.Map{
		"status":  "queued",
		"message": "درخواست بازگرداندن کتاب در صف پردازش قرار گرفت",
	})
}
