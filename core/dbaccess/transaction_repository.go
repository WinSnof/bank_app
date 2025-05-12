package dbaccess

import (
	"context"
	"time"

	"FinanceGolang/core/domain"
	// "errors"
	"gorm.io/gorm"
)

// TransactionRepository интерфейс репозитория транзакций
type TransactionRepository interface {
	Repository[domain.Transaction]
	GetByAccountID(ctx context.Context, accountID uint) ([]domain.Transaction, error)
	GetByCardID(ctx context.Context, cardID uint) ([]domain.Transaction, error)
	GetByType(ctx context.Context, transactionType domain.TransactionType) ([]domain.Transaction, error)
	GetByStatus(ctx context.Context, status domain.TransactionStatus) ([]domain.Transaction, error)
	GetByDateRange(ctx context.Context, startDate, endDate time.Time) ([]domain.Transaction, error)
	GetDailyTransactions(ctx context.Context, date time.Time) ([]domain.Transaction, error)
	GetMonthlyTransactions(ctx context.Context, year int, month time.Month) ([]domain.Transaction, error)
	UpdateStatus(ctx context.Context, id uint, status domain.TransactionStatus) error
	GetTransactionsByAmountRange(ctx context.Context, minAmount, maxAmount float64) ([]domain.Transaction, error)
}

// transactionRepository реализация репозитория транзакций
type transactionRepository struct {
	BaseRepository[domain.Transaction]
}

// TransactionRepositoryInstance создает новый репозиторий транзакций
func TransactionRepositoryInstance(db *gorm.DB) TransactionRepository {
	return &transactionRepository{
		BaseRepository: *NewBaseRepository[domain.Transaction](db),
	}
}

// Create создает новую транзакцию
func (r *transactionRepository) Create(ctx context.Context, transaction *domain.Transaction) error {
	return r.WithTransaction(ctx, func(tx *gorm.DB) error {
		if err := transaction.Validate(); err != nil {
			return ErrInvalidData
		}

		if err := tx.Create(transaction).Error; err != nil {
			return r.HandleError(err)
		}
		return nil
	})
}

// GetByID получает транзакцию по ID
func (r *transactionRepository) GetByID(ctx context.Context, id uint) (*domain.Transaction, error) {
	var transaction domain.Transaction
	if err := r.db.First(&transaction, id).Error; err != nil {
		return nil, r.HandleError(err)
	}
	return &transaction, nil
}

// GetByAccountID получает транзакции по ID счета
func (r *transactionRepository) GetByAccountID(ctx context.Context, accountID uint) ([]domain.Transaction, error) {
	var transactions []domain.Transaction
	if err := r.db.Where("from_account_id = ? OR to_account_id = ?", accountID, accountID).
		Find(&transactions).Error; err != nil {
		return nil, r.HandleError(err)
	}
	return transactions, nil
}

// GetByCardID получает транзакции по ID карты
func (r *transactionRepository) GetByCardID(ctx context.Context, cardID uint) ([]domain.Transaction, error) {
	var transactions []domain.Transaction
	if err := r.db.Where("card_id = ?", cardID).Find(&transactions).Error; err != nil {
		return nil, r.HandleError(err)
	}
	return transactions, nil
}

// GetByType получает транзакции по типу
func (r *transactionRepository) GetByType(ctx context.Context, transactionType domain.TransactionType) ([]domain.Transaction, error) {
	var transactions []domain.Transaction
	if err := r.db.Where("type = ?", transactionType).Find(&transactions).Error; err != nil {
		return nil, r.HandleError(err)
	}
	return transactions, nil
}

// GetByStatus получает транзакции по статусу
func (r *transactionRepository) GetByStatus(ctx context.Context, status domain.TransactionStatus) ([]domain.Transaction, error) {
	var transactions []domain.Transaction
	if err := r.db.Where("status = ?", status).Find(&transactions).Error; err != nil {
		return nil, r.HandleError(err)
	}
	return transactions, nil
}

// GetByDateRange получает транзакции в указанном диапазоне дат
func (r *transactionRepository) GetByDateRange(ctx context.Context, startDate, endDate time.Time) ([]domain.Transaction, error) {
	var transactions []domain.Transaction
	if err := r.db.Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Find(&transactions).Error; err != nil {
		return nil, r.HandleError(err)
	}
	return transactions, nil
}

// GetDailyTransactions получает транзакции за день
func (r *transactionRepository) GetDailyTransactions(ctx context.Context, date time.Time) ([]domain.Transaction, error) {
	var transactions []domain.Transaction
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	if err := r.db.Where("created_at BETWEEN ? AND ?", startOfDay, endOfDay).
		Find(&transactions).Error; err != nil {
		return nil, r.HandleError(err)
	}
	return transactions, nil
}

// GetMonthlyTransactions получает транзакции за месяц
func (r *transactionRepository) GetMonthlyTransactions(ctx context.Context, year int, month time.Month) ([]domain.Transaction, error) {
	var transactions []domain.Transaction
	startOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	if err := r.db.Where("created_at BETWEEN ? AND ?", startOfMonth, endOfMonth).
		Find(&transactions).Error; err != nil {
		return nil, r.HandleError(err)
	}
	return transactions, nil
}

// Update обновляет транзакцию
func (r *transactionRepository) Update(ctx context.Context, transaction *domain.Transaction) error {
	return r.WithTransaction(ctx, func(tx *gorm.DB) error {
		if err := transaction.Validate(); err != nil {
			return ErrInvalidData
		}

		if err := tx.Save(transaction).Error; err != nil {
			return r.HandleError(err)
		}
		return nil
	})
}

// UpdateStatus обновляет статус транзакции
func (r *transactionRepository) UpdateStatus(ctx context.Context, id uint, status domain.TransactionStatus) error {
	return r.WithTransaction(ctx, func(tx *gorm.DB) error {
		if err := tx.Model(&domain.Transaction{}).Where("id = ?", id).Update("status", status).Error; err != nil {
			return r.HandleError(err)
		}
		return nil
	})
}

// Delete удаляет транзакцию
func (r *transactionRepository) Delete(ctx context.Context, id uint) error {
	return r.WithTransaction(ctx, func(tx *gorm.DB) error {
		if err := tx.Delete(&domain.Transaction{}, id).Error; err != nil {
			return r.HandleError(err)
		}
		return nil
	})
}

// List получает список транзакций
func (r *transactionRepository) List(ctx context.Context, offset, limit int) ([]domain.Transaction, error) {
	var transactions []domain.Transaction
	if err := r.db.Offset(offset).Limit(limit).Find(&transactions).Error; err != nil {
		return nil, r.HandleError(err)
	}
	return transactions, nil
}

// Count возвращает количество транзакций
func (r *transactionRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.Model(&domain.Transaction{}).Count(&count).Error; err != nil {
		return 0, r.HandleError(err)
	}
	return count, nil
}

// GetTransactionsByAmountRange получает транзакции в указанном диапазоне сумм
func (r *transactionRepository) GetTransactionsByAmountRange(ctx context.Context, minAmount, maxAmount float64) ([]domain.Transaction, error) {
	var transactions []domain.Transaction
	if err := r.db.Where("amount BETWEEN ? AND ?", minAmount, maxAmount).
		Find(&transactions).Error; err != nil {
		return nil, r.HandleError(err)
	}
	return transactions, nil
}
