package dbaccess

import (
	"context"
	"time"

	"FinanceGolang/core/domain"

	"gorm.io/gorm"
)

// AnalyticsRepository интерфейс репозитория аналитики
type AnalyticsRepository interface {
	GetDailyTransactions(ctx context.Context, date time.Time) ([]domain.Transaction, error)
	GetMonthlyTransactions(ctx context.Context, year int, month time.Month) ([]domain.Transaction, error)
	GetTransactionStats(ctx context.Context, startDate, endDate time.Time) (*domain.TransactionStats, error)
	GetUserStats(ctx context.Context) (*domain.UserStats, error)
	GetAccountStats(ctx context.Context) (*domain.AccountStats, error)
	GetCreditStats(ctx context.Context) (*domain.CreditStats, error)
	GetCardStats(ctx context.Context) (*domain.CardStats, error)
	GetRoleStats(ctx context.Context) (*domain.RoleStats, error)
}

// analyticsRepository реализация репозитория аналитики
type analyticsRepository struct {
	db *gorm.DB
}

// AnalyticsRepositoryInstance создает новый репозиторий аналитики
func AnalyticsRepositoryInstance(db *gorm.DB) AnalyticsRepository {
	return &analyticsRepository{db: db}
}

// GetDailyTransactions получает транзакции за день
func (r *analyticsRepository) GetDailyTransactions(ctx context.Context, date time.Time) ([]domain.Transaction, error) {
	var transactions []domain.Transaction
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	if err := r.db.Where("created_at BETWEEN ? AND ?", startOfDay, endOfDay).
		Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

// GetMonthlyTransactions получает транзакции за месяц
func (r *analyticsRepository) GetMonthlyTransactions(ctx context.Context, year int, month time.Month) ([]domain.Transaction, error) {
	var transactions []domain.Transaction
	startOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	if err := r.db.Where("created_at BETWEEN ? AND ?", startOfMonth, endOfMonth).
		Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

// GetTransactionStats получает статистику по транзакциям
func (r *analyticsRepository) GetTransactionStats(ctx context.Context, startDate, endDate time.Time) (*domain.TransactionStats, error) {
	var stats domain.TransactionStats

	// Общее количество транзакций
	if err := r.db.Model(&domain.Transaction{}).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Count(&stats.TotalTransactions).Error; err != nil {
		return nil, err
	}

	// Общая сумма транзакций
	if err := r.db.Model(&domain.Transaction{}).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&stats.TotalAmount).Error; err != nil {
		return nil, err
	}

	// Средняя сумма транзакции
	if stats.TotalTransactions > 0 {
		stats.AverageAmount = stats.TotalAmount / float64(stats.TotalTransactions)
	}

	// Количество транзакций по типам
	if err := r.db.Model(&domain.Transaction{}).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Select("type, COUNT(*) as count").
		Group("type").
		Scan(&stats.TransactionsByType).Error; err != nil {
		return nil, err
	}

	// Количество транзакций по статусам
	if err := r.db.Model(&domain.Transaction{}).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&stats.TransactionsByStatus).Error; err != nil {
		return nil, err
	}

	return &stats, nil
}

// GetUserStats получает статистику по пользователям
func (r *analyticsRepository) GetUserStats(ctx context.Context) (*domain.UserStats, error) {
	var stats domain.UserStats

	// Общее количество пользователей
	if err := r.db.Model(&domain.User{}).Count(&stats.TotalUsers).Error; err != nil {
		return nil, err
	}

	// Количество активных пользователей
	if err := r.db.Model(&domain.User{}).
		Where("is_active = ?", true).
		Count(&stats.ActiveUsers).Error; err != nil {
		return nil, err
	}

	// Количество пользователей по ролям
	if err := r.db.Model(&domain.User{}).
		Joins("JOIN user_roles ON user_roles.user_id = users.id").
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Select("roles.name, COUNT(*) as count").
		Group("roles.name").
		Scan(&stats.UsersByRole).Error; err != nil {
		return nil, err
	}

	return &stats, nil
}

// GetAccountStats получает статистику по счетам
func (r *analyticsRepository) GetAccountStats(ctx context.Context) (*domain.AccountStats, error) {
	var stats domain.AccountStats

	// Общее количество счетов
	if err := r.db.Model(&domain.Account{}).Count(&stats.TotalAccounts).Error; err != nil {
		return nil, err
	}

	// Общий баланс
	if err := r.db.Model(&domain.Account{}).
		Select("COALESCE(SUM(balance), 0)").
		Scan(&stats.TotalBalance).Error; err != nil {
		return nil, err
	}

	// Количество счетов по типам
	if err := r.db.Model(&domain.Account{}).
		Select("type, COUNT(*) as count").
		Group("type").
		Scan(&stats.AccountsByType).Error; err != nil {
		return nil, err
	}

	return &stats, nil
}

// GetCreditStats получает статистику по кредитам
func (r *analyticsRepository) GetCreditStats(ctx context.Context) (*domain.CreditStats, error) {
	var stats domain.CreditStats

	// Общее количество кредитов
	if err := r.db.Model(&domain.Credit{}).Count(&stats.TotalCredits).Error; err != nil {
		return nil, err
	}

	// Общая сумма кредитов
	if err := r.db.Model(&domain.Credit{}).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&stats.TotalAmount).Error; err != nil {
		return nil, err
	}

	// Общая сумма выплат
	if err := r.db.Model(&domain.Credit{}).
		Select("COALESCE(SUM(total_paid), 0)").
		Scan(&stats.TotalPaid).Error; err != nil {
		return nil, err
	}

	// Количество кредитов по статусам
	if err := r.db.Model(&domain.Credit{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&stats.CreditsByStatus).Error; err != nil {
		return nil, err
	}

	return &stats, nil
}

// GetCardStats получает статистику по картам
func (r *analyticsRepository) GetCardStats(ctx context.Context) (*domain.CardStats, error) {
	var stats domain.CardStats

	// Общее количество карт
	if err := r.db.Model(&domain.Card{}).Count(&stats.TotalCards).Error; err != nil {
		return nil, err
	}

	// Количество активных карт
	if err := r.db.Model(&domain.Card{}).
		Where("is_active = ?", true).
		Count(&stats.ActiveCards).Error; err != nil {
		return nil, err
	}

	// Количество просроченных карт
	now := time.Now().Format("01/06")
	if err := r.db.Model(&domain.Card{}).
		Where("expiry_date < ?", now).
		Count(&stats.ExpiredCards).Error; err != nil {
		return nil, err
	}

	return &stats, nil
}

// GetRoleStats получает статистику по ролям
func (r *analyticsRepository) GetRoleStats(ctx context.Context) (*domain.RoleStats, error) {
	var stats domain.RoleStats

	// Общее количество ролей
	if err := r.db.Model(&domain.Role{}).Count(&stats.TotalRoles).Error; err != nil {
		return nil, err
	}

	// Количество активных ролей
	if err := r.db.Model(&domain.Role{}).
		Where("is_active = ?", true).
		Count(&stats.ActiveRoles).Error; err != nil {
		return nil, err
	}

	// Количество пользователей по ролям
	if err := r.db.Model(&domain.Role{}).
		Select("roles.name, COUNT(DISTINCT user_roles.user_id) as count").
		Joins("LEFT JOIN user_roles ON user_roles.role_id = roles.id").
		Group("roles.name").
		Scan(&stats.UsersByRole).Error; err != nil {
		return nil, err
	}

	return &stats, nil
}
