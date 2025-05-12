package domain

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

var (
	ErrInvalidBalance     = errors.New("invalid balance amount")
	ErrInsufficientFunds  = errors.New("insufficient funds")
	ErrInvalidAccountType = errors.New("invalid account type")
)

type AccountType string

const (
	AccountTypeDebit   AccountType = "DEBIT"
	AccountTypeCredit  AccountType = "CREDIT"
	AccountTypeSavings AccountType = "SAVINGS"
)

type Account struct {
	gorm.Model
	Number        string     `json:"number" gorm:"unique;not null;default:''"`
	Balance       float64    `json:"balance" gorm:"type:decimal(20,2);not null;default:0"`
	UserID        uint       `json:"user_id" gorm:"not null"`
	IsActive      bool       `json:"is_active" gorm:"default:true"`
	InterestRate  float64    `json:"interest_rate" gorm:"type:decimal(5,2);default:0"`
	LastOperation *time.Time `json:"last_operation"`
	DailyLimit    float64    `json:"daily_limit" gorm:"type:decimal(20,2);default:100000"`
	MonthlyLimit  float64    `json:"monthly_limit" gorm:"type:decimal(20,2);default:1000000"`
}

// Validate проверяет все поля счета
func (a *Account) Validate() error {
	if err := a.ValidateBalance(); err != nil {
		return err
	}
	return nil
}

// ValidateBalance проверяет корректность баланса
func (a *Account) ValidateBalance() error {
	if a.Balance < 0 {
		return ErrInvalidBalance
	}
	return nil
}

// CanWithdraw проверяет возможность снятия средств
func (a *Account) CanWithdraw(amount float64) error {
	if amount <= 0 {
		return ErrInvalidBalance
	}
	if a.Balance < amount {
		return ErrInsufficientFunds
	}
	return nil
}

// Withdraw снимает средства со счета
func (a *Account) Withdraw(amount float64) error {
	if err := a.CanWithdraw(amount); err != nil {
		return err
	}
	a.Balance -= amount
	now := time.Now()
	a.LastOperation = &now
	return nil
}

// Deposit пополняет счет
func (a *Account) Deposit(amount float64) error {
	if amount <= 0 {
		return ErrInvalidBalance
	}
	a.Balance += amount
	now := time.Now()
	a.LastOperation = &now
	return nil
}

// BeforeCreate хук для валидации перед созданием
func (a *Account) BeforeCreate(tx *gorm.DB) error {
	if a.Number == "" {
		a.Number = GenerateAccountNumber()
	}
	return a.Validate()
}

// BeforeUpdate хук для валидации перед обновлением
func (a *Account) BeforeUpdate(tx *gorm.DB) error {
	if a.Number == "" {
		a.Number = GenerateAccountNumber()
	}
	return a.Validate()
}

// ToDTO преобразует модель в DTO
func (a *Account) ToDTO() map[string]interface{} {
	return map[string]interface{}{
		"id":             a.ID,
		"number":         a.Number,
		"balance":        a.Balance,
		"is_active":      a.IsActive,
		"interest_rate":  a.InterestRate,
		"last_operation": a.LastOperation,
		"daily_limit":    a.DailyLimit,
		"monthly_limit":  a.MonthlyLimit,
		"created_at":     a.CreatedAt,
		"updated_at":     a.UpdatedAt,
	}
}

// GenerateAccountNumber генерирует уникальный номер счета
func GenerateAccountNumber() string {
	return time.Now().Format("20060102150405") + RandomString(6)
}

// RandomString генерирует случайную строку заданной длины
func RandomString(length int) string {
	const charset = "0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
