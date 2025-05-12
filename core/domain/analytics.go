package domain

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type AnalyticsPeriod string

const (
	Daily   AnalyticsPeriod = "daily"
	Weekly  AnalyticsPeriod = "weekly"
	Monthly AnalyticsPeriod = "monthly"
	Yearly  AnalyticsPeriod = "yearly"
)

type TransactionCategory string

const (
	Income   TransactionCategory = "income"
	Expense  TransactionCategory = "expense"
	Transfer TransactionCategory = "transfer"
)

type Analytics struct {
	ID        uint            `json:"id" gorm:"primaryKey"`
	UserID    uint            `json:"user_id" gorm:"index"`
	AccountID uint            `json:"account_id" gorm:"index"`
	Period    AnalyticsPeriod `json:"period"`
	StartDate time.Time       `json:"start_date"`
	EndDate   time.Time       `json:"end_date"`

	// Общая статистика
	TotalIncome  float64 `json:"total_income"`
	TotalExpense float64 `json:"total_expense"`
	NetIncome    float64 `json:"net_income"`

	// Статистика по категориям
	Categories map[TransactionCategory]float64 `json:"categories" gorm:"-"`

	// Кредитная нагрузка
	CreditPayments float64 `json:"credit_payments"`
	CreditLoad     float64 `json:"credit_load"` // Процент от дохода

	// Прогноз
	BalanceForecast []BalanceForecast `json:"balance_forecast" gorm:"-"`
}

type IncomeExpenseStats struct {
	TotalIncome  float64
	TotalExpense float64
	Categories   map[string]float64
}

type BalanceForecast struct {
	CurrentBalance      float64           `json:"current_balance"`
	MonthlyForecast     []MonthlyForecast `json:"monthly_forecast" gorm:"-"`             // Исключаем из GORM
	MonthlyForecastJSON string            `json:"-" gorm:"column:monthly_forecast_json"` // Для хранения JSON
}

// Scan реализует интерфейс sql.Scanner
func (bf *BalanceForecast) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("тип не []byte")
	}

	var temp struct {
		CurrentBalance  float64           `json:"current_balance"`
		MonthlyForecast []MonthlyForecast `json:"monthly_forecast"`
	}
	if err := json.Unmarshal(bytes, &temp); err != nil {
		return err
	}

	bf.CurrentBalance = temp.CurrentBalance
	bf.MonthlyForecast = temp.MonthlyForecast
	bf.MonthlyForecastJSON = string(bytes)
	return nil
}

// Value реализует интерфейс driver.Valuer
func (bf BalanceForecast) Value() (driver.Value, error) {
	if len(bf.MonthlyForecastJSON) > 0 {
		return []byte(bf.MonthlyForecastJSON), nil
	}

	temp := struct {
		CurrentBalance  float64           `json:"current_balance"`
		MonthlyForecast []MonthlyForecast `json:"monthly_forecast"`
	}{
		CurrentBalance:  bf.CurrentBalance,
		MonthlyForecast: bf.MonthlyForecast,
	}

	bytes, err := json.Marshal(temp)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

type MonthlyForecast struct {
	Month   string  `json:"month"`
	Income  float64 `json:"income"`
	Expense float64 `json:"expense"`
	Balance float64 `json:"balance"`
}

type AnalyticsRequest struct {
	AccountID uint            `json:"account_id" binding:"required"`
	Period    AnalyticsPeriod `json:"period" binding:"required"`
	StartDate time.Time       `json:"start_date" binding:"required"`
	EndDate   time.Time       `json:"end_date" binding:"required"`
}
