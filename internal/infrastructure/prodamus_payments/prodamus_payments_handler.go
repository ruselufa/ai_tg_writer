package prodamus_payments

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Существующие структуры (оставляем без изменений)
type ProdamusPaymentRequest struct {
	OrderID     string  `json:"order_id"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
	Email       string  `json:"email"`
	Recurring   bool    `json:"recurrent"`
}

type ProdamusPaymentResponse struct {
	OrderID     string  `json:"order_id"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
	Email       string  `json:"email"`
	Recurring   bool    `json:"recurrent"`
}

// Новые структуры для подписок
type SubscriptionRequest struct {
	OrderID         string  `json:"order_id"`
	Amount          float64 `json:"amount"`
	Description     string  `json:"description"`
	Email           string  `json:"email"`
	Phone           string  `json:"phone"`
	SubscriptionID  int     `json:"subscription"`
	Do              string  `json:"do"`
	URLSuccess      string  `json:"urlSuccess"`
	URLReturn       string  `json:"urlReturn"`
	URLNotification string  `json:"urlNotification"`
}

type SubscriptionResponse struct {
	PaymentURL     string `json:"payment_url"`
	SubscriptionID int    `json:"subscription_id"`
	Status         string `json:"status"`
}

type WebhookData struct {
	Date          string               `json:"date"`
	OrderID       string               `json:"order_id"`
	Sum           string               `json:"sum"`
	CustomerPhone string               `json:"customer_phone"`
	CustomerEmail string               `json:"customer_email"`
	PaymentType   string               `json:"payment_type"`
	Subscription  *SubscriptionWebhook `json:"subscription"`
}

type SubscriptionWebhook struct {
	Type              string `json:"type"`
	ActionCode        string `json:"action_code"`
	ID                int    `json:"id"`
	Active            int    `json:"active"`
	ActiveManager     int    `json:"active_manager"`
	ActiveUser        int    `json:"active_user"`
	Cost              string `json:"cost"`
	Name              string `json:"name"`
	LimitAutopayments string `json:"limit_autopayments"`
	AutopaymentsNum   string `json:"autopayments_num"`
	DateCreate        string `json:"date_create"`
	DateFirstPayment  string `json:"date_first_payment"`
	DateLastPayment   string `json:"date_last_payment"`
	DateNextPayment   string `json:"date_next_payment"`
	PaymentNum        string `json:"payment_num"`
	Autopayment       string `json:"autopayment"`
}

type ProdamusHandler struct {
	APIKey        string
	APIURL        string
	WebhookSecret string
}

func NewProdamusHandler(apiKey, apiURL, webhookSecret string) *ProdamusHandler {
	return &ProdamusHandler{
		APIKey:        apiKey,
		APIURL:        apiURL,
		WebhookSecret: webhookSecret,
	}
}

// CreateSubscriptionLink создает ссылку для оформления подписки
func (h *ProdamusHandler) CreateSubscriptionLink(userID int64, tariff string, amount float64, subscriptionID int) (string, error) {
	data := url.Values{}

	data.Set("order_id", fmt.Sprintf("sub_%d_%s", userID, time.Now().Format("20060102150405")))
	data.Set("customer_phone", "+79999999999")     // Замените на реальный номер
	data.Set("customer_email", "user@example.com") // Замените на реальный email
	data.Set("subscription", strconv.Itoa(subscriptionID))
	data.Set("do", "link")
	data.Set("urlSuccess", "https://your-bot-domain.com/success")
	data.Set("urlReturn", "https://your-bot-domain.com/return")
	data.Set("urlNotification", "https://your-bot-domain.com/webhook")

	// Создаем подпись
	signature := h.createSignature(data)
	data.Set("signature", signature)

	// Формируем URL
	paymentURL := fmt.Sprintf("%s?%s", h.APIURL, data.Encode())

	return paymentURL, nil
}

// VerifyWebhook проверяет подпись вебхука
func (h *ProdamusHandler) VerifyWebhook(data url.Values, signature string) bool {
	expectedSignature := h.createSignature(data)
	return expectedSignature == signature
}

// ProcessWebhook обрабатывает вебхуки от Prodamus
func (h *ProdamusHandler) ProcessWebhook(data url.Values, signature string) (*WebhookData, error) {
	if !h.VerifyWebhook(data, signature) {
		return nil, fmt.Errorf("invalid signature")
	}

	var webhookData WebhookData
	// Преобразуем данные из url.Values в структуру
	// Здесь нужно реализовать парсинг данных

	return &webhookData, nil
}

// SetSubscriptionActivity управляет статусом подписки
func (h *ProdamusHandler) SetSubscriptionActivity(subscriptionID int, tgUserID int64, active bool) error {
	data := url.Values{}
	data.Set("subscription", strconv.Itoa(subscriptionID))
	data.Set("tg_user_id", strconv.FormatInt(tgUserID, 10))

	if active {
		data.Set("active_manager", "1")
	} else {
		data.Set("active_manager", "0")
	}

	signature := h.createSignature(data)
	data.Set("signature", signature)

	// Отправляем POST запрос
	resp, err := http.PostForm(h.APIURL+"/rest/setActivity/", data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("API returned status: %d", resp.StatusCode)
	}

	return nil
}

// createSignature создает подпись для запроса
func (h *ProdamusHandler) createSignature(data url.Values) string {
	// Сортируем параметры по алфавиту
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}

	// Формируем строку для подписи
	var parts []string
	for _, key := range keys {
		if key != "signature" { // Исключаем параметр signature
			parts = append(parts, fmt.Sprintf("%s=%s", key, data.Get(key)))
		}
	}

	signString := strings.Join(parts, "&")

	// Создаем HMAC-SHA256 подпись
	mac := hmac.New(sha256.New, []byte(h.WebhookSecret))
	mac.Write([]byte(signString))

	return hex.EncodeToString(mac.Sum(nil))
}
