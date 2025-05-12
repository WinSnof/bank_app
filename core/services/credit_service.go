package services

import (
	"FinanceGolang/core/dbaccess"
	"FinanceGolang/core/domain"
	"context"
	"errors"
	"fmt"

	// "strconv"
	"time"
)

type CreditService interface {
	CreateCredit(userID uint, accountID uint, amount float64, termMonths int, description string) (*domain.Credit, error)
	GetCreditByID(id uint) (*domain.Credit, error)
	GetUserCredits(userID uint) ([]domain.Credit, error)
	GetPaymentSchedule(creditID uint) ([]domain.PaymentSchedule, error)
	ProcessPayment(creditID uint, paymentNumber int) error
	ProcessOverduePayments() error
}

type creditService struct {
	creditRepo      dbaccess.CreditRepository
	accountRepo     dbaccess.AccountRepository
	transactionRepo dbaccess.TransactionRepository
	keyRateService  *ExternalService
}

func CreditServiceInstance(
	creditRepo dbaccess.CreditRepository,
	accountRepo dbaccess.AccountRepository,
	transactionRepo dbaccess.TransactionRepository,
	keyRateService *ExternalService,
) CreditService {
	return &creditService{
		creditRepo:      creditRepo,
		accountRepo:     accountRepo,
		transactionRepo: transactionRepo,
		keyRateService:  keyRateService,
	}
}

func (s *creditService) CreateCredit(userID uint, accountID uint, amount float64, termMonths int, description string) (*domain.Credit, error) {
	// Проверяем, что счет принадлежит пользователю
	account, err := s.accountRepo.GetByID(context.Background(), accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %v", err)
	}
	if account.UserID != userID {
		return nil, errors.New("account does not belong to the user")
	}

	// Получаем текущую ключевую ставку
	keyRate, err := s.keyRateService.GetKeyRate()
	if err != nil {
		return nil, fmt.Errorf("failed to get key rate: %v", err)
	}

	// Рассчитываем процентную ставку (ключевая ставка + 5%)
	interestRate := keyRate + 5.0

	// Текущее время для инициализации дат
	now := time.Now()

	// Создаем кредит
	credit := &domain.Credit{
		UserID:        userID,
		AccountID:     accountID,
		Amount:        amount,
		Term:          termMonths,
		InterestRate:  interestRate,
		Status:        domain.CreditStatusActive,
		StartDate:     now,
		EndDate:       now.AddDate(0, termMonths, 0),
		PaymentDay:    now.Day(),
		NextPayment:   now.AddDate(0, 1, 0),
		RemainingDebt: amount,
		LastPayment:   now, // Инициализируем LastPayment текущей датой
	}

	// Сохраняем кредит
	if err := s.creditRepo.Create(context.Background(), credit); err != nil {
		return nil, fmt.Errorf("failed to create credit: %v", err)
	}

	// Зачисляем сумму кредита на счет пользователя
	account.Balance += amount
	if err := s.accountRepo.Update(context.Background(), account); err != nil {
		return nil, fmt.Errorf("failed to update account balance: %v", err)
	}

	// Создаем транзакцию о зачислении кредита
	transaction := &domain.Transaction{
		Type:        domain.TransactionTypeCredit,
		ToAccountID: accountID,
		Amount:      amount,
		Description: fmt.Sprintf("Зачисление по кредиту #%d: %s", credit.ID, description),
		Status:      domain.TransactionStatusCompleted,
	}
	if err := s.transactionRepo.Create(context.Background(), transaction); err != nil {
		return nil, fmt.Errorf("failed to create transaction: %v", err)
	}

	return credit, nil
}

func (s *creditService) GetCreditByID(id uint) (*domain.Credit, error) {
	return s.creditRepo.GetByID(context.Background(), id)
}

func (s *creditService) GetUserCredits(userID uint) ([]domain.Credit, error) {
	return s.creditRepo.GetCreditsByUserID(context.Background(), userID)
}

func (s *creditService) GetPaymentSchedule(creditID uint) ([]domain.PaymentSchedule, error) {
	credit, err := s.creditRepo.GetByID(context.Background(), creditID)
	if err != nil {
		return nil, err
	}

	// Проверяем, если LastPayment пустая, установим её равной StartDate
	if credit.LastPayment.IsZero() {
		credit.LastPayment = credit.StartDate
	}

	var schedule []domain.PaymentSchedule
	remainingAmount := credit.Amount
	monthlyRate := credit.InterestRate / 12 / 100

	for i := 1; i <= credit.Term; i++ {
		interestAmount := remainingAmount * monthlyRate
		principalAmount := credit.CalculateMonthlyPayment() - interestAmount
		remainingAmount -= principalAmount

		// Создаем копию объекта кредита для включения в платеж
		creditCopy := *credit

		schedule = append(schedule, domain.PaymentSchedule{
			CreditID:      credit.ID,
			PaymentNumber: i,
			DueDate:       credit.StartDate.AddDate(0, i, 0),
			Amount:        credit.CalculateMonthlyPayment(),
			Interest:      interestAmount,
			Principal:     principalAmount,
			TotalAmount:   credit.CalculateMonthlyPayment(),
			Status:        domain.PaymentStatusPending,
			Credit:        creditCopy,
		})
	}

	return schedule, nil
}

func (s *creditService) ProcessPayment(creditID uint, paymentNumber int) error {
	// Получаем кредит
	credit, err := s.creditRepo.GetByID(context.Background(), creditID)
	if err != nil {
		return fmt.Errorf("failed to get credit: %v", err)
	}

	// Проверяем, не был ли уже оплачен этот платеж
	transactions, err := s.transactionRepo.GetByType(context.Background(), domain.TransactionTypePayment)
	if err != nil {
		return fmt.Errorf("failed to get payment transactions: %v", err)
	}

	for _, t := range transactions {
		if t.FromAccountID == credit.AccountID && t.Status == domain.TransactionStatusCompleted {
			// Проверяем номер платежа в описании
			var paidPaymentNumber int
			_, err := fmt.Sscanf(t.Description, "Платеж по кредиту #%d, платеж #%d", &creditID, &paidPaymentNumber)
			if err == nil && paidPaymentNumber == paymentNumber {
				return fmt.Errorf("payment #%d already processed", paymentNumber)
			}
		}
	}

	// Получаем график платежей
	schedule, err := s.GetPaymentSchedule(creditID)
	if err != nil {
		return fmt.Errorf("failed to get payment schedule: %v", err)
	}

	// Находим нужный платеж
	var payment *domain.PaymentSchedule
	for i := range schedule {
		if schedule[i].PaymentNumber == paymentNumber {
			payment = &schedule[i]
			break
		}
	}
	if payment == nil {
		return errors.New("payment not found")
	}

	// Проверяем баланс счета
	account, err := s.accountRepo.GetByID(context.Background(), credit.AccountID)
	if err != nil {
		return fmt.Errorf("failed to get account: %v", err)
	}

	if account.Balance < payment.TotalAmount {
		// Если средств недостаточно, начисляем штраф
		penalty := payment.TotalAmount * 0.1
		payment.TotalAmount += penalty
		credit.Status = domain.CreditStatusOverdue
		if err := s.creditRepo.Update(context.Background(), credit); err != nil {
			return fmt.Errorf("failed to update credit status: %v", err)
		}
		return errors.New("insufficient funds, penalty applied")
	}

	// Списываем средства со счета
	account.Balance -= payment.TotalAmount
	if err := s.accountRepo.Update(context.Background(), account); err != nil {
		return fmt.Errorf("failed to update account balance: %v", err)
	}

	// Создаем транзакцию о платеже
	transaction := &domain.Transaction{
		Type:          domain.TransactionTypePayment,
		FromAccountID: credit.AccountID,
		Amount:        payment.TotalAmount,
		Description:   fmt.Sprintf("Платеж по кредиту #%d, платеж #%d", credit.ID, paymentNumber),
		Status:        domain.TransactionStatusCompleted,
	}
	if err := s.transactionRepo.Create(context.Background(), transaction); err != nil {
		return fmt.Errorf("failed to create transaction: %v", err)
	}

	// Обновляем статус платежа
	payment.Status = domain.PaymentStatusPaid
	now := time.Now()
	payment.PaidAt = &now

	// Обновляем кредит
	credit.TotalPaid += payment.TotalAmount
	credit.RemainingDebt = credit.CalculateRemainingDebt()
	credit.LastPayment = now
	credit.NextPayment = credit.CalculateNextPaymentDate()

	// Если это последний платеж, закрываем кредит
	if paymentNumber == credit.Term {
		credit.Status = domain.CreditStatusPaid
	}

	if err := s.creditRepo.Update(context.Background(), credit); err != nil {
		return fmt.Errorf("failed to update credit: %v", err)
	}

	return nil
}

func (s *creditService) ProcessOverduePayments() error {
	// Получаем просроченные кредиты
	overdueCredits, err := s.creditRepo.GetOverdueCredits(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get overdue credits: %v", err)
	}

	for _, credit := range overdueCredits {
		// Получаем график платежей
		schedule, err := s.GetPaymentSchedule(credit.ID)
		if err != nil {
			return fmt.Errorf("failed to get payment schedule for credit %d: %v", credit.ID, err)
		}

		// Находим просроченный платеж
		for _, payment := range schedule {
			if payment.Status == domain.PaymentStatusPending && payment.DueDate.Before(time.Now()) {
				// Пытаемся обработать платеж
				if err := s.ProcessPayment(credit.ID, payment.PaymentNumber); err != nil {
					// Если не удалось обработать платеж, обновляем статус кредита
					credit.Status = domain.CreditStatusOverdue
					if err := s.creditRepo.Update(context.Background(), &credit); err != nil {
						return fmt.Errorf("failed to update credit status: %v", err)
					}
				}
				break
			}
		}
	}

	return nil
}
