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
	fmt.Println("Запуск транскрипции файла: ", audioPath)
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
	fmt.Println("Файл создан")
	_, err = io.Copy(part, file)
	if err != nil {
		return nil, fmt.Errorf("ошибка копирования файла: %v", err)
	}
	fmt.Println("Файл скопирован")
	// Добавляем параметр языка
	writer.WriteField("language", "russian")
	writer.WriteField("format", "json")

	writer.Close()
	fmt.Println("Запрос создан")
	req, err := http.NewRequest("POST", lh.apiURL+"/v1/audio/transcriptions", &requestBody)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %v", err)
	}
	fmt.Println("Запрос создан")
	req.Header.Set("Authorization", "Bearer "+lh.apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	fmt.Println("Запрос отправлен")
	resp, err := lh.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка отправки запроса: %v", err)
	}
	defer resp.Body.Close()
	fmt.Println("Ответ получен")
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %v", err)
	}
	fmt.Println("Ответ прочитан")
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ошибка API: %s - %s", resp.Status, string(body))
	}
	fmt.Println("Ответ проверен")
	var response TranscriptionResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("ошибка парсинга ответа: %v", err)
	}
	fmt.Println("Ответ парсен")
	fmt.Println(response.Text)
	return &response, nil
}
