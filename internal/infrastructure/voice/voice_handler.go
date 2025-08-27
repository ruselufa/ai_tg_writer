package voice

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"ai_tg_writer/internal/infrastructure/database"
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
	postHistoryRepo *database.PostHistoryRepository // Добавляем репозиторий для истории
}

func NewVoiceHandler(bot *tgbotapi.BotAPI, postHistoryRepo *database.PostHistoryRepository) *VoiceHandler {
	return &VoiceHandler{
		bot:             bot,
		whisperHandler:  whisper.NewWhisperHandler(),
		deepseekHandler: deepseek.NewDeepSeekHandler(),
		lemonHandler:    lemon.NewLemonHandler(),
		postHistoryRepo: postHistoryRepo,
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

// ProcessVoiceMessage обрабатывает голосовое сообщение с логированием
func (vh *VoiceHandler) ProcessVoiceMessage(message *tgbotapi.Message) (string, error) {
	voiceSentAt := time.Now().UTC()

	// Создаем запись в истории
	history := &database.PostHistory{
		UserID:        message.From.ID,
		VoiceText:     "", // Пока пустой, заполним после транскрипции
		VoiceFileID:   message.Voice.FileID,
		VoiceDuration: message.Voice.Duration,
		VoiceFileSize: message.Voice.FileSize,
		VoiceSentAt:   voiceSentAt,
		AIModel:       "deepseek",
	}

	// Сохраняем начальную запись
	var historyID int
	if vh.postHistoryRepo != nil {
		err := vh.postHistoryRepo.CreatePostHistory(history)
		if err != nil {
			log.Printf("Ошибка создания записи в истории: %v", err)
			// Продолжаем работу, не прерываем из-за ошибки логирования
		} else {
			historyID = history.ID
		}
	}

	// Скачиваем файл
	filePath, err := vh.DownloadVoiceFile(message.Voice.FileID)
	if err != nil {
		return "", err
	}

	// Удаляем временный файл после обработки
	defer os.Remove(filePath)

	// Отправляем на транскрипцию
	whisperStart := time.Now().UTC()
	log.Printf("Отправляем файл на транскрипцию: %s", filePath)

	transcriptionResp, err := vh.lemonHandler.TranscribeAudio(filePath)
	if err != nil {
		return "", fmt.Errorf("ошибка отправки на транскрипцию: %v", err)
	}

	whisperDuration := time.Since(whisperStart)
	voiceReceivedAt := time.Now().UTC()

	log.Printf("Транскрипция завершена: %s", transcriptionResp.Text)

	// Обновляем историю с результатом транскрипции
	if historyID > 0 && vh.postHistoryRepo != nil {
		whisperDurationMs := int(whisperDuration.Milliseconds())
		err = vh.postHistoryRepo.UpdateVoiceTranscription(historyID, transcriptionResp.Text, &voiceReceivedAt, &whisperDurationMs)
		if err != nil {
			log.Printf("Ошибка обновления истории транскрипции: %v", err)
		} else {
			log.Printf("История транскрипции обновлена для ID: %d", historyID)
		}
	}

	// Сохраняем ID истории в сообщении для последующего использования
	if historyID > 0 {
		// TODO: Передать historyID в message_handler для сохранения в состоянии пользователя
		log.Printf("Создана запись в истории с ID: %d", historyID)
	}

	return transcriptionResp.Text, nil
}

// TranscribeVoiceFile транскрибирует уже скачанный файл с логированием
func (vh *VoiceHandler) TranscribeVoiceFile(filePath string, userID int64, fileID string, duration int, fileSize int) (string, int, error) {
	voiceSentAt := time.Now().UTC()

	// Создаем запись в истории
	history := &database.PostHistory{
		UserID:        userID,
		VoiceText:     "", // Пока пустой, заполним после транскрипции
		VoiceFileID:   fileID,
		VoiceDuration: duration, // Используем переданную длительность
		VoiceFileSize: fileSize, // Используем переданный размер файла
		VoiceSentAt:   voiceSentAt,
		AIModel:       "deepseek",
	}

	// Сохраняем начальную запись
	var historyID int
	if vh.postHistoryRepo != nil {
		err := vh.postHistoryRepo.CreatePostHistory(history)
		if err != nil {
			log.Printf("Ошибка создания записи в истории: %v", err)
			// Продолжаем работу, не прерываем из-за ошибки логирования
		} else {
			historyID = history.ID
		}
	}

	// Отправляем на транскрипцию
	whisperStart := time.Now().UTC()
	log.Printf("Отправляем файл на транскрипцию: %s", filePath)

	transcriptionResp, err := vh.lemonHandler.TranscribeAudio(filePath)
	if err != nil {
		return "", 0, err
	}

	whisperDuration := time.Since(whisperStart)
	voiceReceivedAt := time.Now().UTC()

	log.Printf("Транскрипция завершена: %s", transcriptionResp.Text)

	// Обновляем историю с результатом транскрипции
	if historyID > 0 && vh.postHistoryRepo != nil {
		whisperDurationMs := int(whisperDuration.Milliseconds())
		err = vh.postHistoryRepo.UpdateVoiceTranscription(historyID, transcriptionResp.Text, &voiceReceivedAt, &whisperDurationMs)
		if err != nil {
			log.Printf("Ошибка обновления истории транскрипции: %v", err)
		} else {
			log.Printf("История транскрипции обновлена для ID: %d", historyID)
		}
	}

	return transcriptionResp.Text, historyID, nil
}

// GenerateTelegramPost генерирует красивый Telegram-пост с логированием
func (vh *VoiceHandler) GenerateTelegramPost(text string, userID int64, historyID int) (string, error) {
	aiSentAt := time.Now().UTC()
	aiStart := time.Now().UTC()

	// Обновляем время отправки в AI
	if historyID > 0 && vh.postHistoryRepo != nil {
		err := vh.postHistoryRepo.UpdateAISentAt(historyID, &aiSentAt)
		if err != nil {
			log.Printf("Ошибка обновления времени отправки в AI: %v", err)
		} else {
			log.Printf("Время отправки в AI обновлено для ID: %d", historyID)
		}
	}

	// Генерируем пост
	response, err := vh.deepseekHandler.CreateTelegramPost(text)
	if err != nil {
		return "", err
	}

	aiDuration := time.Since(aiStart)
	aiReceivedAt := time.Now().UTC()

	// Обновляем историю с результатом AI
	if historyID > 0 && vh.postHistoryRepo != nil {
		aiDurationMs := int(aiDuration.Milliseconds())

		// Сначала обновляем ответ AI
		err = vh.postHistoryRepo.UpdateAIResponse(historyID, response, &aiReceivedAt, &aiDurationMs, nil)
		if err != nil {
			log.Printf("Ошибка обновления истории AI: %v", err)
		} else {
			// Теперь обновляем общее время обработки
			// Получаем whisper_duration_ms из БД и рассчитываем общее время
			history, err := vh.postHistoryRepo.GetPostHistoryByID(historyID)
			if err != nil {
				log.Printf("Ошибка получения истории для расчета времени: %v", err)
			} else {
				// Вычисляем общее время: whisper + AI
				var totalDurationMs int
				if history.WhisperDurationMs != nil {
					totalDurationMs = *history.WhisperDurationMs + aiDurationMs
				} else {
					totalDurationMs = aiDurationMs
				}

				err = vh.postHistoryRepo.UpdateProcessingDuration(historyID, totalDurationMs)
				if err != nil {
					log.Printf("Ошибка обновления общего времени обработки: %v", err)
				} else {
					whisperMs := 0
					if history.WhisperDurationMs != nil {
						whisperMs = *history.WhisperDurationMs
					}
					log.Printf("Общее время обработки обновлено: %d мс (whisper: %d + AI: %d)",
						totalDurationMs, whisperMs, aiDurationMs)
				}
			}
		}

		log.Printf("История AI обновлена для ID: %d", historyID)
	}

	return response, nil
}

// MarkPostAsSaved отмечает пост как сохраненный в истории
func (vh *VoiceHandler) MarkPostAsSaved(historyID int) error {
	if vh.postHistoryRepo != nil {
		err := vh.postHistoryRepo.MarkAsSaved(historyID)
		if err != nil {
			log.Printf("Ошибка отметки поста как сохраненного: %v", err)
			return err
		}
		log.Printf("Пост с historyID %d отмечен как сохраненный", historyID)
		return nil
	}
	return fmt.Errorf("postHistoryRepo не инициализирован")
}

// AddVoiceToHistory добавляет голосовое сообщение к существующей записи истории
func (vh *VoiceHandler) AddVoiceToHistory(historyID int, voiceText string, voiceDuration int, voiceFileSize int) error {
	if vh.postHistoryRepo != nil {
		err := vh.postHistoryRepo.AddVoiceToHistory(historyID, voiceText, voiceDuration, voiceFileSize)
		if err != nil {
			log.Printf("Ошибка добавления голосового сообщения к истории: %v", err)
			return err
		}
		log.Printf("Голосовое сообщение добавлено к истории ID: %d", historyID)
		return nil
	}
	return fmt.Errorf("postHistoryRepo не инициализирован")
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
