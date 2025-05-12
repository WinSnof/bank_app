package api

import (
	"FinanceGolang/core/domain"
	"FinanceGolang/core/services"
	"net/http"

	// "strconv"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	userService services.UserService
}

func CreateUserController(userService services.UserService) *UserController {
	return &UserController{userService: userService}
}

func (uc *UserController) GetCurrentUser(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "user not found",
		})
		return
	}

	user, err := uc.userService.GetUserByID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"user":   user,
	})
}

func (uc *UserController) UpdateCurrentUser(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "user not found",
		})
		return
	}

	var user domain.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "invalid request body",
		})
		return
	}

	user.ID = userID.(uint)
	if err := uc.userService.UpdateUser(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "user updated successfully",
	})
}

func (uc *UserController) DeleteCurrentUser(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "user not found",
		})
		return
	}

	if err := uc.userService.DeleteUser(userID.(uint)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "user deleted successfully",
	})
}
