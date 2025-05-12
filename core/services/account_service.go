package services

import (
	"FinanceGolang/core/dbaccess"
	"FinanceGolang/core/domain"
	"context"
	"errors"
	"fmt"
)

type AccountService interface {
	// Базовые операции со счетом
	CreateAccount(account *domain.Account, userID uint) error
	GetAccountByID(id uint) (*domain.Account, error)
	GetAccountByUserID(id uint) ([]domain.Account, error)
	GetAllAccounts() ([]domain.Account, error)

	// Операции с балансом
	Deposit(accountID uint, amount float64, description string) error
	Withdraw(accountID uint, amount float64, description string) error
	Transfer(fromAccountID, toAccountID uint, amount float64, description string) error

	// Операции с транзакциями
	GetTransactions(accountID uint) ([]domain.Transaction, error)
}

type accountService struct {
	accountRepo     dbaccess.AccountRepository
	transactionRepo dbaccess.TransactionRepository
}

func AccountServiceInstance(accountRepo dbaccess.AccountRepository, transactionRepo dbaccess.TransactionRepository) AccountService {
	return &accountService{
		accountRepo:     accountRepo,
		transactionRepo: transactionRepo,
	}
}

// Базовые операции со счетом
func (s *accountService) CreateAccount(account *domain.Account, userID uint) error {
	fmt.Println("Creating account for user ID:", userID)
	account.UserID = userID
	if err := s.accountRepo.Create(context.Background(), account); err != nil {
		return fmt.Errorf("could not create account: %v", err)
	}
	return nil
}

func (s *accountService) GetAccountByID(id uint) (*domain.Account, error) {
	account, err := s.accountRepo.GetByID(context.Background(), id)
	if err != nil {
		return nil, fmt.Errorf("could not get account by ID: %v", err)
	}
	if account == nil {
		return nil, fmt.Errorf("account not found with ID: %d", id)
	}
	return account, nil
}

func (s *accountService) GetAccountByUserID(userID uint) ([]domain.Account, error) {
	accounts, err := s.accountRepo.GetByUserID(context.Background(), userID)
	if err != nil {
		return nil, fmt.Errorf("could not get account by user ID: %v", err)
	}
	if len(accounts) == 0 {
		return nil, fmt.Errorf("no accounts found for user ID: %d", userID)
	}
	return accounts, nil
}

func (s *accountService) GetAllAccounts() ([]domain.Account, error) {
	accounts, err := s.accountRepo.List(context.Background(), 0, 1000)
	if err != nil {
		return nil, fmt.Errorf("could not get all accounts: %v", err)
	}
	if len(accounts) == 0 {
		return nil, fmt.Errorf("no accounts found")
	}
	return accounts, nil
}

// Операции с балансом
func (s *accountService) Deposit(accountID uint, amount float64, description string) error {
	if amount <= 0 {
		return errors.New("amount must be positive")
	}

	// Проверяем существование счета
	if _, err := s.accountRepo.GetByID(context.Background(), accountID); err != nil {
		return fmt.Errorf("failed to get account: %v", err)
	}

	// Создаем транзакцию
	transaction := &domain.Transaction{
		Type:        domain.TransactionTypeDeposit,
		ToAccountID: accountID,
		Amount:      amount,
		Description: description,
		Status:      domain.TransactionStatusCompleted,
	}

	// Обновляем баланс счета
	if err := s.accountRepo.UpdateBalance(context.Background(), accountID, amount); err != nil {
		return fmt.Errorf("failed to update account balance: %v", err)
	}

	// Сохраняем транзакцию
	if err := s.transactionRepo.Create(context.Background(), transaction); err != nil {
		return fmt.Errorf("failed to create transaction: %v", err)
	}

	return nil
}

func (s *accountService) Withdraw(accountID uint, amount float64, description string) error {
	if amount <= 0 {
		return errors.New("amount must be positive")
	}

	account, err := s.accountRepo.GetByID(context.Background(), accountID)
	if err != nil {
		return fmt.Errorf("failed to get account: %v", err)
	}

	if account.Balance < amount {
		return errors.New("insufficient funds")
	}

	// Создаем транзакцию
	transaction := &domain.Transaction{
		Type:          domain.TransactionTypeWithdrawal,
		FromAccountID: accountID,
		Amount:        amount,
		Description:   description,
		Status:        domain.TransactionStatusCompleted,
	}

	// Обновляем баланс счета
	if err := s.accountRepo.UpdateBalance(context.Background(), accountID, -amount); err != nil {
		return fmt.Errorf("failed to update account balance: %v", err)
	}

	// Сохраняем транзакцию
	if err := s.transactionRepo.Create(context.Background(), transaction); err != nil {
		return fmt.Errorf("failed to create transaction: %v", err)
	}

	return nil
}

func (s *accountService) Transfer(fromAccountID, toAccountID uint, amount float64, description string) error {
	if amount <= 0 {
		return errors.New("amount must be positive")
	}

	if fromAccountID == toAccountID {
		return errors.New("cannot transfer to the same account")
	}

	fromAccount, err := s.accountRepo.GetByID(context.Background(), fromAccountID)
	if err != nil {
		return fmt.Errorf("failed to get source account: %v", err)
	}

	// Проверяем существование целевого счета
	if _, err := s.accountRepo.GetByID(context.Background(), toAccountID); err != nil {
		return fmt.Errorf("failed to get destination account: %v", err)
	}

	if fromAccount.Balance < amount {
		return errors.New("insufficient funds")
	}

	// Создаем транзакцию
	transaction := &domain.Transaction{
		Type:          domain.TransactionTypeTransfer,
		FromAccountID: fromAccountID,
		ToAccountID:   toAccountID,
		Amount:        amount,
		Description:   description,
		Status:        domain.TransactionStatusCompleted,
	}

	// Обновляем балансы счетов
	if err := s.accountRepo.UpdateBalance(context.Background(), fromAccountID, -amount); err != nil {
		return fmt.Errorf("failed to update source account balance: %v", err)
	}

	if err := s.accountRepo.UpdateBalance(context.Background(), toAccountID, amount); err != nil {
		return fmt.Errorf("failed to update destination account balance: %v", err)
	}

	// Сохраняем транзакцию
	if err := s.transactionRepo.Create(context.Background(), transaction); err != nil {
		return fmt.Errorf("failed to create transaction: %v", err)
	}

	return nil
}

// Операции с транзакциями
func (s *accountService) GetTransactions(accountID uint) ([]domain.Transaction, error) {
	return s.transactionRepo.GetByAccountID(context.Background(), accountID)
}
