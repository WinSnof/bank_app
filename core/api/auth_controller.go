package api

import (
	"FinanceGolang/core/domain"
	"FinanceGolang/core/services"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authService services.AuthService
}

func CreateAuthController(authService services.AuthService) *AuthController {
	return &AuthController{authService: authService}
}

func (h *AuthController) Register(c *gin.Context) {
	log.Printf("Register request received: %v", c.Request.Body)

	var user domain.User
	if err := c.ShouldBindJSON(&user); err != nil {
		log.Printf("Error binding JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	log.Printf("Attempting to register user: %s", user.Username)

	if err := h.authService.Register(&user); err != nil {
		log.Printf("Registration error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "user created successfully",
	})
}

func (h *AuthController) Login(c *gin.Context) {
	log.Printf("Login request received: %v", c.Request.Body)

	var user domain.User
	if err := c.ShouldBindJSON(&user); err != nil {
		log.Printf("Error binding JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Attempting to login user: %s", user.Username)

	token, err := h.authService.Login(&user)
	if err != nil {
		log.Printf("Login error: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *AuthController) MyUser(c *gin.Context) {
	log.Printf("MyUser request received")

	username, exists := c.Get("username")
	if !exists {
		log.Printf("Username not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	usernameStr, ok := username.(string)
	if !ok {
		log.Printf("Username is not a string")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid username type"})
		return
	}

	log.Printf("Getting user info for: %s", usernameStr)

	user, err := h.authService.GetUserByUsernameWithoutPassword(usernameStr)
	if err != nil {
		log.Printf("Error getting user info: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to retrieve user information"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Hello " + usernameStr, "user": user})
}

func (h *AuthController) AuthStatus(c *gin.Context) {
	log.Printf("AuthStatus request received")

	token := c.GetHeader("Authorization")
	log.Printf("Token received: %s", token)

	isValid, err := h.authService.AuthStatus(token)
	if err != nil || !isValid {
		log.Printf("Auth status error: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "invalid token",
			"isValid": isValid,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "token is valid",
		"isValid": isValid,
	})
}
