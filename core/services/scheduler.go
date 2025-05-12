package services

import (
	"context"
	"fmt"
	"time"

	"FinanceGolang/core/dbaccess"
	"FinanceGolang/core/dbcore"
	"FinanceGolang/core/domain"
)

type Scheduler struct {
	creditRepo      dbaccess.CreditRepository
	accountRepo     dbaccess.AccountRepository
	transactionRepo dbaccess.TransactionRepository
	userRepo        dbaccess.UserRepository
	keyRateService  *ExternalService
}

func NewScheduler(
	creditRepo dbaccess.CreditRepository,
	accountRepo dbaccess.AccountRepository,
	transactionRepo dbaccess.TransactionRepository,
	keyRateService *ExternalService,
) *Scheduler {
	return &Scheduler{
		creditRepo:      creditRepo,
		accountRepo:     accountRepo,
		transactionRepo: transactionRepo,
		userRepo:        dbaccess.UserRepositoryInstance(dbcore.DB),
		keyRateService:  keyRateService,
	}
}

// Start запускает шедулер
func (s *Scheduler) Start() {
	// Проверка платежей каждые 12 часов
	go s.checkPayments()
}

// checkPayments проверяет и обрабатывает платежи по кредитам
func (s *Scheduler) checkPayments() {
	ticker := time.NewTicker(12 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		credits, err := s.creditRepo.GetActiveCredits(context.Background())
		if err != nil {
			fmt.Printf("Ошибка при получении кредитов: %v\n", err)
			continue
		}

		for _, credit := range credits {
			if credit.Status == domain.CreditStatusActive {
				s.processCreditPayment(&credit)
			}
		}
	}
}

// processCreditPayment обрабатывает платеж по кредиту
func (s *Scheduler) processCreditPayment(credit *domain.Credit) {
	// Получаем график платежей
	schedule, err := s.creditRepo.GetCreditsByDateRange(context.Background(), credit.StartDate, credit.EndDate)
	if err != nil {
		fmt.Printf("Ошибка при получении графика платежей: %v\n", err)
		return
	}

	now := time.Now()
	for _, payment := range schedule {
		// Если платеж просрочен и не оплачен
		if payment.NextPayment.Before(now) && payment.Status == domain.CreditStatusActive {
			// Проверяем баланс счета
			account, err := s.accountRepo.GetByID(context.Background(), credit.AccountID)
			if err != nil {
				fmt.Printf("Ошибка при получении счета: %v\n", err)
				continue
			}

			monthlyPayment := payment.CalculateMonthlyPayment()

			// Если на счету достаточно средств
			if account.Balance >= monthlyPayment {
				// Списание средств
				account.Balance -= monthlyPayment
				if err := s.accountRepo.Update(context.Background(), account); err != nil {
					fmt.Printf("Ошибка при списании средств: %v\n", err)
					continue
				}

				// Обновление статуса платежа
				payment.Status = domain.CreditStatusPaid
				now := time.Now()
				payment.LastPayment = now

				// Создаем транзакцию о платеже
				transaction := &domain.Transaction{
					Type:          domain.TransactionTypePayment,
					FromAccountID: credit.AccountID,
					Amount:        monthlyPayment,
					Description:   fmt.Sprintf("Платеж по кредиту #%d", credit.ID),
					Status:        domain.TransactionStatusCompleted,
				}
				if err := s.transactionRepo.Create(context.Background(), transaction); err != nil {
					fmt.Printf("Ошибка при создании транзакции: %v\n", err)
					continue
				}

				// Отправка уведомления
				user, err := s.userRepo.GetByID(context.Background(), account.UserID)
				if err != nil {
					fmt.Printf("Ошибка при получении пользователя: %v\n", err)
					continue
				}

				if user == nil {
					fmt.Printf("Пользователь не найден\n")
					continue
				}

				if err := s.keyRateService.SendPaymentNotification(
					user.Email,
					"Платеж по кредиту",
					monthlyPayment,
				); err != nil {
					fmt.Printf("Ошибка при отправке уведомления: %v\n", err)
				}
			} else {
				// Начисление штрафа за просрочку
				penalty := monthlyPayment * 0.1
				payment.OverdueAmount += penalty
				payment.Status = domain.CreditStatusOverdue

				// Создаем транзакцию о штрафе
				transaction := &domain.Transaction{
					Type:          domain.TransactionTypePayment,
					FromAccountID: credit.AccountID,
					Amount:        penalty,
					Description:   fmt.Sprintf("Штраф за просрочку платежа по кредиту #%d", credit.ID),
					Status:        domain.TransactionStatusCompleted,
				}
				if err := s.transactionRepo.Create(context.Background(), transaction); err != nil {
					fmt.Printf("Ошибка при создании транзакции: %v\n", err)
					continue
				}

				// Отправка уведомления о просрочке
				user, err := s.userRepo.GetByID(context.Background(), account.UserID)
				if err != nil {
					fmt.Printf("Ошибка при получении пользователя: %v\n", err)
					continue
				}

				if user == nil {
					fmt.Printf("Пользователь не найден\n")
					continue
				}

				if err := s.keyRateService.SendPaymentNotification(
					user.Email,
					"Просрочка платежа по кредиту",
					monthlyPayment+penalty,
				); err != nil {
					fmt.Printf("Ошибка при отправке уведомления: %v\n", err)
				}
			}
		}
	}
}

// ProcessPayment обрабатывает платеж по кредиту
func (s *Scheduler) ProcessPayment(creditID uint, paymentNumber int) error {
	credit, err := s.creditRepo.GetByID(context.Background(), creditID)
	if err != nil {
		return fmt.Errorf("failed to get credit: %v", err)
	}

	account, err := s.accountRepo.GetByID(context.Background(), credit.AccountID)
	if err != nil {
		return fmt.Errorf("failed to get account: %v", err)
	}

	// Получаем график платежей
	schedule, err := s.creditRepo.GetPaymentSchedule(context.Background(), creditID)
	if err != nil {
		return fmt.Errorf("failed to get payment schedule: %v", err)
	}

	// Находим нужный платеж
	var payment *domain.PaymentSchedule
	for _, p := range schedule {
		if p.PaymentNumber == paymentNumber {
			payment = &p
			break
		}
	}

	if payment == nil {
		return fmt.Errorf("payment not found")
	}

	// Проверяем достаточно ли средств
	if account.Balance < payment.Amount {
		return fmt.Errorf("insufficient funds")
	}

	// Списываем платеж
	account.Balance -= payment.Amount
	if err := s.accountRepo.Update(context.Background(), account); err != nil {
		return fmt.Errorf("failed to update account: %v", err)
	}

	// Обновляем статус платежа
	payment.Status = "COMPLETED"
	if err := s.creditRepo.UpdatePaymentSchedule(context.Background(), payment); err != nil {
		return fmt.Errorf("failed to update payment schedule: %v", err)
	}

	// Создаем транзакцию
	transaction := &domain.Transaction{
		FromAccountID: credit.AccountID,
		Amount:        payment.Amount,
		Type:          "CREDIT_PAYMENT",
		Status:        "COMPLETED",
		Description:   fmt.Sprintf("Платеж по кредиту #%d", credit.ID),
	}

	if err := s.transactionRepo.Create(context.Background(), transaction); err != nil {
		return fmt.Errorf("failed to create transaction: %v", err)
	}

	// Получаем пользователя для уведомления
	user, err := s.userRepo.GetByID(context.Background(), account.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user: %v", err)
	}

	if user == nil {
		return fmt.Errorf("user not found")
	}

	// Отправляем уведомление
	if err := s.keyRateService.SendPaymentNotification(
		user.Email,
		"Платеж по кредиту",
		payment.Amount,
	); err != nil {
		return fmt.Errorf("failed to send notification: %v", err)
	}

	return nil
}

// sendPaymentOverdueNotification отправляет уведомление о просрочке платежа
func (s *Scheduler) sendPaymentOverdueNotification(email string, creditID uint, amount float64) error {
	return s.keyRateService.SendPaymentNotification(
		email,
		"Просрочка платежа по кредиту",
		amount,
	)
}

// CheckPayments запускает проверку платежей вручную
func (s *Scheduler) CheckPayments() error {
	// Получаем все активные кредиты
	credits, err := s.creditRepo.GetActiveCredits(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get active credits: %v", err)
	}

	for _, credit := range credits {
		// Получаем счет
		account, err := s.accountRepo.GetByID(context.Background(), credit.AccountID)
		if err != nil {
			return fmt.Errorf("failed to get account: %v", err)
		}

		// Получаем пользователя
		user, err := s.userRepo.GetByID(context.Background(), account.UserID)
		if err != nil {
			return fmt.Errorf("failed to get user: %v", err)
		}

		if user == nil {
			return fmt.Errorf("user not found")
		}

		// Проверяем, нужно ли списать платеж
		if time.Now().After(credit.NextPayment) {
			// Получаем график платежей
			schedule, err := s.creditRepo.GetPaymentSchedule(context.Background(), credit.ID)
			if err != nil {
				return fmt.Errorf("failed to get payment schedule: %v", err)
			}

			// Находим следующий платеж
			var nextPayment *domain.PaymentSchedule
			for _, payment := range schedule {
				if payment.Status == "PENDING" {
					nextPayment = &payment
					break
				}
			}

			if nextPayment != nil {
				// Проверяем достаточно ли средств
				if account.Balance >= nextPayment.Amount {
					// Списываем платеж
					if err := s.ProcessPayment(credit.ID, nextPayment.PaymentNumber); err != nil {
						return fmt.Errorf("failed to process payment: %v", err)
					}
				} else {
					// Отмечаем платеж как просроченный
					nextPayment.Status = "OVERDUE"
					if err := s.creditRepo.UpdatePaymentSchedule(context.Background(), nextPayment); err != nil {
						return fmt.Errorf("failed to update payment schedule: %v", err)
					}

					// Отправляем уведомление пользователю
					if err := s.sendPaymentOverdueNotification(user.Email, credit.ID, nextPayment.Amount); err != nil {
						return fmt.Errorf("failed to send notification: %v", err)
					}
				}
			}
		}
	}

	return nil
}

// GetAllCredits возвращает список всех кредитов
func (s *Scheduler) GetAllCredits() ([]domain.Credit, error) {
	return s.creditRepo.GetActiveCredits(context.Background())
}
