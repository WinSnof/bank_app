package api

import (
	"FinanceGolang/core/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CbrController struct {
	externalService *services.ExternalService
}

func CreateCbrController(externalService *services.ExternalService) *CbrController {
	return &CbrController{externalService: externalService}
}

func (cc *CbrController) GetKeyRate(c *gin.Context) {
	keyRate, err := cc.externalService.GetKeyRate()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"rate":   keyRate,
	})
}
