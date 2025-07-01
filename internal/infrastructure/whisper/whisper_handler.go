package whisper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type WhisperHandler struct {
	apiURL string
	client *http.Client
}

type TranscriptionResponse struct {
	FileID        string               `json:"file_id"`
	Status        string               `json:"status"`
	QueuePosition int                  `json:"queue_position"`
	Metrics       TranscriptionMetrics `json:"metrics"`
}

type TranscriptionStatus struct {
	Status         string               `json:"status"`
	CreatedAt      string               `json:"created_at"`
	InputPath      string               `json:"input_path"`
	OutputPath     string               `json:"output_path,omitempty"`
	CompletedAt    string               `json:"completed_at,omitempty"`
	ProcessingTime float64              `json:"processing_time,omitempty"`
	FileSize       int64                `json:"file_size,omitempty"`
	Error          string               `json:"error,omitempty"`
	QueuePosition  int                  `json:"queue_position,omitempty"`
	Metrics        TranscriptionMetrics `json:"metrics"`
}

type TranscriptionMetrics struct {
	QueueSize          int     `json:"queue_size"`
	AvgProcessingTime  float64 `json:"avg_processing_time"`
	AvgProcessingSpeed float64 `json:"avg_processing_speed"`
	FilesProcessed     int     `json:"files_processed"`
}

func NewWhisperHandler() *WhisperHandler {
	apiURL := os.Getenv("WHISPER_API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8001"
	}

	return &WhisperHandler{
		apiURL: apiURL,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// TranscribeAudio отправляет аудио файл на транскрипцию
func (wh *WhisperHandler) TranscribeAudio(audioPath string) (*TranscriptionResponse, error) {
	file, err := os.Open(audioPath)
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия файла: %v", err)
	}
	defer file.Close()

	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	part, err := writer.CreateFormFile("file", filepath.Base(audioPath))
	if err != nil {
		return nil, fmt.Errorf("ошибка создания формы: %v", err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return nil, fmt.Errorf("ошибка копирования файла: %v", err)
	}

	writer.Close()

	req, err := http.NewRequest("POST", wh.apiURL+"/transcribe", &requestBody)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %v", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := wh.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка отправки запроса: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ошибка API: %s - %s", resp.Status, string(body))
	}

	var response TranscriptionResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("ошибка парсинга ответа: %v", err)
	}

	return &response, nil
}

// GetStatus получает статус транскрипции
func (wh *WhisperHandler) GetStatus(fileID string) (*TranscriptionStatus, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/status/%s", wh.apiURL, fileID), nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %v", err)
	}

	resp, err := wh.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка отправки запроса: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ошибка API: %s - %s", resp.Status, string(body))
	}

	var status TranscriptionStatus
	if err := json.Unmarshal(body, &status); err != nil {
		return nil, fmt.Errorf("ошибка парсинга ответа: %v", err)
	}

	return &status, nil
}

// GetMetrics получает метрики сервиса транскрипции
func (wh *WhisperHandler) GetMetrics() (*TranscriptionMetrics, error) {
	req, err := http.NewRequest("GET", wh.apiURL+"/metrics", nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %v", err)
	}

	resp, err := wh.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка отправки запроса: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ошибка API: %s - %s", resp.Status, string(body))
	}

	var metrics TranscriptionMetrics
	if err := json.Unmarshal(body, &metrics); err != nil {
		return nil, fmt.Errorf("ошибка парсинга ответа: %v", err)
	}

	return &metrics, nil
}

// DownloadResult скачивает результат транскрипции
func (wh *WhisperHandler) DownloadResult(fileID string) (string, error) {
	maxRetries := 3
	retryDelay := 2 * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		log.Printf("Попытка %d/%d скачивания результата для %s", attempt, maxRetries, fileID)

		req, err := http.NewRequest("GET", fmt.Sprintf("%s/download/%s", wh.apiURL, fileID), nil)
		if err != nil {
			return "", fmt.Errorf("ошибка создания запроса: %v", err)
		}

		req.Header.Set("Connection", "keep-alive")
		req.Header.Set("Keep-Alive", "timeout=30, max=1000")

		resp, err := wh.client.Do(req)
		if err != nil {
			log.Printf("Попытка %d/%d неудачна для %s: %v", attempt, maxRetries, fileID, err)
			if attempt == maxRetries {
				return "", fmt.Errorf("ошибка при скачивании результата после %d попыток: %v", maxRetries, err)
			}
			time.Sleep(retryDelay * time.Duration(attempt))
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("Попытка %d/%d неудачна для %s: статус %s", attempt, maxRetries, fileID, resp.Status)
			if attempt == maxRetries {
				return "", fmt.Errorf("ошибка при скачивании результата после %d попыток: статус %s", maxRetries, resp.Status)
			}
			time.Sleep(retryDelay * time.Duration(attempt))
			continue
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Попытка %d/%d неудачна для %s: ошибка чтения %v", attempt, maxRetries, fileID, err)
			if attempt == maxRetries {
				return "", fmt.Errorf("ошибка при скачивании результата после %d попыток: %v", maxRetries, err)
			}
			time.Sleep(retryDelay * time.Duration(attempt))
			continue
		}

		log.Printf("Результат успешно скачан для %s", fileID)
		return string(body), nil
	}

	return "", fmt.Errorf("неожиданная ошибка при скачивании результата")
}

// WaitForCompletion ждет завершения транскрипции
func (wh *WhisperHandler) WaitForCompletion(fileID string, maxWaitTime time.Duration) (string, error) {
	startTime := time.Now()
	checkInterval := 2 * time.Second

	for time.Since(startTime) < maxWaitTime {
		status, err := wh.GetStatus(fileID)
		if err != nil {
			log.Printf("Ошибка получения статуса для %s: %v", fileID, err)
			time.Sleep(checkInterval)
			continue
		}

		switch status.Status {
		case "completed":
			return wh.DownloadResult(fileID)
		case "error":
			return "", fmt.Errorf("ошибка транскрипции: %s", status.Error)
		case "queued", "processing":
			log.Printf("Статус для %s: %s (позиция в очереди: %d)", fileID, status.Status, status.QueuePosition)
			time.Sleep(checkInterval)
		default:
			log.Printf("Неизвестный статус для %s: %s", fileID, status.Status)
			time.Sleep(checkInterval)
		}
	}

	return "", fmt.Errorf("превышено время ожидания транскрипции (%v)", maxWaitTime)
}
