package domain

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

var (
	ErrInvalidAmount      = errors.New("invalid transaction amount")
	ErrInvalidType        = errors.New("invalid transaction type")
	ErrInvalidStatus      = errors.New("invalid transaction status")
	ErrTransactionExpired = errors.New("transaction has expired")
)

type TransactionType string

const (
	TransactionTypeTransfer   TransactionType = "TRANSFER"
	TransactionTypeDeposit    TransactionType = "DEPOSIT"
	TransactionTypeWithdrawal TransactionType = "WITHDRAWAL"
	TransactionTypePayment    TransactionType = "PAYMENT"
	TransactionTypeCredit     TransactionType = "CREDIT"
)

type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "PENDING"
	TransactionStatusCompleted TransactionStatus = "COMPLETED"
	TransactionStatusFailed    TransactionStatus = "FAILED"
	TransactionStatusCancelled TransactionStatus = "CANCELLED"
)

type Transaction struct {
	gorm.Model
	Type          TransactionType   `json:"type" gorm:"type:varchar(20);not null"`
	Status        TransactionStatus `json:"status" gorm:"type:varchar(20);not null;default:'PENDING'"`
	Amount        float64           `json:"amount" gorm:"type:decimal(20,2);not null"`
	FromAccountID uint              `json:"from_account_id"`
	ToAccountID   uint              `json:"to_account_id"`
	Description   string            `json:"description" gorm:"type:text"`
	Metadata      string            `json:"metadata" gorm:"type:jsonb"`
	ExpiresAt     time.Time         `json:"expires_at"`
	CompletedAt   *time.Time        `json:"completed_at"`
	FailedAt      *time.Time        `json:"failed_at"`
	Error         string            `json:"error" gorm:"type:text"`
}

// Validate проверяет все поля транзакции
func (t *Transaction) Validate() error {
	if err := t.ValidateType(); err != nil {
		return err
	}
	if err := t.ValidateStatus(); err != nil {
		return err
	}
	if err := t.ValidateAmount(); err != nil {
		return err
	}
	if err := t.ValidateAccounts(); err != nil {
		return err
	}
	return nil
}

// ValidateType проверяет корректность типа транзакции
func (t *Transaction) ValidateType() error {
	switch t.Type {
	case TransactionTypeTransfer, TransactionTypeDeposit, TransactionTypeWithdrawal,
		TransactionTypePayment, TransactionTypeCredit:
		return nil
	default:
		return ErrInvalidType
	}
}

// ValidateStatus проверяет корректность статуса транзакции
func (t *Transaction) ValidateStatus() error {
	switch t.Status {
	case TransactionStatusPending, TransactionStatusCompleted,
		TransactionStatusFailed, TransactionStatusCancelled:
		return nil
	default:
		return ErrInvalidStatus
	}
}

// ValidateAmount проверяет корректность суммы
func (t *Transaction) ValidateAmount() error {
	if t.Amount <= 0 {
		return ErrInvalidAmount
	}
	return nil
}

// ValidateAccounts проверяет корректность счетов
func (t *Transaction) ValidateAccounts() error {
	switch t.Type {
	case TransactionTypeTransfer:
		if t.FromAccountID == 0 || t.ToAccountID == 0 {
			return errors.New("both accounts are required for transfer")
		}
		if t.FromAccountID == t.ToAccountID {
			return errors.New("cannot transfer to the same account")
		}
	case TransactionTypeDeposit:
		if t.ToAccountID == 0 {
			return errors.New("destination account is required for deposit")
		}
	case TransactionTypeWithdrawal:
		if t.FromAccountID == 0 {
			return errors.New("source account is required for withdrawal")
		}
	}
	return nil
}

// IsExpired проверяет, истек ли срок действия транзакции
func (t *Transaction) IsExpired() bool {
	return !t.ExpiresAt.IsZero() && t.ExpiresAt.Before(time.Now())
}

// Complete помечает транзакцию как завершенную
func (t *Transaction) Complete() {
	now := time.Now()
	t.Status = TransactionStatusCompleted
	t.CompletedAt = &now
}

// Fail помечает транзакцию как неудачную
func (t *Transaction) Fail(err error) {
	now := time.Now()
	t.Status = TransactionStatusFailed
	t.FailedAt = &now
	t.Error = err.Error()
}

// Cancel помечает транзакцию как отмененную
func (t *Transaction) Cancel() {
	t.Status = TransactionStatusCancelled
}

// BeforeCreate хук для валидации перед созданием
func (t *Transaction) BeforeCreate(tx *gorm.DB) error {
	return t.Validate()
}

// BeforeUpdate хук для валидации перед обновлением
func (t *Transaction) BeforeUpdate(tx *gorm.DB) error {
	return t.Validate()
}

// ToDTO преобразует модель в DTO
func (t *Transaction) ToDTO() map[string]interface{} {
	amount := t.Amount

	// Для платежей по кредиту, снятий и переводов с этого счета сумма должна быть отрицательной
	if t.Type == TransactionTypePayment ||
		t.Type == TransactionTypeWithdrawal ||
		(t.Type == TransactionTypeTransfer && t.FromAccountID > 0) {
		amount = -amount
	}

	return map[string]interface{}{
		"id":              t.ID,
		"type":            t.Type,
		"amount":          amount,
		"from_account_id": t.FromAccountID,
		"to_account_id":   t.ToAccountID,
		"description":     t.Description,
		"status":          t.Status,
		"created_at":      t.CreatedAt.Format(time.RFC3339),
		"updated_at":      t.UpdatedAt.Format(time.RFC3339),
	}
}
