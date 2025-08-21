package deepseek

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type DeepSeekHandler struct {
	apiKey string
	apiURL string
	client *http.Client
}

type DeepSeekMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type DeepSeekRequest struct {
	Model       string            `json:"model"`
	Messages    []DeepSeekMessage `json:"messages"`
	Temperature float64           `json:"temperature"`
	MaxTokens   int               `json:"max_tokens"`
}

type DeepSeekChoice struct {
	Message DeepSeekMessage `json:"message"`
}

type DeepSeekResponse struct {
	Choices []DeepSeekChoice `json:"choices"`
}

func NewDeepSeekHandler() *DeepSeekHandler {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		log.Println("Предупреждение: DEEPSEEK_API_KEY не установлен")
		return &DeepSeekHandler{
			apiKey: "",
			apiURL: "https://api.deepseek.com/v1/chat/completions",
			client: &http.Client{
				Timeout: 60 * time.Second,
				// ОПТИМИЗАЦИЯ: Пул соединений для лучшей производительности
				Transport: &http.Transport{
					MaxIdleConns:        50,               // Максимум неактивных соединений
					MaxIdleConnsPerHost: 10,               // Максимум на хост
					IdleConnTimeout:     90 * time.Second, // Время жизни соединения
					DisableCompression:  false,            // Включаем сжатие для экономии трафика
					ForceAttemptHTTP2:   true,             // Принудительно используем HTTP/2
					// Дополнительные настройки для стабильности
					MaxConnsPerHost:       100,              // Максимум соединений на хост
					ResponseHeaderTimeout: 30 * time.Second, // Таймаут заголовков ответа
					// Настройки для TLS
					TLSHandshakeTimeout: 10 * time.Second, // Таймаут TLS handshake
				},
			},
		}
	}

	return &DeepSeekHandler{
		apiKey: apiKey,
		apiURL: "https://api.deepseek.com/v1/chat/completions",
		client: &http.Client{
			Timeout: 60 * time.Second,
			// ОПТИМИЗАЦИЯ: Пул соединений для лучшей производительности
			Transport: &http.Transport{
				MaxIdleConns:        50,               // Максимум неактивных соединений
				MaxIdleConnsPerHost: 10,               // Максимум на хост
				IdleConnTimeout:     90 * time.Second, // Время жизни соединения
				DisableCompression:  false,            // Включаем сжатие для экономии трафика
				ForceAttemptHTTP2:   true,             // Принудительно используем HTTP/2
				// Дополнительные настройки для стабильности
				MaxConnsPerHost:       100,              // Максимум соединений на хост
				ResponseHeaderTimeout: 30 * time.Second, // Таймаут заголовков ответа
				// Настройки для TLS
				TLSHandshakeTimeout: 10 * time.Second, // Таймаут TLS handshake
			},
		},
	}
}

// RewriteText переписывает текст с помощью DeepSeek
func (dh *DeepSeekHandler) RewriteText(originalText string) (string, error) {
	if dh.apiKey == "" {
		return "🔧 Функция переписывания текста временно недоступна", nil
	}

	prompt := fmt.Sprintf(`Сделай из этого текста красивый пост для Telegram-канала с использованием MarkdownV2:
- Используй *жирный* для заголовков и важных мыслей.
- Используй _курсив_ для акцентов.
- Для списков используй эмодзи (например, 🔹, ✔️, ▫️).
- Между абзацами и пунктами делай пустую строку.
- Экранируй все спецсимволы MarkdownV2: _ * [ ] ( ) ~ > # + - = | { } . !
- В конце добавь 3-5 релевантных хештегов.

Исходный текст:
\"%s\"

Сделай красивый Telegram-пост с правильной разметкой:`, originalText)

	request := DeepSeekRequest{
		Model: "deepseek-chat",
		Messages: []DeepSeekMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.7,
		MaxTokens:   1000,
	}

	response, err := dh.makeRequest(request)
	if err != nil {
		return "", fmt.Errorf("ошибка DeepSeek API: %v", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("пустой ответ от DeepSeek")
	}

	rewrittenText := strings.TrimSpace(response.Choices[0].Message.Content)
	return rewrittenText, nil
}

// ImproveText улучшает качество текста
func (dh *DeepSeekHandler) ImproveText(text string, style string) (string, error) {
	if dh.apiKey == "" {
		return "🔧 Функция улучшения текста временно недоступна", nil
	}

	var stylePrompt string
	switch style {
	case "formal":
		stylePrompt = "Перепиши текст в формальном, деловом стиле"
	case "casual":
		stylePrompt = "Перепиши текст в разговорном, дружелюбном стиле"
	case "academic":
		stylePrompt = "Перепиши текст в академическом стиле"
	default:
		stylePrompt = "Улучши текст, сделав его более читаемым и структурированным"
	}

	prompt := fmt.Sprintf(`%s. Сохрани основную мысль, но улучши стиль, грамматику и пунктуацию:

"%s"

Улучшенный текст:`, stylePrompt, text)

	request := DeepSeekRequest{
		Model: "deepseek-chat",
		Messages: []DeepSeekMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.7,
		MaxTokens:   1000,
	}

	response, err := dh.makeRequest(request)
	if err != nil {
		return "", fmt.Errorf("ошибка DeepSeek API: %v", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("пустой ответ от DeepSeek")
	}

	improvedText := strings.TrimSpace(response.Choices[0].Message.Content)
	return improvedText, nil
}

// SummarizeText создает краткое изложение текста
func (dh *DeepSeekHandler) SummarizeText(text string) (string, error) {
	if dh.apiKey == "" {
		return "🔧 Функция создания краткого изложения временно недоступна", nil
	}

	prompt := fmt.Sprintf(`Создай краткое изложение следующего текста, выделив основные мысли:

"%s"

Краткое изложение:`, text)

	request := DeepSeekRequest{
		Model: "deepseek-chat",
		Messages: []DeepSeekMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.5,
		MaxTokens:   500,
	}

	response, err := dh.makeRequest(request)
	if err != nil {
		return "", fmt.Errorf("ошибка DeepSeek API: %v", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("пустой ответ от DeepSeek")
	}

	summary := strings.TrimSpace(response.Choices[0].Message.Content)
	return summary, nil
}

// CreateTelegramPost создает красивый пост для Telegram с хештегами
func (dh *DeepSeekHandler) CreateTelegramPost(originalText string) (string, error) {
	if dh.apiKey == "" {
		return "🔧 Функция создания постов временно недоступна", nil
	}

	prompt := fmt.Sprintf(`Создай привлекательный пост для Telegram канала на основе этого текста. 

Требования к форматированию:
- Используй *жирный* для заголовков и важных мыслей.
- Используй _курсив_ для акцентов.
- Для списков используй эмодзи (например, 🔹, ✔️, ▫️).
- Между абзацами и пунктами делай пустую строку.
- Экранируй все спецсимволы MarkdownV2: _ * [ ] ( ) ~ > # + - = | { } . !
- В конце добавь 3-5 релевантных хештегов.

Исходный текст:
"%s"

Создай красивый Telegram пост с правильной разметкой:`, originalText)

	request := DeepSeekRequest{
		Model: "deepseek-chat",
		Messages: []DeepSeekMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.8,
		MaxTokens:   1500,
	}

	response, err := dh.makeRequest(request)
	if err != nil {
		return "", fmt.Errorf("ошибка DeepSeek API: %v", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("пустой ответ от DeepSeek")
	}

	post := strings.TrimSpace(response.Choices[0].Message.Content)
	return post, nil
}

// makeRequest выполняет HTTP запрос к DeepSeek API
func (dh *DeepSeekHandler) makeRequest(request DeepSeekRequest) (*DeepSeekResponse, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("ошибка маршалинга запроса: %v", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), "POST", dh.apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+dh.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := dh.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка HTTP запроса: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ошибка API: %s - %s", resp.Status, string(body))
	}

	var response DeepSeekResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("ошибка парсинга ответа: %v", err)
	}

	return &response, nil
}
