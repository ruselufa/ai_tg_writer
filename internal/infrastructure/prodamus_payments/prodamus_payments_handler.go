package prodamus_payments

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
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

	// Генерируем уникальный order_id
	orderID := fmt.Sprintf("sub_%d_%s", userID, time.Now().Format("20060102150405"))
	
	data.Set("order_id", orderID)
	data.Set("do", "link")
	data.Set("urlSuccess", "https://aiwhisper.ru/payment/success")
	data.Set("urlReturn", "https://aiwhisper.ru/payment/fail")
	data.Set("urlNotification", "https://aiwhisper.ru/payment/webhook")
	
	// Добавляем продукт для подписки
	data.Set("products[0][name]", "Премиум подписка")
	data.Set("products[0][price]", fmt.Sprintf("%.2f", amount))
	data.Set("products[0][quantity]", "1")

	// Создаем подпись
	signature := h.CreateSignature(data)
	data.Set("signature", signature)

	// Формируем URL
	paymentURL := fmt.Sprintf("%s?%s", h.APIURL, data.Encode())

	return paymentURL, nil
}

// CreatePaymentLink создает ссылку для разового платежа
func (h *ProdamusHandler) CreatePaymentLink(userID int64, amount float64, description string) (string, error) {
	data := url.Values{}

	// Генерируем уникальный order_id
	orderID := fmt.Sprintf("pay_%d_%s", userID, time.Now().Format("20060102150405"))
	
	data.Set("order_id", orderID)
	data.Set("amount", fmt.Sprintf("%.2f", amount))
	data.Set("description", description)
	data.Set("do", "link")
	data.Set("urlSuccess", "https://aiwhisper.ru/payment/success")
	data.Set("urlReturn", "https://aiwhisper.ru/payment/fail")
	data.Set("urlNotification", "https://aiwhisper.ru/payment/webhook")

	// Создаем подпись
	signature := h.CreateSignature(data)
	data.Set("signature", signature)

	// Формируем URL
	paymentURL := fmt.Sprintf("%s?%s", h.APIURL, data.Encode())

	return paymentURL, nil
}

// VerifyWebhook проверяет подпись вебхука
func (h *ProdamusHandler) VerifyWebhook(data url.Values, signature string) bool {
	expectedSignature := h.CreateSignature(data)
	return expectedSignature == signature
}

// VerifySubscriptionWebhook проверяет подпись вебхука подписки
func (h *ProdamusHandler) VerifySubscriptionWebhook(data url.Values, signature string) bool {
	expectedSignature := h.CreateSubscriptionSignature(data)
	return strings.ToLower(expectedSignature) == strings.ToLower(signature)
}

// ProcessWebhook обрабатывает вебхуки от Prodamus
func (h *ProdamusHandler) ProcessWebhook(data url.Values, signature string) (*WebhookData, error) {
	if !h.VerifyWebhook(data, signature) {
		return nil, fmt.Errorf("invalid signature")
	}

	var webhookData WebhookData
	
	// Парсим данные из url.Values
	webhookData.Date = data.Get("date")
	webhookData.OrderID = data.Get("order_id")
	webhookData.Sum = data.Get("sum")
	webhookData.CustomerPhone = data.Get("customer_phone")
	webhookData.CustomerEmail = data.Get("customer_email")
	webhookData.PaymentType = data.Get("payment_type")

	// Парсим данные подписки, если они есть
	if subscriptionData := data.Get("subscription"); subscriptionData != "" {
		var subscription SubscriptionWebhook
		if err := json.Unmarshal([]byte(subscriptionData), &subscription); err == nil {
			webhookData.Subscription = &subscription
		}
	}

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

	signature := h.CreateSignature(data)
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

// CreateSignature создает подпись для запроса
func (h *ProdamusHandler) CreateSignature(data url.Values) string {
	// Сортируем параметры по алфавиту
	keys := make([]string, 0, len(data))
	for k := range data {
		if k != "signature" { // Исключаем параметр signature
			keys = append(keys, k)
		}
	}
	
	// Сортируем ключи по алфавиту
	sort.Strings(keys)

	// Формируем строку для подписи
	var parts []string
	for _, key := range keys {
		value := data.Get(key)
		if value != "" { // Добавляем только непустые значения
			parts = append(parts, fmt.Sprintf("%s=%s", key, value))
		}
	}

	signString := strings.Join(parts, "&")

	// Создаем HMAC-SHA256 подпись
	mac := hmac.New(sha256.New, []byte(h.WebhookSecret))
	mac.Write([]byte(signString))

	return hex.EncodeToString(mac.Sum(nil))
}

// CreateSubscriptionSignature создает подпись для подписок (использует JSON как в PHP)
func (h *ProdamusHandler) CreateSubscriptionSignature(data url.Values) string {
	// Преобразуем url.Values в map[string]interface{}
	dataMap := make(map[string]interface{})
	for key, values := range data {
		if key != "signature" {
			if len(values) == 1 {
				dataMap[key] = values[0]
			} else {
				dataMap[key] = values
			}
		}
	}
	
	// Преобразуем все значения в строки (как array_walk_recursive в PHP)
	h.convertToStrings(dataMap)
	
	// Сортируем рекурсивно
	h.sortRecursive(dataMap)
	
	// Преобразуем в JSON (как в PHP)
	jsonData, err := json.Marshal(dataMap)
	if err != nil {
		return ""
	}
	
	// Создаем HMAC-SHA256 подпись
	mac := hmac.New(sha256.New, []byte(h.WebhookSecret))
	mac.Write(jsonData)
	
	return hex.EncodeToString(mac.Sum(nil))
}

// convertToStrings преобразует все значения в строки (как array_walk_recursive в PHP)
func (h *ProdamusHandler) convertToStrings(data map[string]interface{}) {
	for key, value := range data {
		switch v := value.(type) {
		case map[string]interface{}:
			h.convertToStrings(v)
		case []interface{}:
			for i, item := range v {
				if str, ok := item.(string); ok {
					v[i] = str
				} else {
					v[i] = fmt.Sprintf("%v", item)
				}
			}
			data[key] = v
		default:
			data[key] = fmt.Sprintf("%v", v)
		}
	}
}

// sortRecursive сортирует рекурсивно (как _sort в PHP)
func (h *ProdamusHandler) sortRecursive(data map[string]interface{}) {
	// Сортируем ключи
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	
	// Создаем новый отсортированный map
	sortedData := make(map[string]interface{})
	for _, key := range keys {
		value := data[key]
		
		// Рекурсивно сортируем вложенные массивы
		if nestedMap, ok := value.(map[string]interface{}); ok {
			h.sortRecursive(nestedMap)
			sortedData[key] = nestedMap
		} else {
			sortedData[key] = value
		}
	}
	
	// Копируем отсортированные данные обратно
	for k, v := range sortedData {
		data[k] = v
	}
}
