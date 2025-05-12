package services

import (
	"FinanceGolang/core/dbaccess"
	"FinanceGolang/core/domain"
	"FinanceGolang/core/payloads"
	"FinanceGolang/core/security"
	"context"
	"fmt"
	"regexp"
	"time"
)

type CardService interface {
	CreateCard(card *domain.Card, userID uint) (*payloads.UnsecureCard, error)
	GetCardByID(id uint) (*domain.Card, error)
	GetUserCards(userID uint) ([]domain.Card, error)
}

type cardService struct {
	cardRepo    dbaccess.CardRepository
	accountRepo dbaccess.AccountRepository
	publicKey   string
	hmacSecret  []byte
}

func CardServiceInstance(cardRepo dbaccess.CardRepository, accountRepo dbaccess.AccountRepository, publicKey string, hmacSecret []byte) CardService {
	return &cardService{
		cardRepo:    cardRepo,
		accountRepo: accountRepo,
		publicKey:   publicKey,
		hmacSecret:  hmacSecret,
	}
}

func (s *cardService) CreateCard(card *domain.Card, userID uint) (*payloads.UnsecureCard, error) {
	// Проверяем, что счет принадлежит пользователю
	accounts, err := s.accountRepo.GetByUserID(context.Background(), userID)
	if err != nil {
		return nil, fmt.Errorf("could not get user accounts: %v", err)
	}

	accountExists := false
	var accountName string
	for _, account := range accounts {
		if account.ID == card.AccountID {
			accountExists = true
			accountName = account.Number
			break
		}
	}

	if !accountExists {
		return nil, fmt.Errorf("account does not belong to the user")
	}

	var unsecureCard payloads.UnsecureCard

	// Генерируем данные карты
	unsecureCard.Number = security.GenerateCardNumber("4", 16)
	unsecureCard.CVV = security.GenerateCVV()
	unsecureCard.ExpiryDate = security.GenerateExpiryDate()
	unsecureCard.AccountName = accountName
	unsecureCard.AccountID = card.AccountID

	// Проверка валидности номера карты
	if !security.IsValidCardNumber(unsecureCard.Number) {
		return nil, fmt.Errorf("invalid card number: %s", unsecureCard.Number)
	}

	// Дополнительная валидация формата номера карты
	cardRegex := regexp.MustCompile(`^[0-9]{16}$`)
	if !cardRegex.MatchString(unsecureCard.Number) {
		return nil, fmt.Errorf("invalid card number format: %s", unsecureCard.Number)
	}

	// Валидация CVV
	cvvRegex := regexp.MustCompile(`^[0-9]{3}$`)
	if !cvvRegex.MatchString(unsecureCard.CVV) {
		return nil, fmt.Errorf("invalid CVV format: %s", unsecureCard.CVV)
	}

	// Валидация даты истечения срока действия
	expiryRegex := regexp.MustCompile(`^(0[1-9]|1[0-2])\/([0-9]{2})$`)
	if !expiryRegex.MatchString(unsecureCard.ExpiryDate) {
		return nil, fmt.Errorf("invalid expiry date format: %s", unsecureCard.ExpiryDate)
	}

	// Создаем временную карту для валидации
	tempCard := &domain.Card{
		Number:     unsecureCard.Number,
		ExpiryDate: unsecureCard.ExpiryDate,
		CVV:        unsecureCard.CVV,
		AccountID:  unsecureCard.AccountID,
		UserID:     userID,
		IsActive:   true,
	}

	// Валидация карты
	if err := tempCard.Validate(); err != nil {
		return nil, fmt.Errorf("invalid card data: %v (number: %s, expiry: %s, cvv: %s, account_id: %d)",
			err, unsecureCard.Number, unsecureCard.ExpiryDate, unsecureCard.CVV, unsecureCard.AccountID)
	}

	// Шифрование номера карты и срока действия
	encryptedNumber, err := security.EncryptData(unsecureCard.Number)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt card number: %v", err)
	}
	encryptedExpiryDate, err := security.EncryptData(unsecureCard.ExpiryDate)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt expiry date: %v", err)
	}

	// Хеширование CVV
	hashedCVV, err := security.HashCVV(unsecureCard.CVV)
	if err != nil {
		return nil, fmt.Errorf("failed to hash CVV: %v", err)
	}

	// Сохранение зашифрованных данных в структуру
	card.Number = encryptedNumber
	card.ExpiryDate = encryptedExpiryDate
	card.CVV = hashedCVV
	card.CreatedAt = time.Now()
	card.UserID = userID
	card.AccountID = unsecureCard.AccountID
	card.IsActive = true

	// Сохранение карты в базе данных
	if err := s.cardRepo.Create(context.Background(), card); err != nil {
		return nil, fmt.Errorf("failed to save card: %v", err)
	}

	// Устанавливаем ID в unsecureCard для возврата
	unsecureCard.ID = card.ID

	return &unsecureCard, nil
}

func (s *cardService) GetCardByID(id uint) (*domain.Card, error) {
	card, err := s.cardRepo.GetByID(context.Background(), id)
	if err != nil {
		return nil, err
	}
	return card, nil
}

func (s *cardService) GetUserCards(userID uint) ([]domain.Card, error) {
	// Получаем все счета пользователя
	accounts, err := s.accountRepo.GetByUserID(context.Background(), userID)
	if err != nil {
		return nil, fmt.Errorf("could not get user accounts: %v", err)
	}

	// Собираем ID счетов пользователя
	accountIDs := make([]uint, len(accounts))
	for i, account := range accounts {
		accountIDs[i] = account.ID
	}

	// Получаем карты только для счетов пользователя
	var allCards []domain.Card
	for _, accountID := range accountIDs {
		cards, err := s.cardRepo.GetByAccountID(context.Background(), accountID)
		if err != nil {
			return nil, err
		}
		allCards = append(allCards, cards...)
	}

	return allCards, nil
}
