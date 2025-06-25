package voice

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type VoiceHandler struct {
	bot *tgbotapi.BotAPI
}

func NewVoiceHandler(bot *tgbotapi.BotAPI) *VoiceHandler {
	return &VoiceHandler{bot: bot}
}

// DownloadVoiceFile скачивает голосовое сообщение
func (vh *VoiceHandler) DownloadVoiceFile(fileID string) (string, error) {
	// Получаем информацию о файле
	file, err := vh.bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		return "", fmt.Errorf("ошибка получения файла: %v", err)
	}

	// Создаем директорию для аудио файлов
	audioDir := "audio"
	if err := os.MkdirAll(audioDir, 0755); err != nil {
		return "", fmt.Errorf("ошибка создания директории: %v", err)
	}

	// Формируем имя файла
	fileName := fmt.Sprintf("%s_%d.oga", fileID, time.Now().Unix())
	filePath := filepath.Join(audioDir, fileName)

	// Скачиваем файл
	resp, err := http.Get(file.Link(vh.bot.Token))
	if err != nil {
		return "", fmt.Errorf("ошибка скачивания файла: %v", err)
	}
	defer resp.Body.Close()

	// Создаем файл
	out, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("ошибка создания файла: %v", err)
	}
	defer out.Close()

	// Копируем данные
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", fmt.Errorf("ошибка сохранения файла: %v", err)
	}

	log.Printf("Файл сохранен: %s", filePath)
	return filePath, nil
}

// ProcessVoiceMessage обрабатывает голосовое сообщение
func (vh *VoiceHandler) ProcessVoiceMessage(message *tgbotapi.Message) (string, error) {
	// Скачиваем файл
	filePath, err := vh.DownloadVoiceFile(message.Voice.FileID)
	if err != nil {
		return "", err
	}

	// TODO: Здесь будет логика распознавания речи
	// Пока возвращаем заглушку
	recognizedText := "🔧 Распознавание речи находится в разработке. Скоро будет доступно!"

	// Удаляем временный файл
	defer os.Remove(filePath)

	return recognizedText, nil
}

// CleanupOldFiles удаляет старые аудио файлы
func (vh *VoiceHandler) CleanupOldFiles() error {
	audioDir := "audio"
	files, err := os.ReadDir(audioDir)
	if err != nil {
		return err
	}

	// Удаляем файлы старше 1 часа
	cutoff := time.Now().Add(-time.Hour)
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := filepath.Join(audioDir, file.Name())
		info, err := os.Stat(filePath)
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoff) {
			if err := os.Remove(filePath); err != nil {
				log.Printf("Ошибка удаления файла %s: %v", filePath, err)
			} else {
				log.Printf("Удален старый файл: %s", filePath)
			}
		}
	}

	return nil
}
