package repository

import (
    "database/sql"
    "time"

    "github.com/iliyamo/go-learning/internal/model"
)

type LoanRepository struct {
	DB *sql.DB
}

func NewLoanRepository(db *sql.DB) *LoanRepository {
	return &LoanRepository{DB: db}
}

// CreateLoan ثبت یک امانت جدید در پایگاه داده.
// از آنجا که جدول loans دارای ستون‌های loan_date، due_date و return_date است،
// در اینجا از همان نام‌ها استفاده می‌کنیم و فیلد status را نیز درج می‌کنیم.
// مقدار return_date می‌تواند nil باشد که در این صورت به صورت NULL درج می‌شود.

func (r *LoanRepository) CreateLoan(loan *model.Loan) error {
    query := `INSERT INTO loans (user_id, book_id, loan_date, due_date, return_date, status)
        VALUES (?, ?, ?, ?, ?, ?)`
    // Execute the INSERT and capture the resulting sql.Result to obtain the
    // auto-generated primary key.  Without retrieving the LastInsertId, the
    // loan's ID would remain zero in the calling context.
    res, err := r.DB.Exec(query,
        loan.UserID,
        loan.BookID,
        loan.LoanDate,
        loan.DueDate,
        loan.ReturnDate,
        loan.Status,
    )
    if err != nil {
        return err
    }
    // Attempt to set the ID on the provided struct.  Not all drivers support
    // LastInsertId (e.g. Postgres), so ignore the error in that case.
    if id, err := res.LastInsertId(); err == nil {
        loan.ID = uint(id)
    }
    return nil
}


// GetLoansByUser دریافت لیست امانت‌های یک کاربر بر اساس شناسه کاربر
// نتایج بر اساس جدیدترین loan_date مرتب می‌شوند تا آخرین امانت‌ها در ابتدا نمایش داده شوند.
func (r *LoanRepository) GetLoansByUser(userID uint) ([]*model.Loan, error) {
    query := `SELECT id, user_id, book_id, loan_date, due_date, return_date, status
        FROM loans WHERE user_id = ? ORDER BY loan_date DESC`
    rows, err := r.DB.Query(query, userID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var loans []*model.Loan
    for rows.Next() {
        loan := &model.Loan{}
        var returnDate sql.NullTime
        if err := rows.Scan(&loan.ID, &loan.UserID, &loan.BookID, &loan.LoanDate, &loan.DueDate, &returnDate, &loan.Status); err != nil {
            return nil, err
        }
        if returnDate.Valid {
            rt := returnDate.Time
            loan.ReturnDate = &rt
        } else {
            loan.ReturnDate = nil
        }
        loans = append(loans, loan)
    }
    return loans, nil
}

// MarkAsReturned وضعیت امانت را به returned تغییر می‌دهد و تاریخ بازگشت را ثبت می‌کند.
// تنها در صورتی عملیات موفق است که وضعیت فعلی «borrowed» باشد. با این کار از بازگشت مکرر جلوگیری می‌شود.
func (r *LoanRepository) MarkAsReturned(loanID, userID uint) (bool, error) {
    query := `UPDATE loans SET status = ?, return_date = ? WHERE id = ? AND user_id = ? AND status = ?`
    res, err := r.DB.Exec(query, model.StatusReturned, time.Now(), loanID, userID, model.StatusBorrowed)
    if err != nil {
        return false, err
    }
    affected, _ := res.RowsAffected()
    return affected > 0, nil
}

// CheckActiveLoan بررسی می‌کند که آیا کاربر در حال حاضر این کتاب را به امانت دارد یا نه.
// یک امانت فعال یعنی هنوز return_date ثبت نشده و وضعیت برابر borrowed است.
func (r *LoanRepository) CheckActiveLoan(userID, bookID int) (bool, error) {
    query := `SELECT COUNT(*) FROM loans WHERE user_id = ? AND book_id = ? AND status = ? AND return_date IS NULL`
    var count int
    if err := r.DB.QueryRow(query, userID, bookID, model.StatusBorrowed).Scan(&count); err != nil {
        return false, err
    }
    return count > 0, nil
}

// ExistsPendingLoan بررسی وجود درخواست معلق برای یک کتاب از کاربر.
// در نسخه فعلی سیستم وضعیت pending وجود ندارد، بنابراین همواره مقدار false بازمی‌گرداند.
func (r *LoanRepository) ExistsPendingLoan(userID, bookID int) (bool, error) {
    return false, nil
}

// GetLoanByID بازیابی یک رکورد امانت بر اساس شناسه.
// مقادیر return_date ممکن است NULL باشد که در این صورت فیلد ReturnDate نال باقی می‌ماند.
func (r *LoanRepository) GetLoanByID(loanID int) (*model.Loan, error) {
    query := `SELECT id, user_id, book_id, loan_date, due_date, return_date, status FROM loans WHERE id = ?`
    row := r.DB.QueryRow(query, loanID)
    loan := &model.Loan{}
    var returnDate sql.NullTime
    if err := row.Scan(&loan.ID, &loan.UserID, &loan.BookID, &loan.LoanDate, &loan.DueDate, &returnDate, &loan.Status); err != nil {
        if err == sql.ErrNoRows {
            return nil, nil
        }
        return nil, err
    }
    if returnDate.Valid {
        rt := returnDate.Time
        loan.ReturnDate = &rt
    }
    return loan, nil
}

// UpdateStatus تغییر وضعیت یک وام (مثلاً توسط ادمین)
func (r *LoanRepository) UpdateStatus(loanID int, status string) error {
	query := `UPDATE loans SET status = ? WHERE id = ?`
	_, err := r.DB.Exec(query, status, loanID)
	return err
}
