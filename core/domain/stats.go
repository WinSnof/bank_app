package domain

// TransactionStats статистика по транзакциям
type TransactionStats struct {
	TotalTransactions    int64
	TotalAmount          float64
	AverageAmount        float64
	TransactionsByType   map[string]int64
	TransactionsByStatus map[string]int64
}

// UserStats статистика по пользователям
type UserStats struct {
	TotalUsers  int64
	ActiveUsers int64
	UsersByRole map[string]int64
}

// AccountStats статистика по счетам
type AccountStats struct {
	TotalAccounts  int64
	TotalBalance   float64
	AccountsByType map[string]int64
}

// CreditStats статистика по кредитам
type CreditStats struct {
	TotalCredits    int64
	TotalAmount     float64
	TotalPaid       float64
	CreditsByStatus map[string]int64
}

// CardStats статистика по картам
type CardStats struct {
	TotalCards   int64
	ActiveCards  int64
	ExpiredCards int64
}

// RoleStats статистика по ролям
type RoleStats struct {
	TotalRoles  int64
	ActiveRoles int64
	UsersByRole map[string]int64
}
