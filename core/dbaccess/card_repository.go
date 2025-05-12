package dbaccess

import (
	"context"
	"time"

	"FinanceGolang/core/domain"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// CardRepository интерфейс репозитория карт
type CardRepository interface {
	Repository[domain.Card]
	GetByNumber(ctx context.Context, number string) (*domain.Card, error)
	GetByUserID(ctx context.Context, userID uint) ([]domain.Card, error)
	GetByAccountID(ctx context.Context, accountID uint) ([]domain.Card, error)
	GetExpiredCards(ctx context.Context) ([]domain.Card, error)
	GetActiveCards(ctx context.Context) ([]domain.Card, error)
	UpdateStatus(ctx context.Context, id uint, isActive bool) error
	GetDailyUsage(ctx context.Context, id uint, date time.Time) (float64, error)
	GetMonthlyUsage(ctx context.Context, id uint, year int, month time.Month) (float64, error)
}

// cardRepository реализация репозитория карт
type cardRepository struct {
	BaseRepository[domain.Card]
}

// CardRepositoryInstance создает новый репозиторий карт
func CardRepositoryInstance(db *gorm.DB) CardRepository {
	return &cardRepository{
		BaseRepository: *NewBaseRepository[domain.Card](db),
	}
}

// Create создает новую карту
// Валидация данных происходит в сервисном слое, так как:
// 1. Требуется доступ к другим сервисам (проверка принадлежности счета)
// 2. Данные уже зашифрованы и не могут быть валидированы на уровне репозитория
// 3. Валидация включает сложную бизнес-логику
func (r *cardRepository) Create(ctx context.Context, card *domain.Card) error {
	return r.WithTransaction(ctx, func(tx *gorm.DB) error {
		if err := tx.Create(card).Error; err != nil {
			// Логируем детали ошибки
			logrus.WithFields(logrus.Fields{
				"error": err.Error(),
				"card": map[string]interface{}{
					"id":         card.ID,
					"user_id":    card.UserID,
					"account_id": card.AccountID,
					"is_active":  card.IsActive,
				},
			}).Error("Ошибка при создании карты")
			return r.HandleError(err)
		}
		return nil
	})
}

// GetByID получает карту по ID
func (r *cardRepository) GetByID(ctx context.Context, id uint) (*domain.Card, error) {
	var card domain.Card
	if err := r.db.First(&card, id).Error; err != nil {
		return nil, r.HandleError(err)
	}
	return &card, nil
}

// GetByNumber получает карту по номеру
func (r *cardRepository) GetByNumber(ctx context.Context, number string) (*domain.Card, error) {
	var card domain.Card
	if err := r.db.Where("number = ?", number).First(&card).Error; err != nil {
		return nil, r.HandleError(err)
	}
	return &card, nil
}

// GetByUserID получает карты пользователя
func (r *cardRepository) GetByUserID(ctx context.Context, userID uint) ([]domain.Card, error) {
	var cards []domain.Card
	if err := r.db.Where("user_id = ?", userID).Find(&cards).Error; err != nil {
		return nil, r.HandleError(err)
	}
	return cards, nil
}

// GetByAccountID получает карты по ID счета
func (r *cardRepository) GetByAccountID(ctx context.Context, accountID uint) ([]domain.Card, error) {
	var cards []domain.Card
	if err := r.db.Where("account_id = ?", accountID).Find(&cards).Error; err != nil {
		return nil, r.HandleError(err)
	}
	return cards, nil
}

// GetExpiredCards получает просроченные карты
func (r *cardRepository) GetExpiredCards(ctx context.Context) ([]domain.Card, error) {
	var cards []domain.Card
	now := time.Now().Format("01/06")
	if err := r.db.Where("expiry_date < ?", now).Find(&cards).Error; err != nil {
		return nil, r.HandleError(err)
	}
	return cards, nil
}

// GetActiveCards получает активные карты
func (r *cardRepository) GetActiveCards(ctx context.Context) ([]domain.Card, error) {
	var cards []domain.Card
	if err := r.db.Where("is_active = ?", true).Find(&cards).Error; err != nil {
		return nil, r.HandleError(err)
	}
	return cards, nil
}

// Update обновляет карту
func (r *cardRepository) Update(ctx context.Context, card *domain.Card) error {
	return r.WithTransaction(ctx, func(tx *gorm.DB) error {
		if err := card.Validate(); err != nil {
			return ErrInvalidData
		}

		if err := tx.Save(card).Error; err != nil {
			return r.HandleError(err)
		}
		return nil
	})
}

// UpdateStatus обновляет статус карты
func (r *cardRepository) UpdateStatus(ctx context.Context, id uint, isActive bool) error {
	return r.WithTransaction(ctx, func(tx *gorm.DB) error {
		if err := tx.Model(&domain.Card{}).Where("id = ?", id).Update("is_active", isActive).Error; err != nil {
			return r.HandleError(err)
		}
		return nil
	})
}

// Delete удаляет карту
func (r *cardRepository) Delete(ctx context.Context, id uint) error {
	return r.WithTransaction(ctx, func(tx *gorm.DB) error {
		if err := tx.Delete(&domain.Card{}, id).Error; err != nil {
			return r.HandleError(err)
		}
		return nil
	})
}

// List получает список карт
func (r *cardRepository) List(ctx context.Context, offset, limit int) ([]domain.Card, error) {
	var cards []domain.Card
	if err := r.db.Offset(offset).Limit(limit).Find(&cards).Error; err != nil {
		return nil, r.HandleError(err)
	}
	return cards, nil
}

// Count возвращает количество карт
func (r *cardRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.Model(&domain.Card{}).Count(&count).Error; err != nil {
		return 0, r.HandleError(err)
	}
	return count, nil
}

// GetDailyUsage получает дневной лимит использования карты
func (r *cardRepository) GetDailyUsage(ctx context.Context, id uint, date time.Time) (float64, error) {
	var total float64
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	if err := r.db.Model(&domain.Transaction{}).
		Where("card_id = ? AND created_at BETWEEN ? AND ?", id, startOfDay, endOfDay).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error; err != nil {
		return 0, r.HandleError(err)
	}
	return total, nil
}

// GetMonthlyUsage получает месячный лимит использования карты
func (r *cardRepository) GetMonthlyUsage(ctx context.Context, id uint, year int, month time.Month) (float64, error) {
	var total float64
	startOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	if err := r.db.Model(&domain.Transaction{}).
		Where("card_id = ? AND created_at BETWEEN ? AND ?", id, startOfMonth, endOfMonth).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error; err != nil {
		return 0, r.HandleError(err)
	}
	return total, nil
}
