package api

import (
	"FinanceGolang/core/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AdminController struct {
	scheduler *services.Scheduler
}

func CreateAdminController(scheduler *services.Scheduler) *AdminController {
	return &AdminController{
		scheduler: scheduler,
	}
}

// CheckPayments запускает проверку платежей вручную
func (c *AdminController) CheckPayments(ctx *gin.Context) {
	// Запускаем проверку платежей
	c.scheduler.CheckPayments()

	// Возвращаем результат
	ctx.JSON(http.StatusOK, gin.H{
		"message": "Проверка платежей выполнена",
		"status":  "success",
	})
}

// GetAllCredits возвращает список всех кредитов
func (c *AdminController) GetAllCredits(ctx *gin.Context) {
	credits, err := c.scheduler.GetAllCredits()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"credits": credits})
}
