package yookassa

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type Client struct {
	ShopID    string
	SecretKey string
	BaseURL   string
	HTTP      *http.Client
}

func New() *Client {
	return &Client{
		ShopID:    os.Getenv("YK_SHOP_ID"),
		SecretKey: os.Getenv("YK_SECRET_KEY"),
		BaseURL:   "https://api.yookassa.ru/v3",
		HTTP: &http.Client{
			Timeout: 15 * time.Second,
			// Добавляем пул соединений для лучшей производительности при параллельных запросах
			Transport: &http.Transport{
				MaxIdleConns:        100,              // Максимум неактивных соединений
				MaxIdleConnsPerHost: 10,               // Максимум неактивных соединений на хост
				IdleConnTimeout:     90 * time.Second, // Время жизни неактивного соединения
				DisableCompression:  false,            // Включаем сжатие для экономии трафика
				ForceAttemptHTTP2:   true,             // Принудительно используем HTTP/2
			},
		},
	}
}

func (c *Client) authHeader() string {
	creds := c.ShopID + ":" + c.SecretKey
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(creds))
}

func (c *Client) do(idemKey, method, path string, body any, out any) error {
	b, _ := json.Marshal(body)

	log.Printf("🌐 YooKassa HTTP Request:")
	log.Printf("   URL: %s%s", c.BaseURL, path)
	log.Printf("   Method: %s", method)
	log.Printf("   Body: %s", string(b))
	log.Printf("   ShopID: %s (len=%d)", c.ShopID, len(c.ShopID))
	log.Printf("   SecretKey: %s... (len=%d)", c.SecretKey[:min(8, len(c.SecretKey))], len(c.SecretKey))

	req, _ := http.NewRequest(method, c.BaseURL+path, bytes.NewReader(b))
	req.Header.Set("Authorization", c.authHeader())
	req.Header.Set("Content-Type", "application/json")
	if idemKey != "" {
		req.Header.Set("Idempotence-Key", idemKey) // см. Idempotence-Key
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		log.Printf("❌ HTTP request error: %v", err)
		return err
	}
	defer resp.Body.Close()

	// Читаем ответ для логирования
	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("📡 YooKassa HTTP Response:")
	log.Printf("   Status: %d", resp.StatusCode)
	log.Printf("   Body: %s", string(respBody))

	if resp.StatusCode >= 300 {
		log.Printf("❌ YooKassa HTTP error %d: %s", resp.StatusCode, string(respBody))
		return fmt.Errorf("yookassa http %d: %s", resp.StatusCode, string(respBody))
	}
	if out != nil {
		return json.Unmarshal(respBody, out)
	}
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type Amount struct{ Value, Currency string }

// 5.1 Первичный платеж с сохранением метода + customer.id
func (c *Client) CreateInitialPayment(idemKey string, amount Amount, description, customerID, returnURL string, metadata map[string]string) (map[string]any, error) {
	payload := map[string]any{
		"amount":  map[string]string{"value": amount.Value, "currency": amount.Currency},
		"capture": true,
		"confirmation": map[string]string{
			"type": "redirect", "return_url": returnURL,
		},
		"description":         description,
		"save_payment_method": true,
		"customer": map[string]string{
			"id": customerID, // обязателен для привязки
		},
		"metadata": metadata,
		"receipt": map[string]any{
			"customer": map[string]string{
				"email": "noreply@aiwhisper.ru",
			},
			"items": []map[string]any{
				{
					"description": description,
					"amount": map[string]string{
						"value":    amount.Value,
						"currency": amount.Currency,
					},
					"vat_code":        "1", // НДС 20%
					"quantity":        "1",
					"payment_subject": "service",      // Предмет расчета - услуга
					"payment_mode":    "full_payment", // Способ расчета - полный расчет
				},
			},
		},
	}
	var out map[string]any
	err := c.do(idemKey, "POST", "/payments", payload, &out)
	return out, err
}

// 5.2 Рекуррентный платеж по сохраненному payment_method_id + customer_id
func (c *Client) CreateRecurringPayment(idemKey string, amount Amount, description, customerID, paymentMethodID string, metadata map[string]string) (map[string]any, error) {
	payload := map[string]any{
		"amount":            map[string]string{"value": amount.Value, "currency": amount.Currency},
		"capture":           true,
		"payment_method_id": paymentMethodID,
		"customer": map[string]string{
			"id": customerID,
		},
		"description": description,
		"metadata":    metadata,
	}
	var out map[string]any
	err := c.do(idemKey, "POST", "/payments", payload, &out)
	return out, err
}

func (c *Client) GetPayment(id string) (map[string]any, error) {
	var out map[string]any
	err := c.do("", "GET", "/payments/"+id, nil, &out)
	return out, err
}

// CreateCustomer создает customer в YooKassa
func (c *Client) CreateCustomer(idemKey string, email, phone string) (map[string]any, error) {
	payload := map[string]any{
		"email": email,
		"phone": phone,
	}
	var out map[string]any
	err := c.do(idemKey, "POST", "/customers", payload, &out)
	return out, err
}
