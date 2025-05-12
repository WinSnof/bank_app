package domain

import (
	"errors"
	"regexp"
	"time"

	"gorm.io/gorm"
)

var (
	ErrInvalidCardNumber = errors.New("invalid card number")
	ErrInvalidExpiryDate = errors.New("invalid expiry date")
	ErrInvalidCVV        = errors.New("invalid CVV")
	ErrCardExpired       = errors.New("card has expired")
)

// Card представляет модель данных банковской карты.
type Card struct {
	gorm.Model
	Number       string    `json:"number" gorm:"type:text;not null" validate:"required"`
	ExpiryDate   string    `json:"expiry_date" gorm:"type:text;not null" validate:"required"`
	CVV          string    `json:"-" gorm:"type:text;not null" validate:"required"`
	UserID       uint      `json:"user_id" gorm:"not null"`
	AccountID    uint      `json:"account_id" gorm:"not null"`
	IsActive     bool      `json:"is_active" gorm:"default:true"`
	DailyLimit   float64   `json:"daily_limit" gorm:"default:100000"`
	MonthlyLimit float64   `json:"monthly_limit" gorm:"default:1000000"`
	LastUsed     time.Time `json:"last_used"`
}

// Validate проверяет все поля карты
func (c *Card) Validate() error {
	if c.AccountID == 0 {
		return errors.New("account_id is required")
	}
	if err := c.ValidateNumber(); err != nil {
		return err
	}
	if err := c.ValidateExpiryDate(); err != nil {
		return err
	}
	if err := c.ValidateCVV(); err != nil {
		return err
	}
	return nil
}

// ValidateNumber проверяет корректность номера карты
func (c *Card) ValidateNumber() error {
	// Проверка формата (16 цифр)
	numberRegex := regexp.MustCompile(`^\d{16}$`)
	if !numberRegex.MatchString(c.Number) {
		return ErrInvalidCardNumber
	}

	// Проверка по алгоритму Луна
	sum := 0
	alternate := false

	for i := len(c.Number) - 1; i >= 0; i-- {
		digit := int(c.Number[i] - '0')

		if alternate {
			digit *= 2
			if digit > 9 {
				digit = (digit % 10) + 1
			}
		}

		sum += digit
		alternate = !alternate
	}

	if sum%10 != 0 {
		return ErrInvalidCardNumber
	}

	return nil
}

// ValidateExpiryDate проверяет корректность срока действия
func (c *Card) ValidateExpiryDate() error {
	// Проверка формата MM/YY
	expiryRegex := regexp.MustCompile(`^(0[1-9]|1[0-2])/([0-9]{2})$`)
	if !expiryRegex.MatchString(c.ExpiryDate) {
		return ErrInvalidExpiryDate
	}

	// Парсинг даты
	month := c.ExpiryDate[:2]
	year := "20" + c.ExpiryDate[3:]

	expiryTime, err := time.Parse("2006-01", year+"-"+month)
	if err != nil {
		return ErrInvalidExpiryDate
	}

	// Проверка на истечение срока
	if expiryTime.Before(time.Now()) {
		return ErrCardExpired
	}

	return nil
}

// ValidateCVV проверяет корректность CVV
func (c *Card) ValidateCVV() error {
	cvvRegex := regexp.MustCompile(`^\d{3}$`)
	if !cvvRegex.MatchString(c.CVV) {
		return ErrInvalidCVV
	}
	return nil
}

// IsExpired проверяет, истек ли срок действия карты
func (c *Card) IsExpired() bool {
	month := c.ExpiryDate[:2]
	year := "20" + c.ExpiryDate[3:]
	expiryTime, err := time.Parse("2006-01", year+"-"+month)
	if err != nil {
		return true
	}
	return expiryTime.Before(time.Now())
}

// GetCardType определяет тип карты по номеру
func (c *Card) GetCardType() string {
	// Visa
	if matched, _ := regexp.MatchString(`^4[0-9]{12}(?:[0-9]{3})?$`, c.Number); matched {
		return "VISA"
	}
	// MasterCard
	if matched, _ := regexp.MatchString(`^5[1-5][0-9]{14}$`, c.Number); matched {
		return "MASTERCARD"
	}
	// MIR
	if matched, _ := regexp.MatchString(`^2[0-9]{15}$`, c.Number); matched {
		return "MIR"
	}
	return "UNKNOWN"
}

// BeforeUpdate хук для валидации перед обновлением
func (c *Card) BeforeUpdate(tx *gorm.DB) error {
	return c.Validate()
}

// ToDTO преобразует модель в DTO
func (c *Card) ToDTO() map[string]interface{} {
	return map[string]interface{}{
		"id":            c.ID,
		"number":        c.MaskNumber(),
		"expiry_date":   c.ExpiryDate,
		"is_active":     c.IsActive,
		"daily_limit":   c.DailyLimit,
		"monthly_limit": c.MonthlyLimit,
		"last_used":     c.LastUsed,
		"created_at":    c.CreatedAt,
		"updated_at":    c.UpdatedAt,
		"account_id":    c.AccountID,
	}
}

// MaskNumber маскирует номер карты
func (c *Card) MaskNumber() string {
	if len(c.Number) != 16 {
		return c.Number
	}
	return c.Number[:4] + " **** **** " + c.Number[12:]
}
