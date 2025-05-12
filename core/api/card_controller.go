package api

import (
	"FinanceGolang/core/domain"
	"FinanceGolang/core/services"

	// "FinanceGolang/core/dbaccess"
	// "FinanceGolang/core/dbcore"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CardController struct {
	cardService services.CardService
}

func CreateCardController(cardService services.CardService) *CardController {
	return &CardController{cardService: cardService}
}

func (cc *CardController) CreateCard(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "user not found"})
		return
	}

	var card domain.Card
	if err := c.ShouldBindJSON(&card); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "invalid request body",
		})
		return
	}

	if card.AccountID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "account_id is required",
		})
		return
	}

	unsecureCard, err := cc.cardService.CreateCard(&card, userID.(uint))
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "account does not belong to the user" {
			status = http.StatusForbidden
		}
		c.JSON(status, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "card created successfully",
		"card":    unsecureCard,
	})
}

func (cc *CardController) GetCardByID(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "user not found",
		})
		return
	}

	cardID := c.Param("id")
	if cardID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "card ID is required",
		})
		return
	}

	// Преобразуем ID карты из строки в uint
	cardIDUint, err := strconv.ParseUint(cardID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "invalid card ID",
		})
		return
	}

	card, err := cc.cardService.GetCardByID(uint(cardIDUint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	// Проверка, принадлежит ли карта пользователю
	if card.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{
			"status":  "error",
			"message": "card does not belong to the user",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"card":   card.ToDTO(),
	})
}

func (cc *CardController) GetAllCards(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "user not found",
		})
		return
	}

	cards, err := cc.cardService.GetUserCards(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	// Преобразуем карты в DTO
	var cardDTOs []map[string]interface{}
	for _, card := range cards {
		payloads := card.ToDTO()
		// Логируем каждую карту для отладки
		fmt.Printf("Card DTO: %+v\n", payloads)
		cardDTOs = append(cardDTOs, payloads)
	}

	response := gin.H{
		"status": "success",
		"cards":  cardDTOs,
	}
	// Логируем финальный ответ
	fmt.Printf("Response: %+v\n", response)

	c.JSON(http.StatusOK, response)
}
