package domain

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

var (
	ErrInvalidCreditAmount  = errors.New("invalid credit amount")
	ErrInvalidTerm          = errors.New("invalid credit term")
	ErrInvalidInterestRate  = errors.New("invalid interest rate")
	ErrInvalidPaymentDay    = errors.New("invalid payment day")
	ErrCreditAlreadyActive  = errors.New("credit is already active")
	ErrCreditNotActive      = errors.New("credit is not active")
	ErrInvalidPaymentAmount = errors.New("invalid payment amount")
)

type CreditStatus string

const (
	CreditStatusPending   CreditStatus = "PENDING"
	CreditStatusActive    CreditStatus = "ACTIVE"
	CreditStatusPaid      CreditStatus = "PAID"
	CreditStatusOverdue   CreditStatus = "OVERDUE"
	CreditStatusCancelled CreditStatus = "CANCELLED"
)

type PaymentStatus string

const (
	PaymentStatusPending PaymentStatus = "pending"
	PaymentStatusPaid    PaymentStatus = "paid"
	PaymentStatusOverdue PaymentStatus = "overdue"
)

type Credit struct {
	gorm.Model
	AccountID     uint         `json:"account_id" gorm:"not null"`
	UserID        uint         `json:"user_id" gorm:"not null"`
	Amount        float64      `json:"amount" gorm:"type:decimal(20,2);not null"`
	Term          int          `json:"term" gorm:"not null"` // в месяцах
	InterestRate  float64      `json:"interest_rate" gorm:"type:decimal(5,2);not null"`
	Status        CreditStatus `json:"status" gorm:"type:varchar(20);not null;default:'PENDING'"`
	StartDate     time.Time    `json:"start_date"`
	EndDate       time.Time    `json:"end_date"`
	PaymentDay    int          `json:"payment_day" gorm:"not null"` // день месяца для платежа
	NextPayment   time.Time    `json:"next_payment"`
	TotalPaid     float64      `json:"total_paid" gorm:"type:decimal(20,2);default:0"`
	RemainingDebt float64      `json:"remaining_debt" gorm:"type:decimal(20,2);not null"`
	OverdueAmount float64      `json:"overdue_amount" gorm:"type:decimal(20,2);default:0"`
	LastPayment   time.Time    `json:"last_payment"`
}

// Validate проверяет все поля кредита
func (c *Credit) Validate() error {
	if err := c.ValidateAmount(); err != nil {
		return err
	}
	if err := c.ValidateTerm(); err != nil {
		return err
	}
	if err := c.ValidateInterestRate(); err != nil {
		return err
	}
	if err := c.ValidatePaymentDay(); err != nil {
		return err
	}
	if err := c.ValidateStatus(); err != nil {
		return err
	}
	return nil
}

// ValidateAmount проверяет корректность суммы кредита
func (c *Credit) ValidateAmount() error {
	if c.Amount <= 0 {
		return ErrInvalidCreditAmount
	}
	return nil
}

// ValidateTerm проверяет корректность срока кредита
func (c *Credit) ValidateTerm() error {
	if c.Term <= 0 || c.Term > 360 { // максимум 30 лет
		return ErrInvalidTerm
	}
	return nil
}

// ValidateInterestRate проверяет корректность процентной ставки
func (c *Credit) ValidateInterestRate() error {
	if c.InterestRate <= 0 || c.InterestRate > 100 {
		return ErrInvalidInterestRate
	}
	return nil
}

// ValidatePaymentDay проверяет корректность дня платежа
func (c *Credit) ValidatePaymentDay() error {
	if c.PaymentDay < 1 || c.PaymentDay > 31 {
		return ErrInvalidPaymentDay
	}
	return nil
}

// ValidateStatus проверяет корректность статуса
func (c *Credit) ValidateStatus() error {
	switch c.Status {
	case CreditStatusPending, CreditStatusActive, CreditStatusPaid,
		CreditStatusOverdue, CreditStatusCancelled:
		return nil
	default:
		return errors.New("invalid credit status")
	}
}

// CalculateMonthlyPayment рассчитывает ежемесячный платеж
func (c *Credit) CalculateMonthlyPayment() float64 {
	// Формула аннуитетного платежа
	monthlyRate := c.InterestRate / 12 / 100
	denominator := 1 - 1/pow(1+monthlyRate, float64(c.Term))
	return c.Amount * monthlyRate / denominator
}

// CalculateTotalAmount рассчитывает общую сумму к возврату
func (c *Credit) CalculateTotalAmount() float64 {
	return c.CalculateMonthlyPayment() * float64(c.Term)
}

// CalculateRemainingDebt рассчитывает оставшийся долг
func (c *Credit) CalculateRemainingDebt() float64 {
	return c.CalculateTotalAmount() - c.TotalPaid
}

// IsOverdue проверяет, просрочен ли кредит
func (c *Credit) IsOverdue() bool {
	return c.Status == CreditStatusActive && time.Now().After(c.NextPayment)
}

// UpdateStatus обновляет статус кредита
func (c *Credit) UpdateStatus() {
	if c.Status == CreditStatusActive {
		if c.IsOverdue() {
			c.Status = CreditStatusOverdue
		} else if c.RemainingDebt <= 0 {
			c.Status = CreditStatusPaid
		}
	}
}

// MakePayment вносит платеж по кредиту
func (c *Credit) MakePayment(amount float64) error {
	if c.Status != CreditStatusActive && c.Status != CreditStatusOverdue {
		return ErrCreditNotActive
	}
	if amount <= 0 {
		return ErrInvalidPaymentAmount
	}

	c.TotalPaid += amount
	c.RemainingDebt = c.CalculateRemainingDebt()
	c.LastPayment = time.Now()

	// Обновляем дату следующего платежа
	c.NextPayment = c.CalculateNextPaymentDate()

	// Обновляем статус
	c.UpdateStatus()

	return nil
}

// CalculateNextPaymentDate рассчитывает дату следующего платежа
func (c *Credit) CalculateNextPaymentDate() time.Time {
	next := c.LastPayment.AddDate(0, 1, 0)
	// Устанавливаем день платежа
	if c.PaymentDay > 28 {
		// Для месяцев с меньшим количеством дней
		lastDay := time.Date(next.Year(), next.Month()+1, 0, 0, 0, 0, 0, next.Location()).Day()
		if c.PaymentDay > lastDay {
			next = time.Date(next.Year(), next.Month(), lastDay, 0, 0, 0, 0, next.Location())
		} else {
			next = time.Date(next.Year(), next.Month(), c.PaymentDay, 0, 0, 0, 0, next.Location())
		}
	} else {
		next = time.Date(next.Year(), next.Month(), c.PaymentDay, 0, 0, 0, 0, next.Location())
	}
	return next
}

// BeforeCreate хук для валидации перед созданием
func (c *Credit) BeforeCreate(tx *gorm.DB) error {
	return c.Validate()
}

// BeforeUpdate хук для валидации перед обновлением
func (c *Credit) BeforeUpdate(tx *gorm.DB) error {
	return c.Validate()
}

// ToDTO преобразует модель в DTO
func (c *Credit) ToDTO() map[string]interface{} {
	return map[string]interface{}{
		"id":             c.ID,
		"account_id":     c.AccountID,
		"user_id":        c.UserID,
		"amount":         c.Amount,
		"term":           c.Term,
		"interest_rate":  c.InterestRate,
		"status":         c.Status,
		"start_date":     c.StartDate,
		"end_date":       c.EndDate,
		"payment_day":    c.PaymentDay,
		"next_payment":   c.NextPayment,
		"total_paid":     c.TotalPaid,
		"remaining_debt": c.RemainingDebt,
		"overdue_amount": c.OverdueAmount,
		"last_payment":   c.LastPayment,
		"created_at":     c.CreatedAt,
		"updated_at":     c.UpdatedAt,
	}
}

// pow возводит число в степень
func pow(x float64, n float64) float64 {
	if n == 0 {
		return 1
	}
	if n < 0 {
		return 1 / pow(x, -n)
	}
	if int(n)%2 == 0 {
		return pow(x*x, n/2)
	}
	return x * pow(x*x, (n-1)/2)
}

type PaymentSchedule struct {
	ID            uint          `json:"id" gorm:"primaryKey"`
	CreditID      uint          `json:"credit_id"`
	PaymentNumber int           `json:"payment_number"`
	DueDate       time.Time     `json:"due_date"`
	Amount        float64       `json:"amount"`
	Interest      float64       `json:"interest"`
	Principal     float64       `json:"principal"`
	TotalAmount   float64       `json:"total_amount"`
	Status        PaymentStatus `json:"status" gorm:"type:varchar(20)"`
	PaidAt        *time.Time    `json:"paid_at,omitempty"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`

	Credit Credit `gorm:"foreignKey:CreditID" json:"credit"`
}

// ToDTO преобразует структуру PaymentSchedule в DTO
func (p *PaymentSchedule) ToDTO() map[string]interface{} {
	dto := map[string]interface{}{
		"id":             p.ID,
		"credit_id":      p.CreditID,
		"payment_number": p.PaymentNumber,
		"due_date":       p.DueDate,
		"amount":         p.Amount,
		"interest":       p.Interest,
		"principal":      p.Principal,
		"total_amount":   p.TotalAmount,
		"status":         p.Status,
		"created_at":     p.CreatedAt,
		"updated_at":     p.UpdatedAt,
	}

	if p.PaidAt != nil {
		dto["paid_at"] = *p.PaidAt
	}

	// Добавляем информацию о кредите
	if p.Credit.ID > 0 {
		dto["credit"] = p.Credit.ToDTO()
	}

	return dto
}
