package lemon

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewLemonHandler(t *testing.T) {
	// Сохраняем оригинальные значения переменных окружения
	originalAPIKey := os.Getenv("LEMON_API_KEY")
	originalAPIURL := os.Getenv("LEMON_API_URL")
	defer func() {
		os.Setenv("LEMON_API_KEY", originalAPIKey)
		os.Setenv("LEMON_API_URL", originalAPIURL)
	}()

	// Тест 1: Успешное создание хендлера
	os.Setenv("LEMON_API_KEY", "test-api-key")
	os.Setenv("LEMON_API_URL", "https://api.lemon.com")

	handler := NewLemonHandler()

	if handler.apiKey != "test-api-key" {
		t.Errorf("Ожидался apiKey 'test-api-key', получен '%s'", handler.apiKey)
	}

	if handler.apiURL != "https://api.lemon.com" {
		t.Errorf("Ожидался apiURL 'https://api.lemon.com', получен '%s'", handler.apiURL)
	}

	if handler.client == nil {
		t.Error("HTTP клиент не должен быть nil")
	}

	// Проверяем timeout клиента
	if handler.client.Timeout != 300*time.Second {
		t.Errorf("Ожидался timeout 300 секунд, получен %v", handler.client.Timeout)
	}
}

func TestLemonHandler_TranscribeAudio_Success(t *testing.T) {
	// Создаем временный аудио файл для тестирования
	tempDir := t.TempDir()
	audioPath := filepath.Join(tempDir, "test_audio.mp3")

	// Создаем простой тестовый файл
	testData := []byte("fake audio data")
	err := os.WriteFile(audioPath, testData, 0644)
	if err != nil {
		t.Fatalf("Не удалось создать тестовый файл: %v", err)
	}

	// Создаем mock сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем метод и путь
		if r.Method != "POST" {
			t.Errorf("Ожидался метод POST, получен %s", r.Method)
		}
		if r.URL.Path != "/v1/audio/transcriptions" {
			t.Errorf("Ожидался путь /v1/audio/transcriptions, получен %s", r.URL.Path)
		}

		// Проверяем заголовки
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer test-api-key" {
			t.Errorf("Ожидался заголовок Authorization 'Bearer test-api-key', получен '%s'", authHeader)
		}

		contentType := r.Header.Get("Content-Type")
		if contentType == "" {
			t.Error("Content-Type заголовок отсутствует")
		}

		// Проверяем multipart форму
		err := r.ParseMultipartForm(32 << 20)
		if err != nil {
			t.Errorf("Ошибка парсинга multipart формы: %v", err)
		}

		// Проверяем поля формы
		language := r.FormValue("language")
		if language != "russian" {
			t.Errorf("Ожидался язык 'russian', получен '%s'", language)
		}

		format := r.FormValue("format")
		if format != "json" {
			t.Errorf("Ожидался формат 'json', получен '%s'", format)
		}

		// Проверяем наличие файла
		file, header, err := r.FormFile("file")
		if err != nil {
			t.Errorf("Ошибка получения файла из формы: %v", err)
		}
		defer file.Close()

		if header.Filename != "test_audio.mp3" {
			t.Errorf("Ожидалось имя файла 'test_audio.mp3', получено '%s'", header.Filename)
		}

		// Читаем содержимое файла
		fileContent, err := io.ReadAll(file)
		if err != nil {
			t.Errorf("Ошибка чтения файла: %v", err)
		}

		if !bytes.Equal(fileContent, testData) {
			t.Error("Содержимое файла не совпадает с ожидаемым")
		}

		// Возвращаем успешный ответ
		response := TranscriptionResponse{
			Text: "Это тестовая транскрипция",
		}
		jsonResponse, _ := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)
	}))
	defer server.Close()

	// Создаем хендлер с URL mock сервера
	handler := &LemonHandler{
		apiKey: "test-api-key",
		apiURL: server.URL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// Выполняем транскрипцию
	response, err := handler.TranscribeAudio(audioPath)
	if err != nil {
		t.Fatalf("Ошибка транскрипции: %v", err)
	}

	if response.Text != "Это тестовая транскрипция" {
		t.Errorf("Ожидался текст 'Это тестовая транскрипция', получен '%s'", response.Text)
	}
}

func TestLemonHandler_TranscribeAudio_FileNotFound(t *testing.T) {
	handler := &LemonHandler{
		apiKey: "test-api-key",
		apiURL: "https://api.lemon.com",
		client: &http.Client{},
	}

	// Пытаемся транскрибировать несуществующий файл
	_, err := handler.TranscribeAudio("/path/to/nonexistent/file.mp3")
	if err == nil {
		t.Error("Ожидалась ошибка для несуществующего файла")
	}

	if err.Error() != "ошибка создания файла: open /path/to/nonexistent/file.mp3: no such file or directory" {
		t.Errorf("Неожиданная ошибка: %v", err)
	}
}

func TestLemonHandler_TranscribeAudio_APIError(t *testing.T) {
	// Создаем временный аудио файл
	tempDir := t.TempDir()
	audioPath := filepath.Join(tempDir, "test_audio.mp3")

	testData := []byte("fake audio data")
	err := os.WriteFile(audioPath, testData, 0644)
	if err != nil {
		t.Fatalf("Не удалось создать тестовый файл: %v", err)
	}

	// Создаем mock сервер, который возвращает ошибку
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Internal server error"}`))
	}))
	defer server.Close()

	handler := &LemonHandler{
		apiKey: "test-api-key",
		apiURL: server.URL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// Выполняем транскрипцию
	_, err = handler.TranscribeAudio(audioPath)
	if err == nil {
		t.Error("Ожидалась ошибка API")
	}

	expectedError := "ошибка API: 500 Internal Server Error - {\"error\": \"Internal server error\"}"
	if err.Error() != expectedError {
		t.Errorf("Ожидалась ошибка '%s', получена '%s'", expectedError, err.Error())
	}
}

func TestLemonHandler_TranscribeAudio_InvalidJSON(t *testing.T) {
	// Создаем временный аудио файл
	tempDir := t.TempDir()
	audioPath := filepath.Join(tempDir, "test_audio.mp3")

	testData := []byte("fake audio data")
	err := os.WriteFile(audioPath, testData, 0644)
	if err != nil {
		t.Fatalf("Не удалось создать тестовый файл: %v", err)
	}

	// Создаем mock сервер, который возвращает невалидный JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"invalid": json}`)) // Невалидный JSON
	}))
	defer server.Close()

	handler := &LemonHandler{
		apiKey: "test-api-key",
		apiURL: server.URL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// Выполняем транскрипцию
	_, err = handler.TranscribeAudio(audioPath)
	if err == nil {
		t.Error("Ожидалась ошибка парсинга JSON")
	}

	if err.Error() != "ошибка парсинга ответа: invalid character 'j' looking for beginning of value" {
		t.Errorf("Неожиданная ошибка парсинга: %v", err)
	}
}

func TestLemonHandler_TranscribeAudio_NetworkError(t *testing.T) {
	// Создаем временный аудио файл
	tempDir := t.TempDir()
	audioPath := filepath.Join(tempDir, "test_audio.mp3")

	testData := []byte("fake audio data")
	err := os.WriteFile(audioPath, testData, 0644)
	if err != nil {
		t.Fatalf("Не удалось создать тестовый файл: %v", err)
	}

	// Создаем mock сервер, который не отвечает (симулируем сетевую ошибку)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Закрываем соединение сразу, не отвечая
		hj, ok := w.(http.Hijacker)
		if ok {
			conn, _, _ := hj.Hijack()
			conn.Close()
		}
	}))
	defer server.Close()

	handler := &LemonHandler{
		apiKey: "test-api-key",
		apiURL: server.URL,
		client: &http.Client{
			Timeout: 1 * time.Second, // Короткий timeout для быстрого теста
		},
	}

	// Выполняем транскрипцию
	_, err = handler.TranscribeAudio(audioPath)
	if err == nil {
		t.Error("Ожидалась ошибка сети")
	}

	// Проверяем, что ошибка содержит информацию о проблеме с сетью
	expectedErrorPrefix := "ошибка отправки запроса:"
	if !strings.Contains(err.Error(), expectedErrorPrefix) {
		t.Errorf("Неожиданная ошибка сети: %v", err)
	}
}

// Benchmark тест для измерения производительности
func BenchmarkLemonHandler_TranscribeAudio(b *testing.B) {
	// Создаем временный аудио файл
	tempDir := b.TempDir()
	audioPath := filepath.Join(tempDir, "benchmark_audio.mp3")

	testData := make([]byte, 1024) // 1KB тестовых данных
	for i := range testData {
		testData[i] = byte(i % 256)
	}

	err := os.WriteFile(audioPath, testData, 0644)
	if err != nil {
		b.Fatalf("Не удалось создать тестовый файл: %v", err)
	}

	// Создаем mock сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := TranscriptionResponse{
			Text: "Benchmark транскрипция",
		}
		jsonResponse, _ := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)
	}))
	defer server.Close()

	handler := &LemonHandler{
		apiKey: "benchmark-api-key",
		apiURL: server.URL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := handler.TranscribeAudio(audioPath)
		if err != nil {
			b.Fatalf("Ошибка в benchmark: %v", err)
		}
	}
}
