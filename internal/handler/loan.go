package handler

import (
    "net/http"
    "strconv"
    "time"

    "github.com/labstack/echo/v4"

    "github.com/iliyamo/Library-Management-System/internal/model"
    "github.com/iliyamo/Library-Management-System/internal/queue"
    "github.com/iliyamo/Library-Management-System/internal/repository"
    "github.com/iliyamo/Library-Management-System/internal/utils"
)

// loanOpCh acts as a binary semaphore implemented via a buffered channel.
// It ensures that concurrent borrow/return operations are serialized to avoid race conditions.
var loanOpCh = make(chan struct{}, 1)

// RequestLoan handles creation of a new loan request for a book. It expects a
// JSON payload with a single field `book_id`. The currently authenticated user
// is inferred from the JWT claims stored on the context by the JWTAuth
// middleware. If the requested book does not exist, is unavailable or the user
// already has an active loan for it, an appropriate error response is
// returned. Otherwise a new loan record is created, the book's available
// copies are decremented and a loan event is published asynchronously.
func RequestLoan(c echo.Context) error {
    // Bind incoming JSON to a LoanRequest.  The LoanRequest struct
    // definition lives in the model package and only contains BookID.
    var req model.LoanRequest
    if err := c.Bind(&req); err != nil || req.BookID == 0 {
        return c.JSON(http.StatusBadRequest, echo.Map{"error": "درخواست نامعتبر"})
    }

    // Extract the authenticated user's ID from the JWT claims stored on
    // the context.  The claims are set by the JWTAuth middleware.
    claims, ok := c.Get("claims").(*utils.JWTClaims)
    if !ok || claims == nil {
        return c.JSON(http.StatusUnauthorized, echo.Map{"error": "توکن نامعتبر"})
    }
    userID := claims.UserID

    // Acquire repositories from the context.  These were registered in
    // main.go when setting up the Echo server.  They provide access to
    // database operations for books and loans.
    loanRepo := c.Get("loan_repo").(*repository.LoanRepository)
    bookRepo := c.Get("book_repo").(*repository.BookRepository)

    // Verify that the requested book exists.
    book, err := bookRepo.GetBookByID(int(req.BookID))
    if err != nil {
        return c.JSON(http.StatusInternalServerError, echo.Map{"error": "خطا در واکشی کتاب"})
    }
    if book == nil {
        return c.JSON(http.StatusNotFound, echo.Map{"error": "کتاب یافت نشد"})
    }

    // Ensure there is at least one available copy to borrow.
    if book.AvailableCopies < 1 {
        return c.JSON(http.StatusNotFound, echo.Map{"error": "هیچ نسخه‌ای از کتاب موجود نیست"})
    }

    // Check whether the user already has an active (unreturned) loan for this
    // book.  We cast the unsigned ID values to int for compatibility with
    // repository signatures.
    hasActive, err := loanRepo.CheckActiveLoan(int(userID), int(req.BookID))
    if err != nil {
        return c.JSON(http.StatusInternalServerError, echo.Map{"error": "خطا در بررسی امانت"})
    }
    if hasActive {
        return c.JSON(http.StatusConflict, echo.Map{"error": "شما در حال حاضر این کتاب را به امانت دارید"})
    }

    // Acquire the semaphore to ensure exclusive access for this operation
    loanOpCh <- struct{}{}
    defer func() { <-loanOpCh }()

    // Build the loan record with loan_date set to now and due_date set to
    // seven days from now.  return_date is nil until the book is returned.
    now := time.Now()
    // تعیین مدت زمان امانت. اگر کاربر در بدنهٔ درخواست days را مشخص کرده باشد
    // از آن استفاده می‌شود، در غیر این صورت مقدار پیش‌فرض ۷ روز در نظر گرفته می‌شود.
    days := req.Days
    if days <= 0 {
        days = 7
    }
    due := now.Add(time.Duration(days) * 24 * time.Hour)
    loan := &model.Loan{
        UserID:     userID,
        BookID:     req.BookID,
        LoanDate:   now,
        DueDate:    due,
        Status:     model.StatusBorrowed,
        ReturnDate: nil,
    }
    // Insert the loan into the database.  If it fails, return a 500 status.
    if err := loanRepo.CreateLoan(loan); err != nil {
        return c.JSON(http.StatusInternalServerError, echo.Map{"error": "خطا در ثبت امانت"})
    }

    // Decrement the available copies for the book.  We ignore errors from
    // updating the book; even if this update fails, the loan has been
    // recorded.  In a real application you might roll back the loan on
    // failure.
    book.AvailableCopies--
    _, _ = bookRepo.UpdateBook(book)

    
// Publish a LoanRequested event to the message broker.  Include the
// remaining copies and due date so that the consumer can generate
// detailed logs.  Any error is silently ignored to avoid impacting
// the API response.
event := model.LoanEvent{
    EventType:       model.LoanRequested,
    LoanID:          loan.ID,
    UserID:          userID,
    BookID:          req.BookID,
    Time:            time.Now(),
    RemainingCopies: int(book.AvailableCopies),
    DueDate:         due,
}
_ = queue.PublishEvent(event)

// Build a response map with simplified date strings instead of RFC3339.
resp := map[string]interface{}{
    "id":        loan.ID,
    "user_id":   loan.UserID,
    "book_id":   loan.BookID,
    "loan_date": loan.LoanDate.Format("2006-01-02"),
    "due_date":  due.Format("2006-01-02"),
    "status":    loan.Status,
}
if loan.ReturnDate != nil {
    resp["return_date"] = loan.ReturnDate.Format("2006-01-02")
}
// Return the created loan record with simplified dates.
return c.JSON(http.StatusCreated, resp)
}

// ViewMyLoans returns a list of all loans for the currently authenticated user.
// It does not perform any filtering and simply forwards the result from
// LoanRepository.GetLoansByUser.  The user ID is extracted from the JWT
// claims stored on the context.  If the repository returns an error, a 500
// status is returned; otherwise the slice of loans (which may be empty) is
// returned with HTTP 200.
func ViewMyLoans(c echo.Context) error {
    // Get the user ID from the JWT claims.
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
    
// Always return 200 with the list of loans, even if empty.
// Transform the loans into a slice of maps with simplified date strings.
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

// ReturnBook marks a loan as returned and increments the book's available
// copies.  The loan ID is taken from the URL parameter.  Only the user who
// created the loan may return it; if the loan is not found or does not
// belong to the user, a 404 is returned.  If the loan has already been
// returned, a 400 is returned.  On success the status is updated, the book
// inventory is adjusted and a LoanReturned event is published.
func ReturnBook(c echo.Context) error {
    // Parse loan ID from the path parameter.  If it's not a valid integer
    // return a bad request error.
    idStr := c.Param("id")
    loanID, err := strconv.Atoi(idStr)
    if err != nil {
        return c.JSON(http.StatusBadRequest, echo.Map{"error": "شناسه نامعتبر است"})
    }

    // Retrieve the authenticated user's ID.
    claims, ok := c.Get("claims").(*utils.JWTClaims)
    if !ok || claims == nil {
        return c.JSON(http.StatusUnauthorized, echo.Map{"error": "توکن نامعتبر"})
    }
    userID := claims.UserID

    // Load repositories from context.
    loanRepo := c.Get("loan_repo").(*repository.LoanRepository)
    bookRepo := c.Get("book_repo").(*repository.BookRepository)

    // Fetch the loan record to ensure it exists and belongs to this user.
    loan, err := loanRepo.GetLoanByID(loanID)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, echo.Map{"error": "خطا در واکشی امانت"})
    }
    if loan == nil || loan.UserID != userID {
        return c.JSON(http.StatusNotFound, echo.Map{"error": "امانت یافت نشد"})
    }
    // Acquire the semaphore to ensure exclusive access for this return operation
    loanOpCh <- struct{}{}
    defer func() { <-loanOpCh }()

    // Only borrowed loans may be returned.
    if loan.Status != model.StatusBorrowed {
        return c.JSON(http.StatusBadRequest, echo.Map{"error": "این امانت قبلاً بازگردانده شده است"})
    }

    // Mark the loan as returned in the database.  The repository method
    // returns false if no row was updated (e.g. because status was not
    // 'borrowed' or the ID/user mismatch).  We treat that as not found.
    updated, err := loanRepo.MarkAsReturned(uint(loanID), userID)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, echo.Map{"error": "خطا در بروزرسانی وضعیت امانت"})
    }
    if !updated {
        return c.JSON(http.StatusNotFound, echo.Map{"error": "امانت یافت نشد یا قبلاً بازگردانده شده است"})
    }

    // Increment the available copies of the associated book.  We load
    // the book first to obtain its current counts, then adjust and update.
    book, err := bookRepo.GetBookByID(int(loan.BookID))
    if err == nil && book != nil {
        book.AvailableCopies++
        _, _ = bookRepo.UpdateBook(book)
    }

    // Publish a LoanReturned event.  Include the remaining copies after
    // incrementing the inventory.  DueDate is omitted for returned books.
    event := model.LoanEvent{
        EventType:       model.LoanReturned,
        LoanID:          loan.ID,
        UserID:          userID,
        BookID:          loan.BookID,
        Time:            time.Now(),
        RemainingCopies: 0,
    }
    // If we were able to load the book and update its count, include the
    // remaining copies value for logging purposes.
    if book != nil {
        event.RemainingCopies = int(book.AvailableCopies)
    }
    _ = queue.PublishEvent(event)

    return c.JSON(http.StatusOK, echo.Map{"message": "کتاب بازگردانده شد"})
}