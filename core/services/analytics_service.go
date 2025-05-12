package services

import (
	"FinanceGolang/core/dbaccess"
	"FinanceGolang/core/domain"
	"context"
	"time"
)

type AnalyticsService struct {
	transactionRepo dbaccess.TransactionRepository
	accountRepo     dbaccess.AccountRepository
	creditRepo      dbaccess.CreditRepository
}

func NewAnalyticsService(
	transactionRepo dbaccess.TransactionRepository,
	accountRepo dbaccess.AccountRepository,
	creditRepo dbaccess.CreditRepository,
) *AnalyticsService {
	return &AnalyticsService{
		transactionRepo: transactionRepo,
		accountRepo:     accountRepo,
		creditRepo:      creditRepo,
	}
}

// GetIncomeExpenseStats возвращает статистику доходов и расходов
func (s *AnalyticsService) GetIncomeExpenseStats(accountID uint, startDate, endDate time.Time) (*domain.IncomeExpenseStats, error) {
	transactions, err := s.transactionRepo.GetByAccountID(context.Background(), accountID)
	if err != nil {
		return nil, err
	}

	stats := &domain.IncomeExpenseStats{
		TotalIncome:  0,
		TotalExpense: 0,
		Categories:   make(map[string]float64),
	}

	for _, t := range transactions {
		if t.CreatedAt.After(startDate) && t.CreatedAt.Before(endDate) {
			if t.Type == domain.TransactionTypeDeposit {
				stats.TotalIncome += t.Amount
			} else {
				stats.TotalExpense += t.Amount
			}
			stats.Categories[string(t.Type)] += t.Amount
		}
	}

	return stats, nil
}

// GetBalanceForecast возвращает прогноз баланса на указанный период
func (s *AnalyticsService) GetBalanceForecast(accountID uint, months int) (*domain.BalanceForecast, error) {
	account, err := s.accountRepo.GetByID(context.Background(), accountID)
	if err != nil {
		return nil, err
	}

	forecast := &domain.BalanceForecast{
		CurrentBalance:  account.Balance,
		MonthlyForecast: make([]domain.MonthlyForecast, months),
	}

	now := time.Now()
	for i := 0; i < months; i++ {
		monthStart := time.Date(now.Year(), now.Month()+time.Month(i), 1, 0, 0, 0, 0, time.UTC)
		monthEnd := monthStart.AddDate(0, 1, -1)

		// Получаем статистику за предыдущий месяц для прогноза
		stats, err := s.GetIncomeExpenseStats(accountID, monthStart.AddDate(0, -1, 0), monthEnd.AddDate(0, -1, 0))
		if err != nil {
			return nil, err
		}

		// Получаем предстоящие платежи по кредитам
		credit, err := s.creditRepo.GetByAccountID(context.Background(), accountID)
		if err != nil {
			return nil, err
		}

		var creditPayments float64
		if credit != nil && credit.Status == domain.CreditStatusActive {
			// Рассчитываем платеж на основе данных кредита
			monthlyPayment := credit.Amount / float64(credit.Term)
			if credit.NextPayment.After(monthStart) && credit.NextPayment.Before(monthEnd) {
				creditPayments = monthlyPayment
			}
		}

		forecast.MonthlyForecast[i] = domain.MonthlyForecast{
			Month:   monthStart.Format("January 2006"),
			Income:  stats.TotalIncome,
			Expense: stats.TotalExpense + creditPayments,
			Balance: forecast.CurrentBalance + (stats.TotalIncome - stats.TotalExpense - creditPayments),
		}

		forecast.CurrentBalance = forecast.MonthlyForecast[i].Balance
	}

	return forecast, nil
}

// GetSpendingCategories возвращает статистику по категориям расходов
func (s *AnalyticsService) GetSpendingCategories(accountID uint, startDate, endDate time.Time) (map[string]float64, error) {
	transactions, err := s.transactionRepo.GetByDateRange(context.Background(), startDate, endDate)
	if err != nil {
		return nil, err
	}

	categories := make(map[string]float64)
	for _, t := range transactions {
		if t.FromAccountID == accountID && (t.Type == domain.TransactionTypeWithdrawal || t.Type == domain.TransactionTypeTransfer) {
			categories[string(t.Type)] += t.Amount
		}
	}

	return categories, nil
}
