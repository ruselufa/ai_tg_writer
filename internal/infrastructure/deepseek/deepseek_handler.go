package deepseek

import (
	"ai_tg_writer/internal/monitoring"
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
				Timeout: 300 * time.Second, // Увеличиваем до 5 минут для генерации постов
				// ОПТИМИЗАЦИЯ: Пул соединений для лучшей производительности
				Transport: &http.Transport{
					MaxIdleConns:        50,               // Максимум неактивных соединений
					MaxIdleConnsPerHost: 10,               // Максимум на хост
					IdleConnTimeout:     90 * time.Second, // Время жизни соединения
					DisableCompression:  false,            // Включаем сжатие для экономии трафика
					ForceAttemptHTTP2:   true,             // Принудительно используем HTTP/2
					// Дополнительные настройки для стабильности
					MaxConnsPerHost:       100,               // Максимум соединений на хост
					ResponseHeaderTimeout: 120 * time.Second, // Увеличиваем таймаут заголовков ответа
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
			Timeout: 300 * time.Second, // Увеличиваем до 5 минут для генерации постов
			// ОПТИМИЗАЦИЯ: Пул соединений для лучшей производительности
			Transport: &http.Transport{
				MaxIdleConns:        50,               // Максимум неактивных соединений
				MaxIdleConnsPerHost: 10,               // Максимум на хост
				IdleConnTimeout:     90 * time.Second, // Время жизни соединения
				DisableCompression:  false,            // Включаем сжатие для экономии трафика
				ForceAttemptHTTP2:   true,             // Принудительно используем HTTP/2
				// Дополнительные настройки для стабильности
				MaxConnsPerHost:       100,               // Максимум соединений на хост
				ResponseHeaderTimeout: 120 * time.Second, // Увеличиваем таймаут заголовков ответа
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

	// Используем промпт для рерайта из prompts.json
	return dh.CreateContent("rewrite_post", originalText)
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
		MaxTokens:   2000,
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
		MaxTokens:   2000,
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

// CreateContent создает контент для различных платформ на основе промптов
func (dh *DeepSeekHandler) CreateContent(contentType string, originalText string) (string, error) {
	if dh.apiKey == "" {
		return "🔧 Функция создания контента временно недоступна", nil
	}

	// Загружаем промпты из JSON файла
	prompts, err := loadPrompts()
	if err != nil {
		return "", fmt.Errorf("ошибка загрузки промптов: %v", err)
	}

	// Получаем промпт для указанного типа контента
	promptConfig, exists := prompts[contentType]
	fmt.Printf(promptConfig.System)
	if !exists {
		return "", fmt.Errorf("неизвестный тип контента: %s", contentType)
	}

	// Формируем промпт с подстановкой текста
	prompt := strings.Replace(promptConfig.User, "{text}", originalText, -1)

	request := DeepSeekRequest{
		Model: "deepseek-chat",
		Messages: []DeepSeekMessage{
			{
				Role:    "system",
				Content: promptConfig.System,
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.8,
		MaxTokens:   2000,
	}

	response, err := dh.makeRequest(request)
	if err != nil {
		return "", fmt.Errorf("ошибка DeepSeek API: %v", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("пустой ответ от DeepSeek")
	}

	content := strings.TrimSpace(response.Choices[0].Message.Content)
	return content, nil
}

// CreateTelegramPost создает красивый пост для Telegram с хештегами
func (dh *DeepSeekHandler) CreateTelegramPost(originalText string) (string, error) {
	return dh.CreateContent("telegram_post", originalText)
}

// makeRequest выполняет HTTP запрос к DeepSeek API с retry логикой
func (dh *DeepSeekHandler) makeRequest(request DeepSeekRequest) (*DeepSeekResponse, error) {
	const maxRetries = 3
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		log.Printf("🔄 [DeepSeek] Попытка %d/%d", attempt, maxRetries)

		response, err := dh.makeSingleRequest(request)
		if err == nil {
			if attempt > 1 {
				log.Printf("✅ [DeepSeek] Успешно после %d попыток", attempt)
			}
			return response, nil
		}

		lastErr = err
		log.Printf("❌ [DeepSeek] Попытка %d неудачна: %v", attempt, err)

		// Если это не последняя попытка, ждем перед повтором
		if attempt < maxRetries {
			waitTime := time.Duration(attempt) * 2 * time.Second
			log.Printf("⏳ [DeepSeek] Ждем %v перед повтором...", waitTime)
			time.Sleep(waitTime)
		}
	}

	return nil, fmt.Errorf("все попытки исчерпаны, последняя ошибка: %v", lastErr)
}

// makeSingleRequest выполняет один HTTP запрос к DeepSeek API
func (dh *DeepSeekHandler) makeSingleRequest(request DeepSeekRequest) (*DeepSeekResponse, error) {
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
		monitoring.RecordError("api", "deepseek")
		return nil, fmt.Errorf("ошибка API: %s - %s", resp.Status, string(body))
	}

	var response DeepSeekResponse
	if err := json.Unmarshal(body, &response); err != nil {
		monitoring.RecordError("api", "deepseek")
		return nil, fmt.Errorf("ошибка парсинга ответа: %v", err)
	}

	// Записываем метрики использования токенов (примерные значения)
	// В реальном API ответе должны быть поля usage.prompt_tokens и usage.completion_tokens
	inputTokens := len(strings.Split(request.Messages[0].Content, " "))
	outputTokens := len(strings.Split(response.Choices[0].Message.Content, " "))

	monitoring.RecordDeepSeekTokens("input", inputTokens)
	monitoring.RecordDeepSeekTokens("output", outputTokens)
	monitoring.RecordDeepSeekTokens("total", inputTokens+outputTokens)

	return &response, nil
}

// PromptConfig представляет конфигурацию промпта для определенного типа контента
type PromptConfig struct {
	System string `json:"system"`
	User   string `json:"user"`
	Edit   struct {
		System string `json:"system"`
		User   string `json:"user"`
	} `json:"edit"`
}

// loadPrompts загружает промпты из JSON файла
func loadPrompts() (map[string]PromptConfig, error) {
	// Определяем путь к файлу промптов
	promptsPath := "internal/infrastructure/prompts/prompts.json"

	// Читаем файл
	data, err := os.ReadFile(promptsPath)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения файла промптов: %v", err)
	}

	// Парсим JSON
	var prompts map[string]PromptConfig
	if err := json.Unmarshal(data, &prompts); err != nil {
		return nil, fmt.Errorf("ошибка парсинга JSON промптов: %v", err)
	}

	return prompts, nil
}
