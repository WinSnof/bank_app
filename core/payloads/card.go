package payloads

// Структура для выдачи информации о карте
type UnsecureCard struct {
	ID          uint   `json:"id"`
	AccountID   uint   `json:"account_id"`
	AccountName string `json:"account_name"`
	Number      string `json:"number"`
	ExpiryDate  string `json:"expiry_date"` // Срок действия карты (не зашифрованый).
	CVV         string `json:"CVV"`         // CVV код (не хешированый).
}
