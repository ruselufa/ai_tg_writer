package openai

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

type OpenAIHandler struct {
	client *openai.Client
}

func NewOpenAIHandler() *OpenAIHandler {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Println("Предупреждение: OPENAI_API_KEY не установлен")
		return &OpenAIHandler{client: nil}
	}

	client := openai.NewClient(apiKey)
	return &OpenAIHandler{client: client}
}

// RewriteText переписывает текст с помощью ИИ
func (oh *OpenAIHandler) RewriteText(originalText string) (string, error) {
	if oh.client == nil {
		return "🔧 Функция переписывания текста временно недоступна", nil
	}

	// Промпт для переписывания текста
	prompt := fmt.Sprintf(`Перепиши следующий текст, сделав его более читаемым, структурированным и профессиональным. 
Сохрани основную мысль, но улучши стиль, грамматику и пунктуацию:

"%s"

Переписанный текст:`, originalText)

	resp, err := oh.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			MaxTokens:   1000,
			Temperature: 0.7,
		},
	)

	if err != nil {
		return "", fmt.Errorf("ошибка OpenAI API: %v", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("пустой ответ от OpenAI")
	}

	rewrittenText := strings.TrimSpace(resp.Choices[0].Message.Content)
	return rewrittenText, nil
}

// ImproveText улучшает качество текста
func (oh *OpenAIHandler) ImproveText(text string, style string) (string, error) {
	if oh.client == nil {
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

	resp, err := oh.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			MaxTokens:   1000,
			Temperature: 0.7,
		},
	)

	if err != nil {
		return "", fmt.Errorf("ошибка OpenAI API: %v", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("пустой ответ от OpenAI")
	}

	improvedText := strings.TrimSpace(resp.Choices[0].Message.Content)
	return improvedText, nil
}

// SummarizeText создает краткое изложение текста
func (oh *OpenAIHandler) SummarizeText(text string) (string, error) {
	if oh.client == nil {
		return "🔧 Функция создания краткого изложения временно недоступна", nil
	}

	prompt := fmt.Sprintf(`Создай краткое изложение следующего текста, выделив основные мысли:

"%s"

Краткое изложение:`, text)

	resp, err := oh.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			MaxTokens:   500,
			Temperature: 0.5,
		},
	)

	if err != nil {
		return "", fmt.Errorf("ошибка OpenAI API: %v", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("пустой ответ от OpenAI")
	}

	summary := strings.TrimSpace(resp.Choices[0].Message.Content)
	return summary, nil
}
