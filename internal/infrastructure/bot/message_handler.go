package bot

import (
	"ai_tg_writer/internal/infrastructure/voice"
	"log"
	"regexp"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// MessageHandler обрабатывает сообщения
type MessageHandler struct {
	stateManager  *StateManager
	voiceHandler  *voice.VoiceHandler
	inlineHandler *InlineHandler
}

// NewMessageHandler создает новый обработчик сообщений
func NewMessageHandler(stateManager *StateManager, voiceHandler *voice.VoiceHandler, inlineHandler *InlineHandler) *MessageHandler {
	return &MessageHandler{
		stateManager:  stateManager,
		voiceHandler:  voiceHandler,
		inlineHandler: inlineHandler,
	}
}

// HandleMessage обрабатывает входящие сообщения
// Возвращает true, если сообщение было обработано
func (mh *MessageHandler) HandleMessage(bot *Bot, message *tgbotapi.Message) bool {
	userID := message.From.ID
	state := mh.stateManager.GetState(userID)

	if state.WaitingForEmail && message.Text != "" {
		email := strings.TrimSpace(message.Text)
		if mh.isValidEmail(email) {
			// save email
			err := bot.DB.UpdateUserEmail(userID, email)
			if err != nil {
				log.Printf("Ошибка сохранения email: %v", err)
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка сохранения email. Попробуйте позже."))
				return true
			}
			state.WaitingForEmail = false
			log.Printf("Email сохранён для пользователя %d: %s", userID, email)

			// Отправляем сообщение об успешном сохранении email
			successMsg := tgbotapi.NewMessage(message.Chat.ID, "✅ E-mail сохранён! Переходим к оформлению подписки...")
			bot.Send(successMsg)

			// Показываем экран оформления подписки напрямую
			mh.stateManager.UpdateStep(userID, "idle")
			mh.showSubscriptionPurchaseScreen(bot, message.Chat.ID, userID)
			return true // сообщение обработано
		}
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Неверный формат e-mail. Пример: user@example.com\n\nИли используйте /start для отмены."))
		return true // сообщение обработано
	}

	// Проверяем, ожидаем ли голосовое сообщение
	if !state.WaitingForVoice {
		return false // сообщение не обработано
	}

	// Проверяем лимиты использования
	withinLimit, err := mh.stateManager.CheckLimit(userID)
	if err != nil {
		log.Printf("Ошибка проверки лимита: %v", err)
		msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Произошла ошибка при проверке лимита. Попробуйте позже.")
		bot.Send(msg)
		return true
	}
	if !withinLimit {
		msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Вы превысили дневной лимит использования. Для увеличения лимита перейдите на премиум тариф (/subscription).")
		bot.Send(msg)
		return true
	}

	// Обрабатываем голосовое сообщение
	if message.Voice != nil {
		mh.handleVoiceMessage(bot, message)
	}
	return true
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

	// Увеличиваем счетчик использований
	err = mh.stateManager.IncrementUsage(userID)
	if err != nil {
		log.Printf("Ошибка увеличения счетчика: %v", err)
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

// isValidEmail проверяет валидность email адреса
func (mh *MessageHandler) isValidEmail(email string) bool {
	// Простая, но достаточная регулярка для email
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// showSubscriptionPurchaseScreen показывает экран оформления подписки
func (mh *MessageHandler) showSubscriptionPurchaseScreen(bot *Bot, chatID int64, userID int64) {
	text := "💎 *Оформление Premium подписки*\n\n" +
		"✨ *Преимущества Premium:*\n" +
		"• 🚀 Неограниченное количество постов\n" +
		"• ⚡ Приоритетная обработка запросов\n" +
		"• 🎨 Расширенные настройки стилизации\n" +
		"• 📈 Детальная аналитика использования\n" +
		"• 🔧 Эксклюзивные функции и шаблоны\n" +
		"• 💬 Приоритетная техподдержка\n\n" +
		"💰 *Стоимость:* 990₽/месяц\n" +
		"📅 *Период:* 1 месяц\n" +
		"♻️ *Автопродление:* включено\n\n" +
		"📋 *Оферта:* [Пользовательское соглашение](#)\n\n" +
		"Нажмите «Подтвердить покупку» для перехода к оплате:"

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Подтвердить покупку", "confirm_purchase"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 Назад", "subscription"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = &keyboard
	bot.Send(msg)
}
