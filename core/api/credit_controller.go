package api

import (
	"FinanceGolang/core/services"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type CreditController struct {
	creditService services.CreditService
}

func CreateCreditController(creditService services.CreditService) *CreditController {
	return &CreditController{creditService: creditService}
}

type CreateCreditRequest struct {
	AccountID   uint    `json:"account_id" binding:"required"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	TermMonths  int     `json:"term_months" binding:"required,gt=0"`
	Description string  `json:"description"`
}

type CreditResponse struct {
	ID             uint      `json:"id"`
	UserID         uint      `json:"user_id"`
	AccountID      uint      `json:"account_id"`
	Amount         float64   `json:"amount"`
	InterestRate   float64   `json:"interest_rate"`
	TermMonths     int       `json:"term_months"`
	MonthlyPayment float64   `json:"monthly_payment"`
	Status         string    `json:"status"`
	StartDate      time.Time `json:"start_date"`
	EndDate        time.Time `json:"end_date"`
	PaymentDay     int       `json:"payment_day"`
	NextPayment    time.Time `json:"next_payment"`
	TotalPaid      float64   `json:"total_paid"`
	RemainingDebt  float64   `json:"remaining_debt"`
	OverdueAmount  float64   `json:"overdue_amount"`
	LastPayment    time.Time `json:"last_payment"`
	Description    string    `json:"description"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type ProcessPaymentRequest struct {
	PaymentNumber int `json:"payment_number" binding:"required,gt=0"`
}

func (c *CreditController) CreateCredit(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	var req CreateCreditRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Дополнительная обработка - убеждаемся, что description всегда строка
	description := req.Description
	if description == "" {
		description = "Потребительский кредит"
	}

	credit, err := c.creditService.CreateCredit(
		userID.(uint),
		req.AccountID,
		req.Amount,
		req.TermMonths,
		description,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Преобразуем модель в структуру ответа
	response := CreditResponse{
		ID:             credit.ID,
		UserID:         credit.UserID,
		AccountID:      credit.AccountID,
		Amount:         credit.Amount,
		InterestRate:   credit.InterestRate,
		TermMonths:     credit.Term,
		MonthlyPayment: credit.CalculateMonthlyPayment(),
		Status:         string(credit.Status),
		StartDate:      credit.StartDate,
		EndDate:        credit.EndDate,
		PaymentDay:     credit.PaymentDay,
		NextPayment:    credit.NextPayment,
		TotalPaid:      credit.TotalPaid,
		RemainingDebt:  credit.RemainingDebt,
		OverdueAmount:  credit.OverdueAmount,
		LastPayment:    credit.LastPayment,
		Description:    description,
		CreatedAt:      credit.CreatedAt,
		UpdatedAt:      credit.UpdatedAt,
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "credit created successfully",
		"credit":  response,
	})
}

func (c *CreditController) GetCreditByID(ctx *gin.Context) {
	creditID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid credit id"})
		return
	}

	credit, err := c.creditService.GetCreditByID(uint(creditID))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "credit not found"})
		return
	}

	// Проверяем, если LastPayment пустая, установим её равной StartDate
	if credit.LastPayment.IsZero() {
		credit.LastPayment = credit.StartDate
	}

	response := CreditResponse{
		ID:             credit.ID,
		UserID:         credit.UserID,
		AccountID:      credit.AccountID,
		Amount:         credit.Amount,
		InterestRate:   credit.InterestRate,
		TermMonths:     credit.Term,
		MonthlyPayment: credit.CalculateMonthlyPayment(),
		Status:         string(credit.Status),
		StartDate:      credit.StartDate,
		EndDate:        credit.EndDate,
		PaymentDay:     credit.PaymentDay,
		NextPayment:    credit.NextPayment,
		TotalPaid:      credit.TotalPaid,
		RemainingDebt:  credit.RemainingDebt,
		OverdueAmount:  credit.OverdueAmount,
		LastPayment:    credit.LastPayment,
		CreatedAt:      credit.CreatedAt,
		UpdatedAt:      credit.UpdatedAt,
	}

	ctx.JSON(http.StatusOK, response)
}

func (c *CreditController) GetUserCredits(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	credits, err := c.creditService.GetUserCredits(userID.(uint))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	responses := make([]CreditResponse, len(credits))
	for i, credit := range credits {
		// Проверяем, если LastPayment пустая, установим её равной StartDate
		if credit.LastPayment.IsZero() {
			credit.LastPayment = credit.StartDate
		}

		responses[i] = CreditResponse{
			ID:             credit.ID,
			UserID:         credit.UserID,
			AccountID:      credit.AccountID,
			Amount:         credit.Amount,
			InterestRate:   credit.InterestRate,
			TermMonths:     credit.Term,
			MonthlyPayment: credit.CalculateMonthlyPayment(),
			Status:         string(credit.Status),
			StartDate:      credit.StartDate,
			EndDate:        credit.EndDate,
			PaymentDay:     credit.PaymentDay,
			NextPayment:    credit.NextPayment,
			TotalPaid:      credit.TotalPaid,
			RemainingDebt:  credit.RemainingDebt,
			OverdueAmount:  credit.OverdueAmount,
			LastPayment:    credit.LastPayment,
			CreatedAt:      credit.CreatedAt,
			UpdatedAt:      credit.UpdatedAt,
		}
	}

	ctx.JSON(http.StatusOK, gin.H{"credits": responses})
}

func (c *CreditController) GetPaymentSchedule(ctx *gin.Context) {
	creditID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid credit id"})
		return
	}

	schedule, err := c.creditService.GetPaymentSchedule(uint(creditID))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "payment schedule not found"})
		return
	}

	// Получаем кредит, чтобы обновить данные о нем
	credit, err := c.creditService.GetCreditByID(uint(creditID))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "credit not found"})
		return
	}

	// Проверяем, если LastPayment пустая, установим её равной StartDate
	if credit.LastPayment.IsZero() {
		credit.LastPayment = credit.StartDate
	}

	// Заполняем информацию о кредите в каждом платеже и преобразуем в DTO
	scheduleDTOs := make([]map[string]interface{}, len(schedule))
	for i := range schedule {
		schedule[i].Credit = *credit
		scheduleDTOs[i] = schedule[i].ToDTO()
	}

	ctx.JSON(http.StatusOK, gin.H{"schedule": scheduleDTOs})
}

func (c *CreditController) ProcessPayment(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid credit ID"})
		return
	}

	var req ProcessPaymentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.creditService.ProcessPayment(uint(id), req.PaymentNumber); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "payment processed successfully"})
}
