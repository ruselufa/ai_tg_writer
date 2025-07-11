package bot

import (
	"ai_tg_writer/internal/infrastructure/voice"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// MessageHandler обрабатывает сообщения
type MessageHandler struct {
	stateManager *StateManager
	voiceHandler *voice.VoiceHandler
}

// NewMessageHandler создает новый обработчик сообщений
func NewMessageHandler(stateManager *StateManager, voiceHandler *voice.VoiceHandler) *MessageHandler {
	return &MessageHandler{
		stateManager: stateManager,
		voiceHandler: voiceHandler,
	}
}

// HandleMessage обрабатывает входящие сообщения
func (mh *MessageHandler) HandleMessage(bot *Bot, message *tgbotapi.Message) {
	userID := message.From.ID

	// Проверяем, ожидаем ли голосовое сообщение
	state := mh.stateManager.GetState(userID)
	if !state.WaitingForVoice {
		return
	}

	// Обрабатываем голосовое сообщение
	if message.Voice != nil {
		mh.handleVoiceMessage(bot, message)
	}
}

// handleVoiceMessage обрабатывает голосовое сообщение
func (mh *MessageHandler) handleVoiceMessage(bot *Bot, message *tgbotapi.Message) {
	userID := message.From.ID

	// Получаем состояние пользователя
	state := mh.stateManager.GetState(userID)
	log.Printf("[DEBUG] handleVoiceMessage вызван, WaitingForVoice=%v, ApprovalStatus=%s", state.WaitingForVoice, state.ApprovalStatus)

	// Скачиваем файл
	filePath, err := mh.voiceHandler.DownloadVoiceFile(message.Voice.FileID)
	if err != nil {
		log.Printf("Ошибка скачивания файла: %v", err)
		msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка обработки голосового сообщения")
		msg.ReplyToMessageID = message.MessageID
		bot.Send(msg)
		return
	}

	// Определяем, в каком режиме мы находимся
	if state.ApprovalStatus == "editing" {
		// Режим редактирования - добавляем в PendingEdits
		if state.PendingEdits == nil {
			state.PendingEdits = make(map[string]*VoiceTranscription)
		}

		mh.stateManager.AddPendingEdit(userID, message.MessageID, message.Voice.FileID)

		// Обновляем путь к файлу и статус
		if voice, ok := state.PendingEdits[message.Voice.FileID]; ok {
			voice.FilePath = filePath
			voice.Status = "pending"
			voice.Text = ""
		}

		log.Printf("[DEBUG] PendingEdits после добавления: %+v", state.PendingEdits)

		// Отправляем сообщение с кнопками для редактирования
		msg := tgbotapi.NewMessage(message.Chat.ID, "✅ Правки приняты. Хотите добавить ещё правки или применить изменения?")
		keyboard := bot.CreateEditContinueKeyboard()
		msg.ReplyMarkup = keyboard
		bot.Send(msg)
	} else {
		// Обычный режим - добавляем в PendingVoices
		if state.PendingVoices == nil {
			state.PendingVoices = make(map[string]*VoiceTranscription)
		}

		// Добавляем сообщение в очередь
		mh.stateManager.AddPendingVoice(userID, message.MessageID, message.Voice.FileID)

		// Обновляем путь к файлу и статус
		if voice, ok := state.PendingVoices[message.Voice.FileID]; ok {
			voice.FilePath = filePath
			voice.Status = "pending"
			voice.Text = ""
		}

		// Логируем текущее состояние PendingVoices
		log.Printf("[DEBUG] PendingVoices после добавления: %+v", state.PendingVoices)

		// Отправляем сообщение с кнопками
		msg := tgbotapi.NewMessage(message.Chat.ID, "✅ Принято. Хотите продолжить диктовку или уже начинать создание текста?")
		keyboard := bot.CreateContinueKeyboard()
		msg.ReplyMarkup = keyboard
		bot.Send(msg)
	}
}
