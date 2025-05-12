package services

import (
	"bytes"
	"crypto/tls"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	// "strings"
	"time"

	"gopkg.in/gomail.v2"
)

// Структуры для работы с XML-ответом от ЦБ РФ
type KeyRateEnvelope struct {
	XMLName xml.Name    `xml:"Envelope"`
	Body    KeyRateBody `xml:"Body"`
}

type KeyRateBody struct {
	Response KeyRateResponse `xml:"KeyRateXMLResponse"`
}

type KeyRateResponse struct {
	Result KeyRateResult `xml:"KeyRateXMLResult"`
}

type KeyRateResult struct {
	Rows []KeyRateRows `xml:"KeyRate"`
}

type KeyRateRows struct {
	KeyRates []KeyRates `xml:"KR"`
}

type KeyRates struct {
	Date string `xml:"DT" json:"date"`
	Rate string `xml:"Rate" json:"rate"`
}

// Структура для SOAP-запроса
type GetKeyRateXMLRequest struct {
	XMLName  xml.Name `xml:"KeyRateXML"`
	Xmlns    string   `xml:"xmlns,attr"`
	FromDate string   `xml:"fromDate"`
	ToDate   string   `xml:"ToDate"`
}

type ExternalService struct {
	smtpHost     string
	smtpPort     int
	smtpUsername string
	smtpPassword string
	emailFrom    string
}

func NewExternalService(smtpHost string, smtpPort int, smtpUsername, smtpPassword, emailFrom string) *ExternalService {
	return &ExternalService{
		smtpHost:     smtpHost,
		smtpPort:     smtpPort,
		smtpUsername: smtpUsername,
		smtpPassword: smtpPassword,
		emailFrom:    emailFrom,
	}
}

// GetKeyRate получает ключевую ставку от ЦБ РФ
func (s *ExternalService) GetKeyRate() (float64, error) {
	// Формируем даты для запроса (последние 30 дней)
	now := time.Now()
	fromDate := now.AddDate(0, 0, -30).Format("2006-01-02")
	toDate := now.Format("2006-01-02")

	request := GetKeyRateXMLRequest{
		Xmlns:    "http://web.cbr.ru/",
		FromDate: fromDate,
		ToDate:   toDate,
	}

	// Формируем SOAP-запрос
	var root = struct {
		XMLName xml.Name `xml:"soap12:Envelope"`
		Xsi     string   `xml:"xmlns:xsi,attr"`
		Xsd     string   `xml:"xmlns:xsd,attr"`
		Soap12  string   `xml:"xmlns:soap12,attr"`
		Body    struct {
			XMLName xml.Name             `xml:"soap12:Body"`
			Request GetKeyRateXMLRequest `xml:"KeyRateXML"`
		}
	}{
		Xsi:    "http://www.w3.org/2001/XMLSchema-instance",
		Xsd:    "http://www.w3.org/2001/XMLSchema",
		Soap12: "http://www.w3.org/2003/05/soap-envelope",
	}
	root.Body.Request = request

	out, _ := xml.MarshalIndent(&root, " ", "  ")

	// Создаем HTTP-клиент с поддержкой TLS
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	// Отправляем запрос
	resp, err := client.Post(
		"https://www.cbr.ru/DailyInfoWebServ/DailyInfo.asmx",
		"application/soap+xml; charset=utf-8",
		bytes.NewBuffer(out),
	)
	if err != nil {
		return 0, fmt.Errorf("ошибка при запросе к ЦБ РФ: %v", err)
	}
	defer resp.Body.Close()

	// Читаем ответ
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("ошибка при чтении ответа: %v", err)
	}

	// Парсим XML
	var data KeyRateEnvelope
	unmarshalErr := xml.Unmarshal(body, &data)
	if unmarshalErr != nil {
		return 0, fmt.Errorf("ошибка при парсинге XML: %v", err)
	}

	// Получаем последнюю ставку
	if len(data.Body.Response.Result.Rows) == 0 || len(data.Body.Response.Result.Rows[0].KeyRates) == 0 {
		return 0, fmt.Errorf("не найдены данные о ключевой ставке")
	}

	lastRate := data.Body.Response.Result.Rows[0].KeyRates[len(data.Body.Response.Result.Rows[0].KeyRates)-1]
	rate, err := strconv.ParseFloat(lastRate.Rate, 64)
	if err != nil {
		return 0, fmt.Errorf("ошибка при конвертации ставки: %v", err)
	}

	return rate, nil
}

// SendEmail отправляет email уведомление
func (s *ExternalService) SendEmail(to, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.emailFrom)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer(s.smtpHost, s.smtpPort, s.smtpUsername, s.smtpPassword)

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("ошибка при отправке email: %v", err)
	}

	return nil
}

// SendPaymentNotification отправляет уведомление о платеже
func (s *ExternalService) SendPaymentNotification(email, paymentType string, amount float64) error {
	subject := fmt.Sprintf("Уведомление о платеже - %s", paymentType)
	body := fmt.Sprintf(`
		<h1>Уведомление о платеже</h1>
		<p>Тип платежа: %s</p>
		<p>Сумма: %.2f ₽</p>
		<p>Дата: %s</p>
	`, paymentType, amount, time.Now().Format("02.01.2006 15:04:05"))

	return s.SendEmail(email, subject, body)
}
