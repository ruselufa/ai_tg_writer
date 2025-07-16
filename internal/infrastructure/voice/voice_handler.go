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
	"ai_tg_writer/internal/infrastructure/lemon"
	"ai_tg_writer/internal/infrastructure/whisper"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

type VoiceHandler struct {
	bot             *tgbotapi.BotAPI
	whisperHandler  *whisper.WhisperHandler
	deepseekHandler *deepseek.DeepSeekHandler
	lemonHandler    *lemon.LemonHandler
}

func NewVoiceHandler(bot *tgbotapi.BotAPI) *VoiceHandler {
	return &VoiceHandler{
		bot:             bot,
		whisperHandler:  whisper.NewWhisperHandler(),
		deepseekHandler: deepseek.NewDeepSeekHandler(),
		lemonHandler:    lemon.NewLemonHandler(),
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
	oggFileName := fmt.Sprintf("%s_%d.oga", fileID, time.Now().Unix())
	oggFilePath := filepath.Join(audioDir, oggFileName)

	// Скачиваем файл
	resp, err := http.Get(file.Link(vh.bot.Token))
	if err != nil {
		return "", fmt.Errorf("ошибка скачивания файла: %v", err)
	}
	defer resp.Body.Close()

	// Создаем файл
	out, err := os.Create(oggFilePath)
	if err != nil {
		return "", fmt.Errorf("ошибка создания файла: %v", err)
	}
	defer out.Close()

	// Копируем данные
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", fmt.Errorf("ошибка сохранения файла: %v", err)
	}

	log.Printf("Ogg файл сохранен: %s", oggFilePath)
	// Конвертируем в .mp3
	mp3FilePath := filepath.Join(audioDir, fmt.Sprintf("%s.mp3", fileID))
	err = ffmpeg.Input(oggFilePath).
		Output(mp3FilePath, ffmpeg.KwArgs{
			"codec:a":  "libmp3lame",
			"qscale:a": "2",
			"loglevel": "quiet",
		}).
		Run()
	if err != nil {
		return "", fmt.Errorf("ошибка конвертации в mp3: %v", err)
	}

	log.Printf("MP3 файл создан: %s", mp3FilePath)
	return mp3FilePath, nil
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
	// TODO: Обработка аудио
	// transcriptionResp, err := vh.whisperHandler.TranscribeAudio(filePath)
	transcriptionResp, err := vh.lemonHandler.TranscribeAudio(filePath)
	if err != nil {
		return "", fmt.Errorf("ошибка отправки на транскрипцию: %v", err)
	}

	// log.Printf("Файл отправлен на транскрипцию, ID: %s, статус: %s", transcriptionResp.FileID, transcriptionResp.Status)

	// Ждем завершения транскрипции (максимум 5 минут)
	// transcribedText, err := vh.whisperHandler.WaitForCompletion(transcriptionResp.FileID, 5*time.Minute)
	// if err != nil {
	// 	return "", fmt.Errorf("ошибка ожидания транскрипции: %v", err)
	// }

	log.Printf("Транскрипция завершена2: %s", transcriptionResp.Text)
	return transcriptionResp.Text, nil
}

// TranscribeVoiceFile транскрибирует уже скачанный файл
func (vh *VoiceHandler) TranscribeVoiceFile(filePath string) (string, error) {
	// Отправляем на транскрипцию через Whisper
	log.Printf("Отправляем файл на транскрипцию: %s", filePath)
	// transcriptionResp, err := vh.whisperHandler.TranscribeAudio(filePath)
	transcriptionResp, err := vh.lemonHandler.TranscribeAudio(filePath)
	if err != nil {
		return "", fmt.Errorf("ошибка отправки на транскрипцию: %v", err)
	}

	// log.Printf("Файл отправлен на транскрипцию, ID: %s, статус: %s", transcriptionResp.FileID, transcriptionResp.Status)

	// Ждем завершения транскрипции (максимум 5 минут)
	// transcribedText, err := vh.whisperHandler.WaitForCompletion(transcriptionResp.FileID, 5*time.Minute)
	// if err != nil {
	// 	return "", fmt.Errorf("ошибка ожидания транскрипции: %v", err)
	// }

	log.Printf("Транскрипция завершена: %s", transcriptionResp.Text)
	return transcriptionResp.Text, nil
}

// GenerateTelegramPost генерирует красивый Telegram-пост на основе текста
func (vh *VoiceHandler) GenerateTelegramPost(text string) (string, error) {
	return vh.deepseekHandler.CreateTelegramPost(text)
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
