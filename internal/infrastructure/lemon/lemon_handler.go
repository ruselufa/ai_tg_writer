package lemon

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type LemonHandler struct {
	apiKey string
	apiURL string
	client *http.Client
}

type TranscriptionResponse struct {
	Text string `json:"text"`
}

func NewLemonHandler() *LemonHandler {
	apiKey := os.Getenv("LEMON_API_KEY")
	apiURL := os.Getenv("LEMON_API_URL")

	return &LemonHandler{
		apiKey: apiKey,
		apiURL: apiURL,
		client: &http.Client{
			Timeout: 300 * time.Second,
		},
	}
}

func (lh *LemonHandler) TranscribeAudio(audioPath string) (*TranscriptionResponse, error) {
	file, err := os.Open(audioPath)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания файла: %v", err)
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
	// Добавляем параметр языка
	writer.WriteField("language", "russian")
	writer.WriteField("format", "json")

	writer.Close()
	req, err := http.NewRequest("POST", lh.apiURL+"/v1/audio/transcriptions", &requestBody)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+lh.apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := lh.client.Do(req)
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
