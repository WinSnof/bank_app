package security

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/armor"
	"golang.org/x/crypto/bcrypt"
)

// EncryptData шифрует данные с использованием PGP.
func EncryptData(data string) (string, error) {
	// Загрузка публичного ключа из файла
	publicKeyFile := "public_key.asc"

	// Проверка существования файла с публичным ключом
	if _, err := os.Stat(publicKeyFile); os.IsNotExist(err) {
		fmt.Println("Отсутствует ключ, генерируем его")
		MainGenerateKeyPair() // Генерация ключа
	}

	// Чтение публичного ключа из файла
	publicKey, err := ioutil.ReadFile(publicKeyFile)
	if err != nil {
		return "", fmt.Errorf("error reading public key file: %v", err)
	}

	entityList, err := openpgp.ReadArmoredKeyRing(strings.NewReader(string(publicKey)))
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	writer, err := armor.Encode(&buf, "PGP MESSAGE", nil)
	if err != nil {
		return "", err
	}

	plaintext, err := openpgp.Encrypt(writer, entityList, nil, nil, nil)
	if err != nil {
		return "", err
	}
	defer plaintext.Close()

	if _, err := plaintext.Write([]byte(data)); err != nil {
		return "", err
	}

	if err := writer.Close(); err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

// GenerateHMAC генерирует HMAC для данных.
func GenerateHMAC(data string, secret []byte) string {
	h := hmac.New(sha256.New, secret)
	h.Write([]byte(data))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func GenerateCVV() string {
	rand.Seed(time.Now().UnixNano())
	cvv := rand.Intn(900) + 100 // Генерируем случайное число от 100 до 999
	return fmt.Sprintf("%d", cvv)
}

func GenerateExpiryDate() string {
	currentTime := time.Now()
	// Получаем текущий месяц и год
	month := int(currentTime.Month())
	year := currentTime.Year() + 5 // Добавляем 5 лет к текущему году

	// Форматируем как MM/YY, где MM - 01-12, YY - последние две цифры года
	return fmt.Sprintf("%02d/%02d", month, year%100) // Форматируем как MM/YY
}

// HashCVV хеширует CVV с использованием bcrypt.
func HashCVV(cvv string) (string, error) {
	hashedCVV, err := bcrypt.GenerateFromPassword([]byte(cvv), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedCVV), nil
}

// Генерация валидного номера карты
func GenerateCardNumber(prefix string, length int) string {
	var cardNumber strings.Builder

	// Если префикс не указан, выбираем случайный
	if prefix == "" {
		prefixes := []string{"4", "5", "34", "37", "6"}
		prefix = prefixes[rand.Intn(len(prefixes))]
	}

	// Записываем префикс
	cardNumber.WriteString(prefix)

	// Генерируем случайные цифры до нужной длины (минус контрольная цифра)
	for cardNumber.Len() < length-1 {
		cardNumber.WriteString(strconv.Itoa(rand.Intn(10)))
	}

	// Вычисляем контрольную цифру
	checkDigit := calculateLuhnCheckDigit(cardNumber.String())
	cardNumber.WriteString(strconv.Itoa(checkDigit))

	return cardNumber.String()
}

// Проверка валидности номера карты по алгоритму Луна
func IsValidCardNumber(number string) bool {
	sum := 0
	alternate := false

	// Идем по цифрам справа налево
	for i := len(number) - 1; i >= 0; i-- {
		digit, err := strconv.Atoi(string(number[i]))
		if err != nil {
			return false // Нечисловой символ
		}

		if alternate {
			digit *= 2
			if digit > 9 {
				digit = (digit % 10) + 1
			}
		}

		sum += digit
		alternate = !alternate
	}

	// Номер валиден, если сумма кратна 10
	return sum%10 == 0
}

// Вычисление контрольной цифры по алгоритму Луна
func calculateLuhnCheckDigit(number string) int {
	sum := 0
	alternate := true // Начинаем с удвоения первой цифры справа

	// Идем по цифрам справа налево
	for i := len(number) - 1; i >= 0; i-- {
		digit, _ := strconv.Atoi(string(number[i]))

		if alternate {
			digit *= 2
			if digit > 9 {
				digit = (digit % 10) + 1
			}
		}

		sum += digit
		alternate = !alternate
	}

	// Вычисляем контрольную цифру
	return (10 - (sum % 10)) % 10
}
