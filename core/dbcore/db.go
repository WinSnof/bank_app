package dbcore

import (
	"FinanceGolang/core/domain"
	"FinanceGolang/core/settings"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

type DBType string

const (
	SQLite   DBType = "sqlite"
	Postgres DBType = "postgresql"
)

type Handler struct {
	DB *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{DB: db}
}

// InitDB - инициализирует соединение с базой данных
func InitDB() (*gorm.DB, error) {
	// Инициализируем конфигурацию
	if err := settings.Init(); err != nil {
		return nil, fmt.Errorf("ошибка инициализации конфигурации: %v", err)
	}

	cfg := settings.Get()
	if cfg == nil {
		return nil, fmt.Errorf("конфигурация не инициализирована")
	}

	// Настройка логгера GORM
	loggerInstance := logger.New(
		log.New(log.Writer(), "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger.Info,
			Colorful:      true,
		},
	)

	var err error
	var dialector gorm.Dialector

	switch DBType(cfg.DBType) {
	case SQLite:
		dialector = sqlite.Open(cfg.DBPath)
	case Postgres:
		dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)
		dialector = postgres.Open(dsn)
	default:
		return nil, fmt.Errorf("неподдерживаемый тип базы данных: %s", cfg.DBType)
	}

	DB, err = gorm.Open(dialector, &gorm.Config{Logger: loggerInstance})
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к базе данных: %v", err)
	}

	err = CreateTables(DB)
	if err != nil {
		return nil, fmt.Errorf("ошибка при создании таблиц: %v", err)
	}

	// log.Printf("Инициализация базы данных типа: %s", cfg.DBType)
	log.Printf("Успешное подключение к базе данных типа: %s", cfg.DBType)

	return DB, nil
}

// CloseDB - закрывает соединение с базой данных
func CloseDB() {
	if DB != nil {
		db, err := DB.DB()
		if err != nil {
			log.Printf("Ошибка при получении соединения с базой данных: %v", err)
			return
		}
		if err := db.Close(); err != nil {
			log.Printf("Ошибка при закрытии соединения с базой данных: %v", err)
		}
	}
}

func ConnectDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func CreateTables(db *gorm.DB) error {
	// Создаем таблицы
	err := db.AutoMigrate(
		&domain.Role{},
		&domain.User{},
		&domain.UserRole{},
		&domain.Account{},
		&domain.Card{},
		&domain.Transaction{},
		&domain.Credit{},
		&domain.PaymentSchedule{},
		&domain.Analytics{},
		&domain.BalanceForecast{},
	)

	if err != nil {
		return fmt.Errorf("ошибка при создании таблиц: %v", err)
	}

	// Инициализируем роли после создания таблиц
	if err := InitializeRoles(db); err != nil {
		return fmt.Errorf("ошибка при инициализации ролей: %v", err)
	}

	// Создаем админа после создания всех таблиц и инициализации ролей
	if err := createAdmin(db); err != nil {
		return fmt.Errorf("ошибка при создании админа: %v", err)
	}

	return nil
}

func createAdmin(db *gorm.DB) error {
	// Проверяем, существует ли уже админ
	var existingAdmin domain.User
	if err := db.Where("username = ?", "admin").First(&existingAdmin).Error; err == nil {
		// Админ уже существует, ничего не делаем
		return nil
	}

	admin := &domain.User{
		Username: "admin",
		Password: "admin",
		Email:    "admin@example.com",
	}

	// Создаем роли, если их нет
	adminRole := domain.Role{Name: domain.RoleAdmin, Description: "Администратор системы"}
	userRole := domain.Role{Name: domain.RoleUser, Description: "Обычный пользователь"}

	if err := db.FirstOrCreate(&adminRole, domain.Role{Name: domain.RoleAdmin}).Error; err != nil {
		return fmt.Errorf("ошибка при создании роли админа: %v", err)
	}

	if err := db.FirstOrCreate(&userRole, domain.Role{Name: domain.RoleUser}).Error; err != nil {
		return fmt.Errorf("ошибка при создании роли пользователя: %v", err)
	}

	if err := db.Create(admin).Error; err != nil {
		return fmt.Errorf("ошибка при создании админа: %v", err)
	}

	return nil
}

// InitializeRoles создает базовые роли в системе
func InitializeRoles(db *gorm.DB) error {
	// Создаем базовые роли
	roles := []domain.Role{
		{Name: domain.RoleAdmin, Description: "Администратор"},
		{Name: domain.RoleUser, Description: "Пользователь"},
	}

	// Сохраняем роли в базе данных
	for _, role := range roles {
		if err := db.FirstOrCreate(&role, domain.Role{Name: role.Name}).Error; err != nil {
			return fmt.Errorf("ошибка при создании роли %s: %v", role.Name, err)
		}
	}

	return nil
}

func addNumberField(db *gorm.DB) error {
	// Обновляем существующие записи
	var accounts []domain.Account
	if err := db.Find(&accounts).Error; err != nil {
		return fmt.Errorf("ошибка при получении счетов: %v", err)
	}

	for _, account := range accounts {
		if account.Number == "" {
			account.Number = domain.GenerateAccountNumber()
			if err := db.Save(&account).Error; err != nil {
				return fmt.Errorf("ошибка при обновлении счета %d: %v", account.ID, err)
			}
		}
	}

	return nil
}
