package api

import (
	"FinanceGolang/core/services"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type AnalyticsController struct {
	analyticsService *services.AnalyticsService
}

func CreateAnalyticsController(analyticsService *services.AnalyticsService) *AnalyticsController {
	return &AnalyticsController{
		analyticsService: analyticsService,
	}
}

// GetAnalytics возвращает статистику доходов и расходов
func (c *AnalyticsController) GetAnalytics(ctx *gin.Context) {
	accountIDStr := ctx.Query("account_id")
	if accountIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Не указан ID счета"})
		return
	}

	accountID, err := strconv.ParseUint(accountIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID счета"})
		return
	}

	startDateStr := ctx.Query("start_date")
	if startDateStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Не указана дата начала"})
		return
	}

	endDateStr := ctx.Query("end_date")
	if endDateStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Не указана дата окончания"})
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат даты начала"})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат даты окончания"})
		return
	}

	stats, err := c.analyticsService.GetIncomeExpenseStats(uint(accountID), startDate, endDate)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, stats)
}

// GetBalanceForecast возвращает прогноз баланса
func (c *AnalyticsController) GetBalanceForecast(ctx *gin.Context) {
	accountIDStr := ctx.Param("id")
	if accountIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Не указан ID счета"})
		return
	}

	accountID, err := strconv.ParseUint(accountIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID счета"})
		return
	}

	monthsStr := ctx.Query("months")
	months := 6 // По умолчанию прогноз на 6 месяцев
	if monthsStr != "" {
		months, err = strconv.Atoi(monthsStr)
		if err != nil || months <= 0 {
			months = 6
		}
	}

	forecast, err := c.analyticsService.GetBalanceForecast(uint(accountID), months)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, forecast)
}

// Структура для получения параметров из JSON-тела
type SpendingCategoriesRequest struct {
	AccountID uint   `json:"account_id"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

// GetSpendingCategories возвращает статистику по категориям расходов
func (c *AnalyticsController) GetSpendingCategories(ctx *gin.Context) {
	var accountID uint
	var startDateStr, endDateStr string

	// Проверяем Content-Type запроса
	contentType := ctx.GetHeader("Content-Type")

	// Если запрос в формате JSON
	if contentType == "application/json" {
		var req SpendingCategoriesRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса"})
			return
		}

		if req.AccountID == 0 {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Не указан ID счета"})
			return
		}
		accountID = req.AccountID
		startDateStr = req.StartDate
		endDateStr = req.EndDate
	} else {
		// Получаем параметры из URL-запроса
		accountIDStr := ctx.Query("account_id")
		if accountIDStr == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Не указан ID счета"})
			return
		}

		accountIDUint, err := strconv.ParseUint(accountIDStr, 10, 32)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID счета"})
			return
		}
		accountID = uint(accountIDUint)
		startDateStr = ctx.Query("start_date")
		endDateStr = ctx.Query("end_date")
	}

	if startDateStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Не указана дата начала"})
		return
	}

	if endDateStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Не указана дата окончания"})
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат даты начала"})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат даты окончания"})
		return
	}

	categories, err := c.analyticsService.GetSpendingCategories(accountID, startDate, endDate)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, categories)
}
