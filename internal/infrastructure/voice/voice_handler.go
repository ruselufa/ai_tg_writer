package voice

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"ai_tg_writer/internal/infrastructure/deepseek"
	"ai_tg_writer/internal/infrastructure/whisper"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type VoiceHandler struct {
	bot             *tgbotapi.BotAPI
	whisperHandler  *whisper.WhisperHandler
	deepseekHandler *deepseek.DeepSeekHandler
}

func NewVoiceHandler(bot *tgbotapi.BotAPI) *VoiceHandler {
	return &VoiceHandler{
		bot:             bot,
		whisperHandler:  whisper.NewWhisperHandler(),
		deepseekHandler: deepseek.NewDeepSeekHandler(),
	}
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

	// Удаляем временный файл после обработки
	defer os.Remove(filePath)

	// Отправляем на транскрипцию через Whisper
	log.Printf("Отправляем файл на транскрипцию: %s", filePath)
	transcriptionResp, err := vh.whisperHandler.TranscribeAudio(filePath)
	if err != nil {
		return "", fmt.Errorf("ошибка отправки на транскрипцию: %v", err)
	}

	log.Printf("Файл отправлен на транскрипцию, ID: %s, статус: %s", transcriptionResp.FileID, transcriptionResp.Status)

	// Ждем завершения транскрипции (максимум 5 минут)
	transcribedText, err := vh.whisperHandler.WaitForCompletion(transcriptionResp.FileID, 5*time.Minute)
	if err != nil {
		return "", fmt.Errorf("ошибка ожидания транскрипции: %v", err)
	}

	log.Printf("Транскрипция завершена: %s", transcribedText)

	// Переписываем текст с помощью DeepSeek
	rewrittenText, err := vh.deepseekHandler.RewriteText(transcribedText)
	if err != nil {
		log.Printf("Ошибка переписывания текста: %v, возвращаем исходный текст", err)
		return transcribedText, nil
	}

	return rewrittenText, nil
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
