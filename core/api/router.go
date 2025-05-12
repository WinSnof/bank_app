package api

import (
	"FinanceGolang/core/dbaccess"
	"FinanceGolang/core/dbcore"
	"FinanceGolang/core/security"
	"FinanceGolang/core/services"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gin-contrib/cache"
	"github.com/gin-contrib/cache/persistence"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Константы для API путей
const (
	APIPathRegister     = "/register"
	APIPathLogin        = "/login"
	APIPathMyUser       = "/my"
	APIPathAuthStatus   = "/security-status"
	APIPathUsers        = "/users"
	APIPathMe           = "/me"
	APIPathAccounts     = "/accounts"
	APIPathAll          = "/all"
	APIPathDeposit      = "/deposit"
	APIPathWithdraw     = "/withdraw"
	APIPathTransfer     = "/transfer"
	APIPathTransactions = "/transactions"
	APIPathCards        = "/cards"
	APIPathCredits      = "/credits"
	APIPathSchedule     = "/schedule"
	APIPathPayment      = "/payment"
	APIPathAnalytics    = "/analytics"
	APIPathForecast     = "/forecast"
	APIPathKeyRate      = "/keyrate"
)

// Константы для сообщений об ошибках
const (
	ErrUnauthorized   = "Не авторизован"
	ErrInternalServer = "Внутренняя ошибка сервера"
)

type Router struct{}

// NewRouter создает новый экземпляр маршрутизатора
func NewRouter() *Router {
	return &Router{}
}

// createAuthService создает сервис аутентификации
func (r *Router) createAuthService() services.AuthService {
	userRepo := dbaccess.UserRepositoryInstance(dbcore.DB)
	return services.AuthServiceInstance(userRepo)
}

// createUserService создает сервис пользователей
func (r *Router) createUserService() services.UserService {
	userRepo := dbaccess.UserRepositoryInstance(dbcore.DB)
	return services.UserServiceInstance(userRepo)
}

// createAccountService создает сервис счетов
func (r *Router) createAccountService() services.AccountService {
	accountRepo := dbaccess.AccountRepositoryInstance(dbcore.DB)
	transactionRepo := dbaccess.TransactionRepositoryInstance(dbcore.DB)
	return services.AccountServiceInstance(accountRepo, transactionRepo)
}

// createCardService создает сервис карт
func (r *Router) createCardService() services.CardService {
	cardRepo := dbaccess.CardRepositoryInstance(dbcore.DB)
	accountRepo := dbaccess.AccountRepositoryInstance(dbcore.DB)

	// Читаем публичный ключ из файла
	publicKeyBytes, err := ioutil.ReadFile("public_key.asc")
	if err != nil {
		logrus.WithError(err).Error("Ошибка чтения публичного ключа")
		publicKeyBytes = []byte("card_public_key_" + time.Now().Format("20060102150405"))
	}

	// Читаем HMAC секрет из файла
	hmacSecretBytes, err := ioutil.ReadFile("private_key.asc")
	if err != nil {
		logrus.WithError(err).Error("Ошибка чтения приватного ключа")
		hmacSecretBytes = []byte("card_hmac_secret_" + time.Now().Format("20060102150405"))
	}

	return services.CardServiceInstance(cardRepo, accountRepo, string(publicKeyBytes), hmacSecretBytes)
}

// createCreditService создает сервис кредитов
func (r *Router) createCreditService() services.CreditService {
	return services.CreditServiceInstance(
		dbaccess.CreditRepositoryInstance(dbcore.DB),
		dbaccess.AccountRepositoryInstance(dbcore.DB),
		dbaccess.TransactionRepositoryInstance(dbcore.DB),
		services.NewExternalService("", 0, "", "", ""),
	)
}

// createAnalyticsService создает сервис аналитики
func (r *Router) createAnalyticsService() *services.AnalyticsService {
	accountRepo := dbaccess.AccountRepositoryInstance(dbcore.DB)
	transactionRepo := dbaccess.TransactionRepositoryInstance(dbcore.DB)
	creditRepo := dbaccess.CreditRepositoryInstance(dbcore.DB)
	return services.NewAnalyticsService(transactionRepo, accountRepo, creditRepo)
}

// LoggerMiddleware логирует информацию о запросах
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Продолжаем обработку запроса
		c.Next()

		// Логируем информацию после обработки
		latency := time.Since(start)
		status := c.Writer.Status()
		clientIP := c.ClientIP()

		logrus.WithFields(logrus.Fields{
			"status":  status,
			"latency": latency,
			"client":  clientIP,
			"method":  method,
			"path":    path,
		}).Info("Обработан HTTP запрос")
	}
}

// ErrorHandlerMiddleware обрабатывает ошибки и возвращает стандартизированный ответ
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Проверяем наличие ошибок
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			status := http.StatusInternalServerError
			message := ErrInternalServer

			// Определяем тип ошибки и устанавливаем соответствующий статус
			switch err.(type) {
			case *gin.Error:
				status = http.StatusBadRequest
				message = err.Error()
			case error:
				if err.Error() == "unauthorized" {
					status = http.StatusUnauthorized
					message = ErrUnauthorized
				}
			}

			c.JSON(status, gin.H{
				"error": message,
			})
			c.Abort()
		}
	}
}

// CORSMiddleware обрабатывает CORS заголовки
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Проверяем на дублирование заголовков
		authHeaders := c.Request.Header["Authorization"]
		if len(authHeaders) > 1 {
			// Берем только первый заголовок
			c.Request.Header.Set("Authorization", authHeaders[0])
		}

		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// CacheMiddleware создает middleware для кэширования
func CacheMiddleware(store *persistence.InMemoryStore) gin.HandlerFunc {
	return cache.CachePage(store, time.Minute*5, func(c *gin.Context) {
		c.Next()
	})
}

// CompressionMiddleware сжимает ответы
func CompressionMiddleware() gin.HandlerFunc {
	return gzip.Gzip(gzip.DefaultCompression)
}

// RegisterAuthRoutes регистрирует маршруты аутентификации
func (r *Router) RegisterAuthRoutes(g *gin.RouterGroup) {
	authService := r.createAuthService()
	authController := CreateAuthController(authService)

	g.POST(APIPathRegister, authController.Register)
	g.POST(APIPathLogin, authController.Login)
	g.GET(APIPathMyUser, security.AuthMiddleware(security.AuthMiddlewareDeps{
		ValidateUserFromToken: authService.ValidateUserFromToken,
	}), authController.MyUser)
	g.GET(APIPathAuthStatus, security.AuthMiddleware(security.AuthMiddlewareDeps{
		ValidateUserFromToken: authService.ValidateUserFromToken,
	}), authController.AuthStatus)
}

// RegisterUserRoutes регистрирует маршруты пользователей
func (r *Router) RegisterUserRoutes(g *gin.RouterGroup) {
	authService := r.createAuthService()
	userService := r.createUserService()
	userController := CreateUserController(userService)

	users := g.Group(APIPathUsers)
	users.Use(security.AuthMiddleware(security.AuthMiddlewareDeps{
		ValidateUserFromToken: authService.ValidateUserFromToken,
	}))
	{
		users.GET(APIPathMe, userController.GetCurrentUser)
		users.PUT(APIPathMe, userController.UpdateCurrentUser)
		users.DELETE(APIPathMe, userController.DeleteCurrentUser)
	}
}

// RegisterAccountRoutes регистрирует маршруты счетов
func (r *Router) RegisterAccountRoutes(g *gin.RouterGroup) {
	authService := r.createAuthService()
	accountService := r.createAccountService()
	accountController := CreateAccountController(accountService)

	g.POST(APIPathAccounts, security.AuthMiddleware(security.AuthMiddlewareDeps{
		ValidateUserFromToken: authService.ValidateUserFromToken,
	}), accountController.CreateAccount)
	g.GET(APIPathAccounts, security.AuthMiddleware(security.AuthMiddlewareDeps{
		ValidateUserFromToken: authService.ValidateUserFromToken,
	}), accountController.GetAccountByUserID)
	g.GET(APIPathAccounts+APIPathAll, security.AuthMiddleware(security.AuthMiddlewareDeps{
		ValidateUserFromToken: authService.ValidateUserFromToken,
	}), accountController.GetAccountsAll)

	accountGroup := g.Group(APIPathAccounts + "/:id")
	accountGroup.Use(security.AuthMiddleware(security.AuthMiddlewareDeps{
		ValidateUserFromToken: authService.ValidateUserFromToken,
	}))
	{
		accountGroup.GET("", accountController.GetAccountByID)
		accountGroup.POST(APIPathDeposit, accountController.Deposit)
		accountGroup.POST(APIPathWithdraw, accountController.Withdraw)
		accountGroup.POST(APIPathTransfer, accountController.Transfer)
		accountGroup.GET(APIPathTransactions, accountController.GetTransactions)
	}
}

// RegisterCardRoutes регистрирует маршруты карт
func (r *Router) RegisterCardRoutes(g *gin.RouterGroup) {
	authService := r.createAuthService()
	cardService := r.createCardService()
	cardController := CreateCardController(cardService)

	g.POST(APIPathCards, security.AuthMiddleware(security.AuthMiddlewareDeps{
		ValidateUserFromToken: authService.ValidateUserFromToken,
	}), cardController.CreateCard)
	g.GET(APIPathCards, security.AuthMiddleware(security.AuthMiddlewareDeps{
		ValidateUserFromToken: authService.ValidateUserFromToken,
	}), cardController.GetAllCards)

	// Маршрут для получения информации о конкретной карте
	g.GET(APIPathCards+"/:id", security.AuthMiddleware(security.AuthMiddlewareDeps{
		ValidateUserFromToken: authService.ValidateUserFromToken,
	}), cardController.GetCardByID)
}

// RegisterKeyRateRoutes регистрирует маршруты ключевой ставки
func (r *Router) RegisterKeyRateRoutes(g *gin.RouterGroup) {
	authService := r.createAuthService()
	externalService := services.NewExternalService("", 0, "", "", "")
	cbrController := CreateCbrController(externalService)

	g.GET("", security.AuthMiddleware(security.AuthMiddlewareDeps{
		ValidateUserFromToken: authService.ValidateUserFromToken,
	}), cbrController.GetKeyRate)
}

// RegisterCreditRoutes регистрирует маршруты кредитов
func (r *Router) RegisterCreditRoutes(g *gin.RouterGroup) {
	authService := r.createAuthService()
	creditService := r.createCreditService()
	creditController := CreateCreditController(creditService)

	credits := g.Group(APIPathCredits)
	credits.Use(security.AuthMiddleware(security.AuthMiddlewareDeps{
		ValidateUserFromToken: authService.ValidateUserFromToken,
	}))
	{
		credits.POST("", creditController.CreateCredit)
		credits.GET("", creditController.GetUserCredits)
		credits.GET("/:id", creditController.GetCreditByID)
		credits.GET("/:id"+APIPathSchedule, creditController.GetPaymentSchedule)
		credits.POST("/:id"+APIPathPayment, creditController.ProcessPayment)
	}
}

// RegisterAnalyticsRoutes регистрирует маршруты аналитики
func (r *Router) RegisterAnalyticsRoutes(g *gin.RouterGroup) {
	authService := r.createAuthService()
	analyticsService := r.createAnalyticsService()
	analyticsController := CreateAnalyticsController(analyticsService)

	analytics := g.Group(APIPathAnalytics)
	analytics.Use(security.AuthMiddleware(security.AuthMiddlewareDeps{
		ValidateUserFromToken: authService.ValidateUserFromToken,
	}))
	{
		analytics.GET("", analyticsController.GetAnalytics)
		analytics.GET("/accounts/:id"+APIPathForecast, analyticsController.GetBalanceForecast)
		analytics.GET("/spending-categories", analyticsController.GetSpendingCategories)
		analytics.POST("/spending-categories", analyticsController.GetSpendingCategories)
	}
}

// RegisterAdminRoutes регистрирует маршруты админской части
func (r *Router) RegisterAdminRoutes(g *gin.RouterGroup) {
	authService := r.createAuthService()
	scheduler := services.NewScheduler(
		dbaccess.CreditRepositoryInstance(dbcore.DB),
		dbaccess.AccountRepositoryInstance(dbcore.DB),
		dbaccess.TransactionRepositoryInstance(dbcore.DB),
		services.NewExternalService("", 0, "", "", ""),
	)
	adminController := CreateAdminController(scheduler)

	admin := g.Group("/admin")
	admin.Use(security.AuthMiddleware(security.AuthMiddlewareDeps{
		ValidateUserFromToken: authService.ValidateUserFromToken,
	}))
	{
		admin.GET("/credits", adminController.GetAllCredits)
		admin.POST("/scheduler/check-payments", adminController.CheckPayments)
	}
}

// InitRoutes инициализирует все маршруты приложения
func (r *Router) InitRoutes() *gin.Engine {
	router := gin.Default()

	// Настраиваем обработку путей
	router.RedirectTrailingSlash = true
	router.RedirectFixedPath = true
	router.HandleMethodNotAllowed = true

	// Добавляем middleware в правильном порядке
	router.Use(LoggerMiddleware())
	router.Use(CORSMiddleware())
	router.Use(ErrorHandlerMiddleware())

	// Отключаем сжатие для определенных путей
	router.Use(func(c *gin.Context) {
		if c.Request.URL.Path == "/api/security/login" || c.Request.URL.Path == "/api/security/register" {
			c.Next()
			return
		}
		CompressionMiddleware()(c)
	})

	api := router.Group("/api")
	{
		// Регистрируем маршруты аутентификации
		securityGroup := api.Group("/security")
		{
			authController := CreateAuthController(r.createAuthService())
			securityGroup.POST("/register", authController.Register)
			securityGroup.POST("/login", authController.Login)
			securityGroup.GET("/my", security.AuthMiddleware(security.AuthMiddlewareDeps{
				ValidateUserFromToken: r.createAuthService().ValidateUserFromToken,
			}), authController.MyUser)
			securityGroup.GET("/security-status", security.AuthMiddleware(security.AuthMiddlewareDeps{
				ValidateUserFromToken: r.createAuthService().ValidateUserFromToken,
			}), authController.AuthStatus)
		}

		r.RegisterUserRoutes(api)
		r.RegisterAccountRoutes(api)
		r.RegisterCardRoutes(api)
		r.RegisterCreditRoutes(api)
		r.RegisterAnalyticsRoutes(api)
		r.RegisterAdminRoutes(api)
		r.RegisterKeyRateRoutes(api.Group(APIPathKeyRate))
	}

	return router
}
