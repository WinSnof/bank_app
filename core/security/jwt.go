package security

import (
	"FinanceGolang/core/domain"
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	jwt.StandardClaims
}

var jwtSecret = []byte("your_secret_key")

var ErrTokenExpired = errors.New("token expired")

func GenerateToken(userID uint, username, email string) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		Email:    email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 72).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ParseToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

func CutToken(tokenString string) (string, error) {
	if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
		tokenString = tokenString[7:]
	}
	return tokenString, nil
}

type AuthMiddlewareDeps struct {
	ValidateUserFromToken func(tokenString string) (*domain.User, error)
}

func AuthMiddleware(deps AuthMiddlewareDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем все заголовки Authorization
		authHeaders := c.Request.Header["Authorization"]
		if len(authHeaders) == 0 {
			c.JSON(401, gin.H{
				"status":  "error",
				"message": "Authorization header is required",
			})
			c.Abort()
			return
		}

		// Берем первый заголовок
		tokenString := authHeaders[0]

		// Удаляем префикс "Bearer " если он есть
		tokenString, err := CutToken(tokenString)
		if err != nil {
			c.JSON(401, gin.H{
				"status":  "error",
				"message": "Invalid token format",
			})
			c.Abort()
			return
		}

		// Проверяем токен и получаем пользователя
		user, err := deps.ValidateUserFromToken(tokenString)
		if err != nil {
			c.JSON(401, gin.H{
				"status":  "error",
				"message": err.Error(),
			})
			c.Abort()
			return
		}

		c.Set("userID", user.ID)
		c.Set("username", user.Username)
		c.Set("user", user)
		c.Next()
	}
}
