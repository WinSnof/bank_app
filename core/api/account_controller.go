package api

import (
	"FinanceGolang/core/domain"
	"FinanceGolang/core/services"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AccountController struct {
	accountService services.AccountService
}

func CreateAccountController(accountService services.AccountService) *AccountController {
	return &AccountController{accountService: accountService}
}

// Структуры запросов
type DepositRequest struct {
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Description string  `json:"description"`
}

type WithdrawRequest struct {
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Description string  `json:"description"`
}

type TransferRequest struct {
	ToAccountID uint    `json:"to_account_id" binding:"required"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Description string  `json:"description"`
}

// Базовые операции со счетом
func (h *AccountController) CreateAccount(c *gin.Context) {
	var account domain.Account
	if err := c.ShouldBindJSON(&account); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}
	if err := h.accountService.CreateAccount(&account, c.MustGet("userID").(uint)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": err.Error(),
			"error":  "could not create account",
		})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "account created successfully",
		"account": account,
	})
}

func (h *AccountController) GetAccountByUserID(c *gin.Context) {
	userID, exists := c.MustGet("userID").(uint)

	fmt.Println("UserID from context:", userID)
	fmt.Println("UserID exists:", exists)

	account, err := h.accountService.GetAccountByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}
	if account == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":   "success",
			"message":  "no accounts found",
			"accounts": account,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"accounts": account,
	})
}

func (h *AccountController) GetAccountsAll(c *gin.Context) {
	accounts, err := h.accountService.GetAllAccounts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}
	if len(accounts) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"status":   "success",
			"message":  "no accounts found",
			"accounts": accounts,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"accounts": accounts,
	})
}

// Операции с балансом
func (h *AccountController) Deposit(c *gin.Context) {
	accountIDStr := c.Param("id")
	if accountIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "account ID is required"})
		return
	}

	accountID, err := strconv.ParseUint(accountIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account ID"})
		return
	}

	var req DepositRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.accountService.Deposit(uint(accountID), req.Amount, req.Description); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deposit successful"})
}

func (h *AccountController) Withdraw(c *gin.Context) {
	accountIDStr := c.Param("id")
	if accountIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "account ID is required"})
		return
	}

	accountID, err := strconv.ParseUint(accountIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account ID"})
		return
	}

	var req WithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.accountService.Withdraw(uint(accountID), req.Amount, req.Description); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "withdrawal successful"})
}

func (h *AccountController) Transfer(c *gin.Context) {
	fromAccountIDStr := c.Param("id")
	if fromAccountIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "account ID is required"})
		return
	}

	fromAccountID, err := strconv.ParseUint(fromAccountIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account ID"})
		return
	}

	var req TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.accountService.Transfer(uint(fromAccountID), req.ToAccountID, req.Amount, req.Description); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "transfer successful"})
}

// Операции с транзакциями
func (h *AccountController) GetTransactions(c *gin.Context) {
	accountIDStr := c.Param("id")
	if accountIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "account ID is required"})
		return
	}

	accountID, err := strconv.ParseUint(accountIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account ID"})
		return
	}

	transactions, err := h.accountService.GetTransactions(uint(accountID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Преобразуем транзакции в DTO
	transactionDTOs := make([]map[string]interface{}, len(transactions))
	for i, t := range transactions {
		transactionDTOs[i] = t.ToDTO()
	}

	c.JSON(http.StatusOK, gin.H{"transactions": transactionDTOs})
}

// Получение информации о конкретном счете
func (h *AccountController) GetAccountByID(c *gin.Context) {
	accountIDStr := c.Param("id")
	if accountIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "account ID is required",
		})
		return
	}

	accountID, err := strconv.ParseUint(accountIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "invalid account ID",
		})
		return
	}

	account, err := h.accountService.GetAccountByID(uint(accountID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"account": account,
	})
}
