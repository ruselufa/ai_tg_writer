package bot

import (
	"ai_tg_writer/internal/domain"
	"ai_tg_writer/internal/infrastructure/database"
	"ai_tg_writer/internal/infrastructure/voice"
	"ai_tg_writer/internal/monitoring"
	"ai_tg_writer/internal/service"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// InlineHandler обрабатывает inline-команды
type InlineHandler struct {
	stateManager        *StateManager
	voiceHandler        *voice.VoiceHandler
	subscriptionService *service.SubscriptionService
	postHistoryRepo     *database.PostHistoryRepository
	prompts             map[string]Prompt
}

// NewInlineHandler создает новый обработчик inline-команд
func NewInlineHandler(stateManager *StateManager, voiceHandler *voice.VoiceHandler, subscriptionService *service.SubscriptionService, postHistoryRepo *database.PostHistoryRepository) *InlineHandler {
	// Загружаем промпты
	promptsFile, err := os.ReadFile("internal/infrastructure/prompts/prompts.json")
	if err != nil {
		log.Printf("Ошибка чтения файла промптов: %v", err)
		promptsFile = []byte("{}")
	}

	var prompts map[string]Prompt
	if err := json.Unmarshal(promptsFile, &prompts); err != nil {
		log.Printf("Ошибка разбора файла промптов: %v", err)
		prompts = make(map[string]Prompt)
	}

	return &InlineHandler{
		stateManager:        stateManager,
		voiceHandler:        voiceHandler,
		subscriptionService: subscriptionService,
		postHistoryRepo:     postHistoryRepo,
		prompts:             prompts,
	}
}

// HandleCallback обрабатывает callback от инлайн-кнопок
func (ih *InlineHandler) HandleCallback(bot *Bot, callback *tgbotapi.CallbackQuery) {
	monitoring.Debug("Callback от пользователя %d: %s", callback.From.ID, callback.Data)
	monitoring.RecordUserInteraction("callback", "unknown") // TODO: определить тариф пользователя

	// Отмечаем пользователя как активного при нажатии кнопки
	monitoring.MarkUserActiveGlobal(callback.From.ID)

	switch callback.Data {
	case "create_post":
		ih.handleCreatePost(bot, callback)
	case "create_telegram_post":
		ih.handleCreateScript(bot, callback, "telegram_post")
	case "create_script_youtube":
		ih.handleCreateScript(bot, callback, "youtube_script")
	case "create_script_reels":
		ih.handleCreateScript(bot, callback, "reels_script")
	case "create_post_instagram":
		ih.handleCreateScript(bot, callback, "instagram_post")
	case "continue_dictation":
		ih.handleContinueDictation(bot, callback)
	case "start_creation":
		ih.handleStartCreation(bot, callback)
	case "edit_start_creation":
		ih.handleEditStartCreation(bot, callback)
	case "approve":
		ih.handleApprove(bot, callback)
	case "edit_post":
		ih.handleEditPost(bot, callback)
	case "save_post":
		ih.handleSavePost(bot, callback)
	case "main_menu":
		ih.handleMainMenu(bot, callback)
	case "help":
		ih.handleHelp(bot, callback)
	case "profile":
		ih.handleProfile(bot, callback)
	case "post_history":
		ih.handlePostHistory(bot, callback, 1)
	case "payment_history":
		ih.handlePaymentHistory(bot, callback)
	case "post_history_1", "post_history_2", "post_history_3", "post_history_4", "post_history_5", "post_history_6", "post_history_7", "post_history_8", "post_history_9", "post_history_10":
		// Извлекаем номер страницы из callback data
		pageStr := callback.Data[len("post_history_"):]
		page, _ := strconv.Atoi(pageStr)
		ih.handlePostHistory(bot, callback, page)
	case "view_post_1", "view_post_2", "view_post_3", "view_post_4", "view_post_5", "view_post_6", "view_post_7", "view_post_8", "view_post_9", "view_post_10":
		// Извлекаем номер поста из callback data
		postStr := callback.Data[len("view_post_"):]
		postNumber, _ := strconv.Atoi(postStr)
		ih.handleViewPost(bot, callback, postNumber)
	case "subscription":
		ih.handleSubscription(bot, callback)
	case "buy_premium":
		ih.handleBuyPremium(bot, callback)
	case "confirm_purchase":
		ih.handleConfirmPurchase(bot, callback)
	case "cancel_subscription":
		ih.handleCancelSubscription(bot, callback)
	case "confirm_cancel_subscription":
		ih.handleConfirmCancelSubscription(bot, callback)
	case "retry_payment":
		ih.handleRetryPayment(bot, callback)
	case "change_payment_method":
		ih.handleChangePaymentMethod(bot, callback)
	case "styling_settings":
		ih.handleStylingSettings(bot, callback)
	case "test_formatting":
		ih.handleTestFormatting(bot, callback)
	case "rewrite_post_start":
		ih.handleRewritePostStart(bot, callback)
	case "rewrite_post_direct":
		ih.handleRewritePostDirect(bot, callback)
	case "rewrite_post_voice":
		ih.handleRewritePostVoice(bot, callback)
	case "no_action":
		// Игнорируем нажатие на пробел-заглушку
		return
	default:
		// Проверяем, не является ли это callback для страниц истории или просмотра постов
		if strings.HasPrefix(callback.Data, "post_history_") {
			pageStr := callback.Data[len("post_history_"):]
			if page, err := strconv.Atoi(pageStr); err == nil {
				ih.handlePostHistory(bot, callback, page)
				return
			}
		}
		if strings.HasPrefix(callback.Data, "view_post_") {
			postStr := callback.Data[len("view_post_"):]
			if postNumber, err := strconv.Atoi(postStr); err == nil {
				ih.handleViewPost(bot, callback, postNumber)
				return
			}
		}
		ih.handleUnknownCallback(bot, callback)
	}
}

// handleCreatePost обрабатывает выбор создания поста
func (ih *InlineHandler) handleCreatePost(bot *Bot, callback *tgbotapi.CallbackQuery) {
	userID := callback.From.ID

	// Проверяем подписку пользователя
	subscriptionStatus, canCreate, remainingFree, err := ih.checkUserSubscriptionStatus(userID)
	if err != nil {
		log.Printf("Ошибка проверки подписки: %v", err)
		// В случае ошибки разрешаем создание
		subscriptionStatus = "error"
		canCreate = true
	}

	if canCreate {
		// Обновляем состояние
		ih.stateManager.UpdateStep(userID, "selecting_content_type")
		ih.stateManager.SetContentType(userID, "telegram_post")
		ih.stateManager.ClearVoiceMessages(userID)
		ih.stateManager.SetCurrentPost(userID, nil)

		// Создаем клавиатуру с типами контента
		keyboard := bot.CreateContentTypeKeyboard()

		msg := tgbotapi.NewEditMessageText(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
			"✅ Выберите тип контента для создания:",
		)
		msg.ReplyMarkup = &keyboard

		bot.Send(msg)
	} else {
		// Показываем информацию о подписке и предлагаем оформить
		keyboard := ih.createSubscriptionKeyboard(userID, subscriptionStatus, remainingFree)

		var messageText string
		switch subscriptionStatus {
		case "cancelled":
			messageText = "❌ Ваша подписка была отменена.\n\n"
		case "expired":
			messageText = "⏰ Срок действия подписки истек.\n\n"
		case "no_subscription":
			messageText = "💎 У вас нет активной подписки.\n\n"
		default:
			messageText = "💎 Требуется подписка для создания контента.\n\n"
		}

		if remainingFree > 0 {
			messageText += fmt.Sprintf("🎁 У вас осталось %d бесплатных созданий в этом месяце.\n\n", remainingFree)
		} else {
			messageText += "🎁 Бесплатные создания на этот месяц закончились.\n\n"
		}

		messageText += "💳 Оформите подписку для неограниченного создания контента!"

		msg := tgbotapi.NewEditMessageText(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
			messageText,
		)
		msg.ReplyMarkup = &keyboard

		bot.Send(msg)
	}
}

// handleCreateScript обрабатывает выбор типа скрипта
func (ih *InlineHandler) handleCreateScript(bot *Bot, callback *tgbotapi.CallbackQuery, contentType string) {
	userID := callback.From.ID

	log.Printf("[DEBUG] handleCreateScript вызван для userID=%d, contentType=%s", userID, contentType)

	// Обновляем состояние
	ih.stateManager.UpdateStep(userID, "waiting_for_voice")
	ih.stateManager.SetContentType(userID, contentType)
	ih.stateManager.SetWaitingForVoice(userID, true)
	ih.stateManager.ClearVoiceMessages(userID)

	// Явно выставляем ожидание голосового
	state := ih.stateManager.GetState(userID)
	state.WaitingForVoice = true

	// Определяем название типа контента
	var contentName string
	switch contentType {
	case "youtube_script":
		contentName = "сценарий для видео на YouTube"
	case "reels_script":
		contentName = "сценарий для Reels в Instagram"
	case "instagram_post":
		contentName = "пост в Instagram"
	case "telegram_post":
		contentName = "пост в Telegram"
	default:
		contentName = "контент"
	}

	msg := tgbotapi.NewEditMessageText(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		"🎤 На какую тему вы хотите написать "+contentName+"?\n\n"+
			"Отправьте голосовое сообщение с вашими идеями:",
	)

	bot.Send(msg)
}

// handleStartCreation обрабатывает начало создания контента
func (ih *InlineHandler) handleStartCreation(bot *Bot, callback *tgbotapi.CallbackQuery) {
	userID := callback.From.ID
	state := ih.stateManager.GetState(userID)

	// Логируем текущее состояние PendingVoices
	log.Printf("[DEBUG] PendingVoices при старте создания: %+v", state.PendingVoices)

	// Проверяем, есть ли голосовые сообщения для обработки
	if state.PendingVoices == nil || len(state.PendingVoices) == 0 {
		msg := tgbotapi.NewEditMessageText(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
			"❌ Нет голосовых сообщений для обработки. Отправьте голосовое сообщение.",
		)
		bot.Send(msg)
		return
	}

	// Проверяем, что у всех сообщений есть путь к файлу
	hasInvalidFiles := false
	for _, voice := range state.PendingVoices {
		if voice.FilePath == "" {
			hasInvalidFiles = true
			break
		}
	}

	if hasInvalidFiles {
		msg := tgbotapi.NewEditMessageText(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
			"❌ Ошибка: некоторые голосовые сообщения не были корректно загружены. Попробуйте отправить их снова.",
		)
		bot.Send(msg)
		return
	}

	// Отправляем сообщение о начале обработки
	msg := tgbotapi.NewEditMessageText(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		"⏳ Начинаю обработку голосовых сообщений...",
	)
	bot.Send(msg)

	// Обрабатываем голосовые сообщения последовательно
	results := make([]string, 0)
	var firstHistoryID int

	// Обрабатываем каждое голосовое сообщение последовательно для правильной группировки
	voiceCount := 0
	var allVoiceTexts []string
	var totalDuration int
	var totalFileSize int

	for fileID, voice := range state.PendingVoices {
		voiceCount++

		// Транскрибируем файл
		isFirstMessage := voiceCount == 1
		text, historyID, err := ih.voiceHandler.TranscribeVoiceFile(voice.FilePath, userID, fileID, voice.Duration, voice.FileSize, isFirstMessage, firstHistoryID)
		if err != nil {
			log.Printf("Ошибка обработки голосового сообщения: %v", err)
			ih.stateManager.UpdateVoiceTranscription(userID, fileID, "", err)
			continue
		}

		// Сохраняем результат
		results = append(results, text)
		allVoiceTexts = append(allVoiceTexts, text)
		totalDuration += voice.Duration
		totalFileSize += voice.FileSize
		ih.stateManager.UpdateVoiceTranscription(userID, fileID, text, nil)

		// Первое сообщение создает запись в истории
		if voiceCount == 1 {
			firstHistoryID = historyID
			log.Printf("Установлен первый historyID для поста: %d", firstHistoryID)
		}

		// Удаляем временный файл
		if err := os.Remove(voice.FilePath); err != nil {
			log.Printf("Ошибка удаления временного файла %s: %v", voice.FilePath, err)
		}
	}

	// Обновляем запись истории с полной информацией
	if firstHistoryID > 0 && len(allVoiceTexts) > 0 {
		combinedText := strings.Join(allVoiceTexts, "\n\n")
		err := ih.voiceHandler.UpdateVoiceHistoryComplete(firstHistoryID, combinedText, totalDuration, totalFileSize)
		if err != nil {
			log.Printf("Ошибка обновления полной истории голосовых сообщений: %v", err)
		}
	}

	// Проверяем результаты
	if len(results) == 0 {
		msg := tgbotapi.NewMessage(
			callback.Message.Chat.ID,
			"❌ Не удалось обработать голосовые сообщения. Попробуйте еще раз.",
		)
		bot.Send(msg)
		return
	}

	// Формируем фрагменты идей
	var fragments []string
	for i, result := range results {
		fragments = append(fragments, fmt.Sprintf("Фрагмент идей %d: %s", i+1, result))
	}
	allMessages := strings.Join(fragments, "\n\n")

	// Определяем тип контента в зависимости от режима
	contentType := state.ContentType
	if contentType == "" {
		contentType = "telegram_post" // значение по умолчанию для обратной совместимости
	}

	// Если это режим рерайта с голосовыми указаниями, используем специальную логику
	var postText string
	var err error
	if state.RewriteMode == "voice" {
		// Получаем исходный текст поста
		originalText := ih.stateManager.GetRewritingPost(userID)
		if originalText == "" {
			msg := tgbotapi.NewMessage(
				callback.Message.Chat.ID,
				"❌ Ошибка: исходный текст поста не найден.",
			)
			bot.Send(msg)
			return
		}

		// Используем промпт для рерайта с указаниями
		postText, err = ih.voiceHandler.GenerateContent("rewrite_post", fmt.Sprintf("Исходный пост:\n%s\n\nУказания по рерайту:\n%s", originalText, allMessages), userID, firstHistoryID)
	} else {
		// Обычная генерация контента
		postText, err = ih.voiceHandler.GenerateContent(contentType, allMessages, userID, firstHistoryID)
	}
	if err != nil {
		log.Printf("Ошибка генерации поста: %v", err)
		msg := tgbotapi.NewMessage(
			callback.Message.Chat.ID,
			"❌ Не удалось сгенерировать пост. Попробуйте еще раз.",
		)
		bot.Send(msg)
		return
	}

	// Сохраняем сгенерированный текст
	ih.stateManager.SetLastGeneratedText(userID, postText)

	// Форматируем пост с entities
	formatter := NewTelegramPostFormatter(state.PostStyling)
	cleanText, entities := formatter.FormatPost(postText)

	// Создаем новый пост
	post := Post{
		ContentType: state.ContentType,
		Content:     cleanText,
		Messages:    results,
		Entities:    entities,
		Styling:     state.PostStyling,
		HistoryID:   firstHistoryID,
	}

	// Если это режим рерайта, добавляем исходный текст в сообщения
	if state.RewriteMode == "voice" {
		originalText := ih.stateManager.GetRewritingPost(userID)
		if originalText != "" {
			post.Messages = append([]string{originalText}, post.Messages...)
		}
	}

	// Сохраняем пост
	ih.stateManager.SetCurrentPost(userID, &post)
	ih.stateManager.SetApprovalStatus(userID, "pending")

	// Очищаем состояние рерайта после успешного создания поста
	if state.RewriteMode != "" {
		ih.stateManager.ClearRewriteState(userID)
	}

	// Отправляем результат с кнопками согласования
	keyboard := bot.CreateApprovalKeyboard()
	messageID, err := bot.SendFormattedMessageWithKeyboard(
		callback.Message.Chat.ID,
		cleanText,
		entities,
		keyboard,
	)
	if err != nil {
		log.Printf("Ошибка отправки форматированного сообщения: %v", err)
		// Отправляем без форматирования в случае ошибки
		resultMsg := tgbotapi.NewMessage(callback.Message.Chat.ID, cleanText)
		resultMsg.ReplyMarkup = keyboard
		bot.Send(resultMsg)
	} else {
		// Сохраняем ID сообщения с готовым постом
		ih.stateManager.SetPostMessageID(userID, messageID)
		log.Printf("Сохранили ID сообщения с постом: %d", messageID)
	}
}

// handleAddMore обрабатывает добавление еще голосовых сообщений
func (ih *InlineHandler) handleAddMore(bot *Bot, callback *tgbotapi.CallbackQuery) {
	userID := callback.From.ID

	// Обновляем состояние
	ih.stateManager.UpdateStep(userID, "waiting_for_voice")
	ih.stateManager.SetWaitingForVoice(userID, true)

	msg := tgbotapi.NewEditMessageText(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		"🎤 Отправьте следующее голосовое сообщение:",
	)

	bot.Send(msg)
}

// handleEditPost обрабатывает редактирование поста
func (ih *InlineHandler) handleEditPost(bot *Bot, callback *tgbotapi.CallbackQuery) {
	userID := callback.From.ID
	state := ih.stateManager.GetState(userID)

	// Проверяем, есть ли текущий пост для редактирования
	if state.CurrentPost == nil {
		msg := tgbotapi.NewEditMessageText(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
			"❌ Нет поста для редактирования.",
		)
		bot.Send(msg)
		return
	}

	// Очищаем старые правки и устанавливаем состояние редактирования
	ih.stateManager.ClearEditMessages(userID)
	ih.stateManager.ClearPendingEdits(userID)
	ih.stateManager.UpdateStep(userID, "editing")
	ih.stateManager.SetWaitingForVoice(userID, true)
	ih.stateManager.SetApprovalStatus(userID, "editing")

	// Отправляем сообщение с инструкциями по редактированию
	msg := tgbotapi.NewEditMessageText(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		"🎤 Отправьте голосовые сообщения с изменениями для поста:\n\n"+
			"Текущий текст:\n"+state.CurrentPost.Content,
	)

	bot.Send(msg)
}

// handleMainMenu обрабатывает возврат в главное меню
func (ih *InlineHandler) handleMainMenu(bot *Bot, callback *tgbotapi.CallbackQuery) {
	userID := callback.From.ID

	// Сохраняем текущий пост, если есть
	state := ih.stateManager.GetState(userID)
	if state.CurrentPost != nil {
		ih.stateManager.SavePost(userID, *state.CurrentPost)
		log.Printf("Пост сохранен в БД (заглушка) при выходе в меню: %s", state.CurrentPost.ContentType)
	}

	// Получаем информацию о подписке
	sub, _ := bot.SubscriptionService.GetUserSubscription(userID)
	// Получаем общее использование (лимит 5 запросов навсегда)
	used, _ := bot.DB.GetUserUsageTotal(userID)
	const freeLimit = 5 // Общий лимит запросов навсегда

	var subLabel string
	if sub != nil && sub.Active {
		subLabel = "💎 Подписка: Premium"
	} else {
		remaining := freeLimit - used
		if remaining < 0 {
			remaining = 0
		}
		subLabel = fmt.Sprintf("💎 Подписка (%d/%d)", remaining, freeLimit)
	}

	// Полностью очищаем состояние
	ih.stateManager.UpdateStep(userID, "idle")
	ih.stateManager.SetCurrentPost(userID, nil)
	ih.stateManager.ClearVoiceMessages(userID)
	ih.stateManager.ClearPendingVoices(userID)
	ih.stateManager.ClearEditMessages(userID)
	ih.stateManager.ClearPendingEdits(userID)
	ih.stateManager.SetApprovalStatus(userID, "idle")
	ih.stateManager.SetWaitingForVoice(userID, false)

	// Формируем главное меню с динамичной подписью подписки
	text := "Привет! Ты попал в бота, который умеет волшебство:\n\n" +
		"Запиши аудио → получи готовый пост для соцсетей за пару секунд.\n\n" +
		"Что получаешь:\n\n" +
		"— Экономию часов на написании текстов\n" +
		"— Готовый стильный пост без ошибок\n" +
		"— Контент, который можно сразу публиковать\n\n" +
		"Хочешь пост за 1 минуту? Отправь голосовое прямо сейчас"

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📝 Создать пост/сценарий", "create_post")),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 Сделать рерайт поста", "rewrite_post_start")),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👤 Мой профиль", "profile"),
			tgbotapi.NewInlineKeyboardButtonData(subLabel, "subscription")),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❓ Помощь", "help")),
	)

	msg := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, text)
	msg.ReplyMarkup = &keyboard
	bot.Send(msg)
}

// handleContinueDictation обрабатывает продолжение диктовки
func (ih *InlineHandler) handleContinueDictation(bot *Bot, callback *tgbotapi.CallbackQuery) {
	userID := callback.From.ID
	state := ih.stateManager.GetState(userID)

	// Устанавливаем состояние ожидания голосового сообщения
	state.WaitingForVoice = true
	ih.stateManager.SetWaitingForVoice(userID, true)

	// Определяем текст сообщения в зависимости от режима
	var messageText string
	if state.ApprovalStatus == "editing" {
		messageText = "🎤 Отправьте ещё голосовые сообщения с правками. Когда закончите, нажмите кнопку \"Применить изменения\"."
	} else {
		messageText = "🎤 Отправьте ещё голосовые сообщения. Когда закончите, нажмите кнопку \"Начать создание\"."
	}

	// Отправляем сообщение
	msg := tgbotapi.NewEditMessageText(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		messageText,
	)

	var keyboard tgbotapi.InlineKeyboardMarkup
	if state.ApprovalStatus == "editing" {
		keyboard = bot.CreateEditContinueKeyboard()
	} else {
		keyboard = bot.CreateContinueKeyboard()
	}

	msg.ReplyMarkup = &keyboard
	bot.Send(msg)
}

// handleSavePost обрабатывает сохранение поста
func (ih *InlineHandler) handleSavePost(bot *Bot, callback *tgbotapi.CallbackQuery) {
	userID := callback.From.ID
	state := ih.stateManager.GetState(userID)

	if state.CurrentPost == nil {
		msg := tgbotapi.NewEditMessageText(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
			"❌ Ошибка: нет текущего поста для сохранения",
		)
		bot.Send(msg)
		return
	}

	// Безопасно увеличиваем счетчик использований с проверкой подписки
	ih.incrementUsageIfNeeded(userID)

	// Сохраняем пост в БД (заглушка)
	ih.stateManager.SavePost(userID, *state.CurrentPost)
	log.Printf("Пост сохранен в БД (заглушка): %s", state.CurrentPost.ContentType)

	// Отмечаем в истории как сохраненный
	if state.CurrentPost.HistoryID > 0 {
		log.Printf("Отмечаем пост как сохраненный в истории ID: %d", state.CurrentPost.HistoryID)
		err := ih.voiceHandler.MarkPostAsSaved(state.CurrentPost.HistoryID)
		if err != nil {
			log.Printf("Ошибка отметки поста как сохраненного: %v", err)
		}
	}

	// Сохраняем данные поста перед очисткой состояния
	postContent := state.CurrentPost.Content
	postEntities := state.CurrentPost.Entities

	// Очищаем состояние
	ih.stateManager.UpdateStep(userID, "idle")
	ih.stateManager.SetCurrentPost(userID, nil)
	ih.stateManager.ClearVoiceMessages(userID)
	ih.stateManager.ClearPendingVoices(userID)
	ih.stateManager.ClearEditMessages(userID)
	ih.stateManager.ClearPendingEdits(userID)
	ih.stateManager.SetApprovalStatus(userID, "approved")

	// СНАЧАЛА удаляем старое сообщение с кнопками
	deleteMsg := tgbotapi.NewDeleteMessage(callback.Message.Chat.ID, callback.Message.MessageID)
	bot.Send(deleteMsg)

	// ЗАТЕМ отправляем готовый пост БЕЗ кнопок управления
	_, err := bot.SendFormattedMessage(
		callback.Message.Chat.ID,
		postContent,
		postEntities,
	)
	if err != nil {
		log.Printf("Ошибка отправки готового поста: %v", err)
	}

	// И НАКОНЕЦ отправляем НОВОЕ сообщение с главным меню
	keyboard := bot.CreateMainKeyboard()
	newMsg := tgbotapi.NewMessage(
		callback.Message.Chat.ID,
		"✅ Пост успешно сохранен! Текст остался в чате.\n\nПривет! Я помогу тебе создать мощный контент из твоих идей. Выбери, что хочешь создать:",
	)
	newMsg.ReplyMarkup = keyboard
	bot.Send(newMsg)
}

// handleApprove обрабатывает согласие с результатом
func (ih *InlineHandler) handleApprove(bot *Bot, callback *tgbotapi.CallbackQuery) {
	userID := callback.From.ID

	// Безопасно увеличиваем счетчик использований с проверкой подписки
	ih.incrementUsageIfNeeded(userID)

	// Сохраняем пост в БД (заглушка)
	state := ih.stateManager.GetState(userID)
	var postContent string
	var postEntities []MessageEntity

	if state.CurrentPost != nil {
		ih.stateManager.SavePost(userID, *state.CurrentPost)
		log.Printf("Пост сохранен в БД (заглушка): %s", state.CurrentPost.ContentType)

		// Отмечаем в истории как сохраненный
		if state.CurrentPost.HistoryID > 0 {
			log.Printf("Отмечаем пост как сохраненный в истории ID: %d", state.CurrentPost.HistoryID)
			err := ih.voiceHandler.MarkPostAsSaved(state.CurrentPost.HistoryID)
			if err != nil {
				log.Printf("Ошибка отметки поста как сохраненного: %v", err)
			}
		}

		// Сохраняем данные поста перед очисткой состояния
		postContent = state.CurrentPost.Content
		postEntities = state.CurrentPost.Entities
	}

	// Очищаем состояние
	ih.stateManager.UpdateStep(userID, "idle")
	ih.stateManager.SetCurrentPost(userID, nil)
	ih.stateManager.ClearVoiceMessages(userID)
	ih.stateManager.ClearPendingVoices(userID)
	ih.stateManager.SetApprovalStatus(userID, "approved")

	// СНАЧАЛА удаляем старое сообщение с кнопками
	deleteMsg := tgbotapi.NewDeleteMessage(callback.Message.Chat.ID, callback.Message.MessageID)
	bot.Send(deleteMsg)

	// ЗАТЕМ отправляем готовый пост БЕЗ кнопок управления
	if postContent != "" {
		_, err := bot.SendFormattedMessage(
			callback.Message.Chat.ID,
			postContent,
			postEntities,
		)
		if err != nil {
			log.Printf("Ошибка отправки готового поста: %v", err)
		}
	}

	// И НАКОНЕЦ отправляем НОВОЕ сообщение с главным меню
	text := "✅ Пост сохранен! Текст остался в чате.\n\nПривет! Я помогу тебе создать мощный контент из твоих идей. Выбери, что хочешь создать:"
	keyboard := bot.CreateMainKeyboard()

	newMsg := tgbotapi.NewMessage(
		callback.Message.Chat.ID,
		text,
	)
	newMsg.ReplyMarkup = keyboard
	bot.Send(newMsg)
}

// handleEditStartCreation обрабатывает начало создания правок
func (ih *InlineHandler) handleEditStartCreation(bot *Bot, callback *tgbotapi.CallbackQuery) {
	userID := callback.From.ID
	state := ih.stateManager.GetState(userID)

	// Проверяем, есть ли голосовые сообщения для правок
	if state.PendingEdits == nil || len(state.PendingEdits) == 0 {
		msg := tgbotapi.NewEditMessageText(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
			"❌ Нет голосовых сообщений с правками для обработки. Отправьте голосовое сообщение с изменениями.",
		)
		bot.Send(msg)
		return
	}

	// Проверяем, что у всех сообщений есть путь к файлу
	hasInvalidFiles := false
	for _, voice := range state.PendingEdits {
		if voice.FilePath == "" {
			hasInvalidFiles = true
			break
		}
	}

	if hasInvalidFiles {
		msg := tgbotapi.NewEditMessageText(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
			"❌ Ошибка: некоторые голосовые сообщения с правками не были корректно загружены. Попробуйте отправить их снова.",
		)
		bot.Send(msg)
		return
	}

	// Отправляем сообщение о начале обработки
	msg := tgbotapi.NewEditMessageText(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		"⏳ Начинаю обработку правок...",
	)
	bot.Send(msg)

	// Обрабатываем голосовые сообщения с правками последовательно
	results := make([]string, 0)
	var firstHistoryID int

	// Обрабатываем каждое голосовое сообщение с правками последовательно
	editCount := 0
	var allEditTexts []string
	var totalEditDuration int
	var totalEditFileSize int

	for fileID, voice := range state.PendingEdits {
		editCount++

		// Транскрибируем файл
		isFirstMessage := editCount == 1
		log.Printf("Обрабатываем правку %d: duration=%d, fileSize=%d, isFirstMessage=%v", editCount, voice.Duration, voice.FileSize, isFirstMessage)
		text, historyID, err := ih.voiceHandler.TranscribeVoiceFile(voice.FilePath, userID, fileID, voice.Duration, voice.FileSize, isFirstMessage, firstHistoryID)
		if err != nil {
			log.Printf("Ошибка обработки голосового сообщения с правками: %v", err)
			continue
		}

		// Сохраняем результат
		results = append(results, text)
		allEditTexts = append(allEditTexts, text)
		totalEditDuration += voice.Duration
		totalEditFileSize += voice.FileSize

		// Сохраняем первый historyID для поста
		if firstHistoryID == 0 && historyID > 0 {
			firstHistoryID = historyID
			log.Printf("Установлен первый historyID для поста (правки): %d", firstHistoryID)
		}

		// Удаляем временный файл
		if err := os.Remove(voice.FilePath); err != nil {
			log.Printf("Ошибка удаления временного файла %s: %v", voice.FilePath, err)
		}
	}

	// Обновляем запись истории с полной информацией о правках
	if firstHistoryID > 0 && len(allEditTexts) > 0 {
		combinedEditText := strings.Join(allEditTexts, "\n\n")
		err := ih.voiceHandler.UpdateVoiceHistoryComplete(firstHistoryID, combinedEditText, totalEditDuration, totalEditFileSize)
		if err != nil {
			log.Printf("Ошибка обновления полной истории правок: %v", err)
		}
	}

	// Проверяем результаты
	if len(results) == 0 {
		msg := tgbotapi.NewMessage(
			callback.Message.Chat.ID,
			"❌ Не удалось обработать голосовые сообщения с правками. Попробуйте еще раз.",
		)
		bot.Send(msg)
		return
	}

	// Формируем текст правок
	editText := strings.Join(results, "\n\n")

	// Получаем исходный текст
	originalText := state.LastGeneratedText
	if originalText == "" {
		originalText = "Исходный текст недоступен"
	}

	// Формируем запрос для ИИ с исходным текстом и правками
	prompt := fmt.Sprintf("Исходный текст:\n%s\n\nПравки пользователя:\n%s\n\nПожалуйста, внесите изменения в исходный текст согласно правкам пользователя.", originalText, editText)

	// Генерируем обновленный контент через VoiceHandler
	contentType := state.ContentType
	if contentType == "" {
		contentType = "telegram_post" // значение по умолчанию для обратной совместимости
	}
	updatedText, err := ih.voiceHandler.GenerateContent(contentType, prompt, userID, firstHistoryID)
	if err != nil {
		log.Printf("Ошибка генерации обновленного поста: %v", err)
		msg := tgbotapi.NewMessage(
			callback.Message.Chat.ID,
			"❌ Не удалось сгенерировать обновленный пост. Попробуйте еще раз.",
		)
		bot.Send(msg)
		return
	}

	// Форматируем обновленный пост с entities
	formatter := NewTelegramPostFormatter(state.PostStyling)
	cleanText, entities := formatter.FormatPost(updatedText)

	// Обновляем пост
	state.CurrentPost.Content = cleanText
	state.CurrentPost.Entities = entities
	state.CurrentPost.Messages = append(state.CurrentPost.Messages, results...)
	// Обновляем HistoryID на новую запись с правками
	state.CurrentPost.HistoryID = firstHistoryID
	// Сохраняем обновленный пост в состоянии
	ih.stateManager.SetCurrentPost(userID, state.CurrentPost)
	ih.stateManager.SetLastGeneratedText(userID, updatedText)
	ih.stateManager.SetApprovalStatus(userID, "pending")

	// Отправляем обновленный результат с кнопками согласования
	keyboard := bot.CreateEditApprovalKeyboard()
	messageID, err := bot.SendFormattedMessageWithKeyboard(
		callback.Message.Chat.ID,
		cleanText,
		entities,
		keyboard,
	)
	if err != nil {
		log.Printf("Ошибка отправки форматированного сообщения: %v", err)
		// Отправляем без форматирования в случае ошибки
		resultMsg := tgbotapi.NewMessage(callback.Message.Chat.ID, cleanText)
		resultMsg.ReplyMarkup = keyboard
		bot.Send(resultMsg)
	} else {
		// Сохраняем ID сообщения с обновленным постом
		ih.stateManager.SetPostMessageID(userID, messageID)
		log.Printf("Сохранили ID сообщения с обновленным постом: %d", messageID)
	}
}

// processVoiceMessages обрабатывает все голосовые сообщения пользователя
func (ih *InlineHandler) processVoiceMessages(bot *Bot, callback *tgbotapi.CallbackQuery, state *UserState) {
	userID := callback.From.ID

	// Объединяем все голосовые сообщения
	allMessages := strings.Join(state.VoiceMessages, "\n\n")

	// Получаем промпты для текущего типа контента
	log.Printf("Загруженные промпты: %v", ih.prompts)
	contentPrompts, ok := ih.prompts[state.ContentType]
	if !ok {
		log.Printf("Ошибка: не найдены промпты для типа контента '%s'", state.ContentType)
		msg := tgbotapi.NewEditMessageText(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
			"❌ Ошибка: неизвестный тип контента",
		)
		bot.Send(msg)
		return
	}

	// Проверяем наличие всех необходимых промптов
	log.Printf("Доступные промпты для %s: %v", state.ContentType, contentPrompts)

	var resultText string
	// TODO: Интеграция с OpenAI API
	if state.CurrentPost != nil {
		// Если это редактирование, используем промпты для редактирования
		// TODO: Использовать contentPrompts["edit"]["system"] как system prompt
		// и contentPrompts["edit"]["user"] как user prompt, заменив:
		// {current_text} на state.CurrentPost.Content
		// {new_text} на allMessages
		resultText = "🤖 Обновленный текст на основе существующего контента и новых идей:\n\n" +
			"[Тип контента: " + state.ContentType + ", режим: редактирование]\n\n" +
			"Существующий текст:\n" + state.CurrentPost.Content + "\n\n" +
			"Новые идеи:\n" + allMessages
	} else {
		// Если это новый пост, используем обычные промпты
		// TODO: Использовать contentPrompts["system"] как system prompt
		// и contentPrompts["user"] как user prompt, заменив:
		// {text} на allMessages
		resultText = "🤖 Сгенерированный текст на основе ваших идей:\n\n" +
			"[Тип контента: " + state.ContentType + ", режим: создание]\n\n" +
			allMessages
	}

	// Создаем новый пост
	post := Post{
		ContentType: state.ContentType,
		Content:     resultText,
		Messages:    state.VoiceMessages,
	}

	// Если это редактирование, добавляем предыдущие сообщения
	if state.CurrentPost != nil {
		post.Messages = append(state.CurrentPost.Messages, state.VoiceMessages...)
	}

	// Сохраняем пост
	ih.stateManager.SetCurrentPost(userID, &post)

	// Отправляем результат с кнопками действий
	keyboard := bot.CreatePostActionKeyboard()
	msg := tgbotapi.NewEditMessageText(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		resultText,
	)
	msg.ReplyMarkup = &keyboard

	bot.Send(msg)

	// Очищаем голосовые сообщения
	ih.stateManager.ClearVoiceMessages(userID)
	ih.stateManager.UpdateStep(userID, "idle")
}

// handleHelp обрабатывает кнопку помощи
func (ih *InlineHandler) handleHelp(bot *Bot, callback *tgbotapi.CallbackQuery) {
	text := `📚 Справка по использованию бота

🎯 Основные возможности:
• Создание постов для Telegram
• Сценарии для YouTube видео
• Сценарии для Instagram Reels
• Посты для Instagram

🎤 Как использовать:
1. Выберите тип контента
2. Отправьте голосовое сообщение с идеями
3. При необходимости добавьте еще сообщения
4. Получите готовый контент

💡 Советы:
• Говорите четко и структурированно
• Можно отправлять несколько голосовых подряд
• Бот автоматически объединит все сообщения`

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("💬 Написать в поддержку", "https://t.me/socialflow_support_bot"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 Назад в меню", "main_menu"),
		),
	)

	msg := tgbotapi.NewEditMessageText(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
	)
	msg.ReplyMarkup = &keyboard

	bot.Send(msg)
}

// handleProfile обрабатывает кнопку профиля
func (ih *InlineHandler) handleProfile(bot *Bot, callback *tgbotapi.CallbackQuery) {
	userID := callback.From.ID

	// Получаем информацию о подписке пользователя
	sub, _ := bot.SubscriptionService.GetUserSubscription(userID)
	available := bot.SubscriptionService.GetAvailableTariffs()
	var premium domain.Tariff
	if len(available) > 0 {
		premium = available[0]
	}

	var messageText string
	var keyboard tgbotapi.InlineKeyboardMarkup

	if sub == nil || !sub.Active {
		// Нет подписки
		messageText = fmt.Sprintf(`👤 Ваш профиль

🆔 ID: %d
📊 Тариф: Бесплатный
⏰ Срок действия: бессрочно

💎 *Премиум-тариф* – %s

— Создание неограниченного количества постов.
— Приоритетная поддержка
— Быстрая скорость генерации постов
— Рерайтинг посто по ссылке в Телеграм

💰 Стоимость: %.0f₽/месяц`, userID, premium.Description, premium.Price)

		keyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("📚 История постов", "post_history"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("💰 История оплат", "payment_history"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("💳 Приобрести подписку", "buy_premium"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🔙 Назад в меню", "main_menu"),
			),
		)
	} else {
		// Есть активная подписка
		nextPay := sub.NextPayment.Format("02.01.2006")
		messageText = fmt.Sprintf(`👤 Ваш профиль

🆔 ID: %d
💎 Подписка: Premium
📅 Следующий платеж: %s
✅ Статус: активна`, userID, nextPay)

		keyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("📚 История постов", "post_history"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("💰 История оплат", "payment_history"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("❌ Отменить подписку и отвязать карту", "cancel_subscription"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🔙 Назад в меню", "main_menu"),
			),
		)
	}

	msg := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, messageText)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = &keyboard
	bot.Send(msg)
}

// handleSubscription обрабатывает кнопку подписки
func (ih *InlineHandler) handleSubscription(bot *Bot, callback *tgbotapi.CallbackQuery) {
	userID := callback.From.ID

	// Информация о подписке
	sub, _ := bot.SubscriptionService.GetUserSubscription(userID)
	// Лимит бесплатных запросов
	used, _ := bot.DB.GetUserUsageToday(userID)
	const freeLimit = 5

	var text string
	var keyboard tgbotapi.InlineKeyboardMarkup

	if sub == nil || !sub.Active {
		remaining := freeLimit - used
		if remaining < 0 {
			remaining = 0
		}
		text = fmt.Sprintf(`💎 Подписка

📊 Текущий тариф: *Бесплатный*
⏰ Срок действия: бессрочно
📈 Осталось бесплатных постов в этом месяце: *%d/%d*

Подключи Premium тариф и получи:

— Создание неограниченного количества постов.
— Приоритетная поддержка
— Быстрая скорость генерации постов
— Рерайтинг посто по ссылке в Телеграм

💳 Стоимость: 990₽/месяц`, remaining, freeLimit)

		keyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("💰 Купить подписку", "buy_premium"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🔙 Назад в меню", "main_menu"),
			),
		)
	} else {
		var subStatus string
		if sub.Status == string(domain.SubscriptionStatusCancelled) {
			subStatus = "Подписка активна до"
		} else {
			subStatus = "Следующий платеж"
		}
		// надо поставить московское время
		nextPay := sub.NextPayment.In(time.FixedZone("UTC+3", 3*60*60)).Format("02.01.2006 15:04 МСК")
		text = fmt.Sprintf(`💎 Подписка

📊 Текущий тариф: *Premium*
📅 %s: %s
✅ Статус: активна`, subStatus, nextPay)

		var rows [][]tgbotapi.InlineKeyboardButton
		if sub.Status == string(domain.SubscriptionStatusActive) && sub.YKPaymentMethodID != nil && sub.YKLastPaymentID != nil {
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("❌ Отменить подписку и отвязать карту", "cancel_subscription"),
			))
		}

		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 Назад в меню", "main_menu"),
		))

		keyboard = tgbotapi.NewInlineKeyboardMarkup(rows...)
	}

	msg := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = &keyboard
	bot.Send(msg)
}

// handleStylingSettings обрабатывает настройки стилизации
func (ih *InlineHandler) handleStylingSettings(bot *Bot, callback *tgbotapi.CallbackQuery) {
	userID := callback.From.ID
	state := ih.stateManager.GetState(userID)
	styling := state.PostStyling

	text := `🎨 Настройки стилизации постов

Текущие настройки:
• Жирный текст: ` + ih.formatBool(styling.UseBold) + `
• Курсив: ` + ih.formatBool(styling.UseItalic) + `
• Зачеркивание: ` + ih.formatBool(styling.UseStrikethrough) + `
• Код: ` + ih.formatBool(styling.UseCode) + `
• Ссылки: ` + ih.formatBool(styling.UseLinks) + `
• Хештеги: ` + ih.formatBool(styling.UseHashtags) + `
• Упоминания: ` + ih.formatBool(styling.UseMentions) + `
• Подчеркивание: ` + ih.formatBool(styling.UseUnderline) + `
• Блоки кода: ` + ih.formatBool(styling.UsePre) + `

Выберите, что хотите изменить:`

	keyboard := bot.CreateStylingSettingsKeyboard()
	msg := tgbotapi.NewEditMessageText(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
	)
	msg.ReplyMarkup = &keyboard
	bot.Send(msg)
}

// formatBool форматирует булево значение для отображения
func (ih *InlineHandler) formatBool(value bool) string {
	if value {
		return "✅ Включено"
	}
	return "❌ Отключено"
}

// handleTestFormatting обрабатывает тест форматирования
func (ih *InlineHandler) handleTestFormatting(bot *Bot, callback *tgbotapi.CallbackQuery) {
	// Тестовый текст с разными типами разметки
	testText := `*🔥 Тест форматирования Telegram* 🔥

Этот текст демонстрирует различные возможности форматирования в Telegram:

*Жирный текст* - для заголовков и важных моментов
_Курсив_ - для акцентов и выделения
~Зачеркнутый текст~ - для исправлений
` + "`" + `код` + "`" + ` - для технических терминов

🔹 *Списки с разметкой:*
✔️ _Пункт 1_ - с курсивом
✔️ *Пункт 2* - с жирным
✔️ ` + "`" + `Пункт 3` + "`" + ` - с кодом

🔗 *Ссылки:*
[Telegram API](https://core.telegram.org/api/entities)

#Тест #Форматирование #Telegram`

	// Создаем форматтер с настройками по умолчанию
	styling := DefaultPostStyling()
	formatter := NewTelegramPostFormatter(styling)

	// Парсим Markdown в entities напрямую (без FormatPost)
	cleanText, entities := formatter.ParseMarkdownToEntities(testText)

	// Отправляем с форматированием
	_, err := bot.SendFormattedMessage(callback.Message.Chat.ID, cleanText, entities)
	if err != nil {
		log.Printf("Ошибка отправки тестового сообщения: %v", err)
		// Отправляем без форматирования в случае ошибки
		msg := tgbotapi.NewMessage(callback.Message.Chat.ID, testText)
		bot.Send(msg)
	}

	// Отправляем сообщение об успехе
	msg := tgbotapi.NewEditMessageText(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		"✅ Тестовое сообщение с форматированием отправлено! Проверьте чат выше.",
	)
	keyboard := bot.CreateMainKeyboard()
	msg.ReplyMarkup = &keyboard
	bot.Send(msg)
}

// handleBuyPremium обрабатывает покупку премиум подписки
func (ih *InlineHandler) handleBuyPremium(bot *Bot, callback *tgbotapi.CallbackQuery) {
	userID := callback.From.ID

	user, _ := bot.DB.GetOrCreateUser(userID, callback.From.UserName, callback.From.FirstName, callback.From.LastName)
	if user.Email == "" {
		// помечаем ожидание email
		ih.stateManager.GetState(userID).WaitingForEmail = true
		msg := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID,
			"📧 *Введите ваш e-mail*\n\n"+
				"Для получения кассового чека нужен e-mail адрес.\n"+
				"Пример: user@example.com\n\n"+
				"💡 Для отмены используйте /start")
		msg.ParseMode = "Markdown"
		bot.Send(msg)
		return
	}

	// Показываем экран оформления подписки с преимуществами
	// Рассчитываем дату окончания подписки (текущая дата + 1 месяц)
	subscriptionEndDate := time.Now().AddDate(0, 1, 0)
	formattedDate := subscriptionEndDate.Format("02.01.2006")

	text := "💎 *Оформление Premium подписки*\n\n" +
		"✨ *Преимущества Premium:*\n" +
		"• 🚀 Неограниченное количество постов\n" +
		"• ⚡ Приоритетная обработка запросов\n" +
		"• 🎨 Расширенные настройки стилизации\n" +
		"• 📈 Детальная аналитика использования\n" +
		"• 🔧 Эксклюзивные функции и шаблоны\n" +
		"• 💬 Приоритетная техподдержка\n\n" +
		"💰 *Стоимость:* 990₽/месяц\n" +
		"📅 *Период:* 1 месяц (до " + formattedDate + ")\n" +
		"♻️ *Автопродление:* включено\n\n" +
		// "📋 *Оферта:* [Пользовательское соглашение](#)\n\n" +
		"Нажмите «Подтвердить покупку» для перехода к оплате:"

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Подтвердить покупку", "confirm_purchase"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 Назад", "subscription"),
		),
	)

	msg := tgbotapi.NewEditMessageText(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
	)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = &keyboard
	bot.Send(msg)
}

// handleConfirmPurchase обрабатывает подтверждение покупки
func (ih *InlineHandler) handleConfirmPurchase(bot *Bot, callback *tgbotapi.CallbackQuery) {
	userID := callback.From.ID

	// Создаем ссылку на оплату подписки
	paymentURL, err := bot.CreateSubscriptionLink(userID, "premium", 1.0)
	if err != nil {
		msg := tgbotapi.NewEditMessageText(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
			"❌ Ошибка создания ссылки на оплату. Попробуйте позже.",
		)
		bot.Send(msg)
		return
	}

	// Создаем кнопку для перехода к оплате
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("💳 Перейти к оплате", paymentURL),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 Назад", "buy_premium"),
		),
	)

	msg := tgbotapi.NewEditMessageText(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		"💳 *Переход к оплате*\n\n"+
			"Нажмите кнопку ниже для перехода к оплате.\n"+
			"После успешной оплаты ваша подписка будет активирована автоматически.",
	)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = &keyboard
	bot.Send(msg)
}

// handleCancelSubscription обрабатывает отмену подписки
func (ih *InlineHandler) handleCancelSubscription(bot *Bot, callback *tgbotapi.CallbackQuery) {
	// Создаем кнопки подтверждения
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Да, отменить", "confirm_cancel_subscription"),
			tgbotapi.NewInlineKeyboardButtonData("❌ Нет, оставить", "subscription"),
		),
	)

	msg := tgbotapi.NewEditMessageText(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		// "⚠️ *Подтверждение отмены подписки*\n\n"+
		// 	"Вы уверены, что хотите отменить подписку и отвязать карту?\n\n"+
		// 	"ℹ️ *Важно:* Ваша подписка будет работать до конца оплаченного периода.\n"+
		// 	"После этого вы потеряете доступ к премиум функциям.\n\n"+
		// 	"💡 Вы можете возобновить подписку в любое время.",
		"Жаль, что ты решил уйти. Мы старались сделать удобный сервис, который снимет рутину и поможет в создании контента.\n\n"+
			"Твоя подписка будет активна до конца оплаченного периода, после чего доступ к премиум функциям будет ограничен.\n\n"+
			"Но ты всегда можно возобновить подписку.",
	)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = &keyboard

	bot.Send(msg)
}

// handleConfirmCancelSubscription подтверждает отмену подписки
func (ih *InlineHandler) handleConfirmCancelSubscription(bot *Bot, callback *tgbotapi.CallbackQuery) {
	userID := callback.From.ID

	// Отменяем подписку
	err := bot.SubscriptionService.CancelSubscription(userID)
	if err != nil {
		msg := tgbotapi.NewEditMessageText(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
			"❌ Ошибка при отмене подписки. Попробуйте позже.",
		)
		bot.Send(msg)
		return
	}

	// Создаем сообщение об успешной отмене с кнопкой возврата в меню
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 Назад в меню", "main_menu"),
		),
	)

	msg := tgbotapi.NewEditMessageText(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		"✅ *Подписка отменена*\n\n"+
			"Ты ушёл, сказав: «Прощай»,\n"+
			"Но скучал, и — невзначай,\n"+
			"Возвратился вновь к нам ты —\n"+
			"Снова вместе, как мечты\n\n"+
			"Бот не в обиде, бот прощает,\n"+
			"Тепло встречать он обожает,\n"+
			"Добро пожаловать назад,\n"+
			"Подписке — быть, и это факт!",
	)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = &keyboard

	bot.Send(msg)
}

// handleRetryPayment обрабатывает кнопку повторной оплаты
func (ih *InlineHandler) handleRetryPayment(bot *Bot, callback *tgbotapi.CallbackQuery) {
	userID := callback.From.ID

	// Получаем информацию о подписке
	subscription, err := bot.SubscriptionService.GetUserSubscription(userID)
	if err != nil {
		log.Printf("❌ Error getting subscription for user %d: %v", userID, err)
		subscription = nil
	}

	var messageText string
	var keyboard tgbotapi.InlineKeyboardMarkup

	if subscription != nil && subscription.Status == "suspended" {
		// Подписка приостановлена - предлагаем восстановить
		messageText = "🔄 *Восстановление подписки*\n\n" +
			"Ваша подписка была приостановлена после 3 неудачных попыток списания.\n\n" +
			"Для восстановления доступа используйте новую карту:"

		keyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("💳 Использовать новую карту", "change_payment_method"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🔙 Назад в меню", "main_menu"),
			),
		)
	} else if subscription != nil && subscription.FailedAttempts > 0 {
		// Подписка имеет неудачные попытки - пытаемся повторить
		messageText = "🔄 *Повторная попытка списания*\n\n" +
			"Запускаем повторную попытку списания с вашей карты..."

		// Пытаемся повторить списание
		err := bot.SubscriptionService.RetryPayment(userID)
		if err != nil {
			messageText = "❌ *Ошибка повторной попытки*\n\n" +
				"Не удалось запустить повторную попытку: " + err.Error() + "\n\n" +
				"Попробуйте позже или используйте новую карту."

			keyboard = tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("💳 Использовать новую карту", "change_payment_method"),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("🔙 Назад в меню", "main_menu"),
				),
			)
		} else {
			messageText = "🔄 *Повторная попытка списания*\n\n" +
				"Запущена повторная попытка списания с вашей карты.\n" +
				"Вы получите уведомление о результате."

			keyboard = tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("🔙 Назад в меню", "main_menu"),
				),
			)
		}
	} else {
		// Подписка не имеет неудачных попыток
		messageText = "ℹ️ *Информация*\n\n" +
			"У вашей подписки нет неудачных попыток списания.\n" +
			"Если у вас возникли проблемы, обратитесь в поддержку."

		keyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🔙 Назад в меню", "main_menu"),
			),
		)
	}

	msg := tgbotapi.NewEditMessageText(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		messageText,
	)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = &keyboard
	bot.Send(msg)
}

// handleChangePaymentMethod обрабатывает кнопку изменения способа оплаты
func (ih *InlineHandler) handleChangePaymentMethod(bot *Bot, callback *tgbotapi.CallbackQuery) {
	userID := callback.From.ID

	// Получаем информацию о подписке
	subscription, err := bot.SubscriptionService.GetUserSubscription(userID)
	if err != nil {
		log.Printf("❌ Error getting subscription for user %d: %v", userID, err)
		subscription = nil
	}

	var messageText string
	var keyboard tgbotapi.InlineKeyboardMarkup

	if subscription != nil && subscription.Status == "suspended" {
		// Подписка приостановлена - предлагаем восстановить
		messageText = "💳 *Восстановление подписки*\n\n" +
			"Ваша подписка была приостановлена после 3 неудачных попыток списания.\n\n" +
			"Для восстановления доступа используйте новую карту:"

		// Получаем новую ссылку для оплаты
		paymentURL, err := bot.SubscriptionService.ChangePaymentMethod(userID)
		if err != nil {
			messageText = "❌ *Ошибка восстановления*\n\n" +
				"Не удалось создать ссылку для оплаты: " + err.Error() + "\n\n" +
				"Попробуйте позже или обратитесь в поддержку."

			keyboard = tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("🔄 Попробовать снова", "change_payment_method"),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("🔙 Назад в меню", "main_menu"),
				),
			)
		} else {
			keyboard = tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonURL("💳 Оплатить новой картой", paymentURL),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("🔙 Назад в меню", "main_menu"),
				),
			)
		}
	} else {
		// Подписка активна или не найдена - предлагаем стандартную покупку
		messageText = "💳 *Изменение способа оплаты*\n\n" +
			"Для изменения способа оплаты перейдите к покупке премиум подписки."

		keyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("💳 Перейти к оплате", "buy_premium"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🔙 Назад в меню", "main_menu"),
			),
		)
	}

	msg := tgbotapi.NewEditMessageText(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		messageText,
	)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = &keyboard
	bot.Send(msg)
}

// handleUnknownCallback обрабатывает неизвестные callback
func (ih *InlineHandler) handleUnknownCallback(bot *Bot, callback *tgbotapi.CallbackQuery) {
	msg := tgbotapi.NewEditMessageText(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		"❌ Неизвестное действие",
	)
	bot.Send(msg)
}

// checkUserSubscriptionStatus проверяет статус подписки пользователя
func (ih *InlineHandler) checkUserSubscriptionStatus(userID int64) (string, bool, int, error) {
	// Получаем подписку пользователя
	subscription, err := ih.subscriptionService.GetUserSubscription(userID)
	if err != nil {
		return "error", false, 0, err
	}

	// Если нет подписки
	if subscription == nil {
		// Проверяем бесплатные создания
		remainingFree, err := ih.getRemainingFreeCreations(userID)
		if err != nil {
			return "no_subscription", false, 0, err
		}
		return "no_subscription", remainingFree > 0, remainingFree, nil
	}

	// Если подписка активна
	if subscription.Status == "active" && subscription.Active {
		return "active", true, 0, nil
	}

	// Если подписка отменена, проверяем grace period (30 дней)
	if subscription.Status == "cancelled" {
		if subscription.CancelledAt != nil {
			gracePeriodEnd := subscription.CancelledAt.AddDate(0, 0, 30)
			if time.Now().Before(gracePeriodEnd) {
				return "cancelled", true, 0, nil
			}
		}
		// Grace period истек, проверяем бесплатные создания
		remainingFree, err := ih.getRemainingFreeCreations(userID)
		if err != nil {
			return "cancelled", false, 0, err
		}
		return "cancelled", remainingFree > 0, remainingFree, nil
	}

	// Если подписка истекла или имеет статус, который не дает неограниченного доступа
	// Проверяем бесплатные создания
	remainingFree, err := ih.getRemainingFreeCreations(userID)
	if err != nil {
		return "expired", false, 0, err
	}
	return "expired", remainingFree > 0, remainingFree, nil
}

// getRemainingFreeCreations возвращает количество оставшихся бесплатных созданий
func (ih *InlineHandler) getRemainingFreeCreations(userID int64) (int, error) {
	// Получаем количество использований за текущий месяц из usage_stats
	usageThisMonth, err := ih.stateManager.db.GetUserUsageThisMonth(userID)
	if err != nil {
		log.Printf("Ошибка получения количества использований за месяц: %v", err)
		return 0, err
	}

	// Максимум бесплатных созданий в месяц - 5
	const maxFreePostsPerMonth = 5
	remaining := maxFreePostsPerMonth - usageThisMonth

	if remaining < 0 {
		remaining = 0
	}

	log.Printf("[DEBUG] У пользователя %d использований за месяц: %d, осталось бесплатных: %d", userID, usageThisMonth, remaining)

	return remaining, nil
}

// createSubscriptionKeyboard создает клавиатуру для подписки
func (ih *InlineHandler) createSubscriptionKeyboard(userID int64, subscriptionStatus string, remainingFree int) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	// Кнопка оформления подписки
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("💳 Оформить подписку", "buy_premium"),
	))

	// Если есть бесплатные создания, добавляем кнопку продолжить
	if remainingFree > 0 {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🎁 Продолжить создание", "create_post"),
		))
	}

	// Кнопка возврата в главное меню
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🏠 Главное меню", "main_menu"),
	))

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// handlePostHistory обрабатывает просмотр истории постов
func (ih *InlineHandler) handlePostHistory(bot *Bot, callback *tgbotapi.CallbackQuery, page int) {
	userID := callback.From.ID
	const postsPerPage = 10
	offset := (page - 1) * postsPerPage

	// Получаем посты пользователя с is_saved=TRUE
	posts, err := ih.postHistoryRepo.GetUserSavedPosts(userID, postsPerPage, offset)
	if err != nil {
		log.Printf("Ошибка получения истории постов: %v", err)
		msg := tgbotapi.NewEditMessageText(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
			"❌ Ошибка получения истории постов",
		)
		bot.Send(msg)
		return
	}

	log.Printf("Получено постов: %d", len(posts))
	for i, post := range posts {
		log.Printf("Пост %d: ID=%d, AIResponse=%s", i+1, post.ID, post.AIResponse)
	}

	// Получаем общее количество постов для пагинации
	totalPosts, err := ih.postHistoryRepo.GetUserSavedPostsCount(userID)
	if err != nil {
		log.Printf("Ошибка получения количества постов: %v", err)
	}

	var messageText string
	if len(posts) == 0 {
		messageText = "📚 История постов\n\nУ вас пока нет сохраненных постов."
	} else {
		messageText = fmt.Sprintf("📚 История постов (страница %d)\n\n", page)
		for i, post := range posts {
			postNumber := offset + i + 1
			// Очищаем HTML теги из текста поста
			formatter := NewTelegramPostFormatter(DefaultPostStyling())
			cleanText, _ := formatter.ParseHTMLToEntities(post.AIResponse)

			// Безопасно обрезаем очищенный текст до 30 символов с проверкой UTF-8
			shortText := cleanText
			if len(shortText) > 30 {
				// Проверяем, что обрезание не нарушает UTF-8
				runes := []rune(shortText)
				if len(runes) > 30 {
					shortText = string(runes[:30]) + "..."
				}
			}
			// Очищаем текст от недопустимых символов
			shortText = strings.Map(func(r rune) rune {
				if r < 32 && r != '\n' && r != '\t' {
					return -1 // Удаляем управляющие символы
				}
				return r
			}, shortText)
			messageText += fmt.Sprintf("%d. %s\n", postNumber, shortText)
		}
	}

	// Создаем клавиатуру
	var keyboardRows [][]tgbotapi.InlineKeyboardButton

	// Кнопки с номерами постов
	if len(posts) > 0 {
		var postButtons []tgbotapi.InlineKeyboardButton
		for i := range posts {
			postNumber := offset + i + 1
			postButtons = append(postButtons, tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%d", postNumber),
				fmt.Sprintf("view_post_%d", postNumber),
			))
		}
		// Разбиваем кнопки на ряды по 5
		for i := 0; i < len(postButtons); i += 5 {
			end := i + 5
			if end > len(postButtons) {
				end = len(postButtons)
			}
			keyboardRows = append(keyboardRows, postButtons[i:end])
		}
	}

	// Кнопки пагинации
	if totalPosts > postsPerPage {
		var paginationRow []tgbotapi.InlineKeyboardButton

		if page > 1 {
			paginationRow = append(paginationRow, tgbotapi.NewInlineKeyboardButtonData(
				"⬅️ Назад",
				fmt.Sprintf("post_history_%d", page-1),
			))
		}

		if offset+postsPerPage < totalPosts {
			paginationRow = append(paginationRow, tgbotapi.NewInlineKeyboardButtonData(
				"Вперед ➡️",
				fmt.Sprintf("post_history_%d", page+1),
			))
		}

		if len(paginationRow) > 0 {
			keyboardRows = append(keyboardRows, paginationRow)
		}
	}

	// Кнопка возврата в профиль
	keyboardRows = append(keyboardRows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🔙 Назад в профиль", "profile"),
	))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(keyboardRows...)

	msg := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, messageText)
	msg.ReplyMarkup = &keyboard
	bot.Send(msg)
}

// handleViewPost обрабатывает просмотр конкретного поста
func (ih *InlineHandler) handleViewPost(bot *Bot, callback *tgbotapi.CallbackQuery, postNumber int) {
	userID := callback.From.ID
	const postsPerPage = 10

	// Вычисляем страницу и позицию поста
	page := ((postNumber - 1) / postsPerPage) + 1
	positionOnPage := (postNumber - 1) % postsPerPage
	offset := (page - 1) * postsPerPage

	// Получаем посты для страницы
	posts, err := ih.postHistoryRepo.GetUserSavedPosts(userID, postsPerPage, offset)
	if err != nil {
		log.Printf("Ошибка получения постов: %v", err)
		msg := tgbotapi.NewEditMessageText(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
			"❌ Ошибка получения поста",
		)
		bot.Send(msg)
		return
	}

	// Проверяем, что пост существует
	if positionOnPage >= len(posts) {
		msg := tgbotapi.NewEditMessageText(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
			"❌ Пост не найден",
		)
		bot.Send(msg)
		return
	}

	post := posts[positionOnPage]

	// Формируем заголовок поста
	headerText := fmt.Sprintf("📚 Пост #%d\n\n", postNumber)

	// Обрабатываем HTML разметку в тексте поста
	formatter := NewTelegramPostFormatter(DefaultPostStyling())
	cleanText, entities := formatter.ParseHTMLToEntities(post.AIResponse)

	// Объединяем заголовок с очищенным текстом
	fullText := headerText + cleanText

	// Создаем клавиатуру с кнопкой назад
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 Назад к истории", fmt.Sprintf("post_history_%d", page)),
		),
	)

	// Отправляем с форматированием
	_, err = bot.SendFormattedMessageWithKeyboard(
		callback.Message.Chat.ID,
		fullText,
		entities,
		keyboard,
	)
	if err != nil {
		log.Printf("Ошибка отправки форматированного сообщения: %v", err)
		// Отправляем без форматирования в случае ошибки
		msg := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, headerText+post.AIResponse)
		msg.ReplyMarkup = &keyboard
		bot.Send(msg)
	} else {
		// Удаляем старое сообщение, так как отправили новое
		deleteMsg := tgbotapi.NewDeleteMessage(callback.Message.Chat.ID, callback.Message.MessageID)
		bot.Send(deleteMsg)
	}
}

// handlePaymentHistory обрабатывает просмотр истории оплат пользователя
func (ih *InlineHandler) handlePaymentHistory(bot *Bot, callback *tgbotapi.CallbackQuery) {
	userID := callback.From.ID

	// Получаем историю всех платежей пользователя
	subscriptions, err := ih.subscriptionService.GetUserPaymentHistory(userID)
	if err != nil {
		log.Printf("Ошибка получения истории оплат: %v", err)
		msg := tgbotapi.NewEditMessageText(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
			"❌ Ошибка получения истории оплат",
		)
		bot.Send(msg)
		return
	}

	var messageText string
	if len(subscriptions) == 0 {
		messageText = "💰 История оплат\n\nУ вас пока нет истории платежей."
	} else {
		messageText = "💰 История оплат\n\n"
		for i, sub := range subscriptions {
			// Форматируем дату создания
			createdDate := sub.CreatedAt.Format("02.01.2006 15:04")

			// Форматируем статус
			var statusEmoji, statusText string
			switch sub.Status {
			case "active":
				statusEmoji = "✅"
				statusText = "Активна"
			case "pending":
				statusEmoji = "⏳"
				statusText = "Ожидает оплаты"
			case "cancelled":
				statusEmoji = "❌"
				statusText = "Отменена"
			case "expired":
				statusEmoji = "⏰"
				statusText = "Истекла"
			case "suspended":
				statusEmoji = "🚫"
				statusText = "Приостановлена"
			default:
				statusEmoji = "❓"
				statusText = sub.Status
			}

			// Форматируем сумму
			amountText := fmt.Sprintf("%.0f₽", sub.Amount)

			// Форматируем тариф
			tariffText := sub.Tariff
			if tariffText == "" {
				tariffText = "Не указан"
			}

			messageText += fmt.Sprintf("%d. %s %s\n", i+1, statusEmoji, statusText)
			messageText += fmt.Sprintf("   💰 Сумма: %s\n", amountText)
			messageText += fmt.Sprintf("   📅 Дата: %s\n", createdDate)
			messageText += fmt.Sprintf("   🏷️ Тариф: %s\n", tariffText)

			// Добавляем дополнительную информацию в зависимости от статуса
			if sub.Status == "active" && !sub.NextPayment.IsZero() {
				nextPay := sub.NextPayment.Format("02.01.2006")
				messageText += fmt.Sprintf("   ⏰ Следующий платеж: %s\n", nextPay)
			}
			if sub.CancelledAt != nil {
				cancelledDate := sub.CancelledAt.Format("02.01.2006")
				messageText += fmt.Sprintf("   🚫 Отменена: %s\n", cancelledDate)
			}
			if sub.FailedAttempts > 0 {
				messageText += fmt.Sprintf("   ⚠️ Неудачных попыток: %d\n", sub.FailedAttempts)
			}

			messageText += "\n"
		}
	}

	// Создаем клавиатуру с кнопкой возврата
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 Назад в профиль", "profile"),
		),
	)

	msg := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, messageText)
	msg.ReplyMarkup = &keyboard
	bot.Send(msg)
}

// shouldIncrementUsage проверяет, нужно ли увеличивать счетчик использований
// Возвращает true, если счетчик НЕ нужно увеличивать (есть активная подписка или grace period)
func (ih *InlineHandler) shouldIncrementUsage(userID int64) (bool, error) {
	// Получаем подписку пользователя
	subscription, err := ih.subscriptionService.GetUserSubscription(userID)
	if err != nil {
		log.Printf("Ошибка получения подписки для проверки счетчика: %v", err)
		return false, err
	}

	// Если нет подписки, нужно увеличивать счетчик
	if subscription == nil {
		log.Printf("[DEBUG] У пользователя %d нет подписки - увеличиваем счетчик", userID)
		return false, nil
	}

	log.Printf("[DEBUG] Проверка счетчика для пользователя %d: status=%s, active=%v, cancelled_at=%v, next_payment=%v",
		userID, subscription.Status, subscription.Active, subscription.CancelledAt, subscription.NextPayment)

	// Если подписка активна, НЕ нужно увеличивать счетчик
	if subscription.Status == "active" && subscription.Active {
		log.Printf("[DEBUG] У пользователя %d активная подписка - НЕ увеличиваем счетчик", userID)
		return true, nil
	}

	// Если подписка отменена, проверяем grace period
	if subscription.Status == "cancelled" {
		if subscription.CancelledAt != nil {
			gracePeriodEnd := subscription.CancelledAt.AddDate(0, 0, 30)
			now := time.Now()

			// Если grace period активен И подписка еще не истекла, НЕ нужно увеличивать счетчик
			if now.Before(gracePeriodEnd) && now.Before(subscription.NextPayment) {
				log.Printf("[DEBUG] У пользователя %d активный grace period - НЕ увеличиваем счетчик", userID)
				return true, nil
			} else {
				log.Printf("[DEBUG] У пользователя %d grace period истек или подписка истекла - увеличиваем счетчик", userID)
			}
		}
	}

	// Во всех остальных случаях нужно увеличивать счетчик
	log.Printf("[DEBUG] У пользователя %d нет активной подписки или grace period - увеличиваем счетчик", userID)
	return false, nil
}

// safeIncrementUsage безопасно увеличивает счетчик использований с проверкой подписки
func (ih *InlineHandler) safeIncrementUsage(userID int64) error {
	// Проверяем, нужно ли увеличивать счетчик использований
	shouldNotIncrement, err := ih.shouldIncrementUsage(userID)
	if err != nil {
		log.Printf("Ошибка проверки необходимости увеличения счетчика: %v", err)
		// В случае ошибки увеличиваем счетчик для безопасности
		return ih.stateManager.IncrementUsage(userID)
	}

	if !shouldNotIncrement {
		// Увеличиваем счетчик только если нет активной подписки или grace period
		return ih.stateManager.IncrementUsage(userID)
	} else {
		log.Printf("Счетчик не увеличен - у пользователя активная подписка или grace period")
		return nil
	}
}

// incrementUsageIfNeeded увеличивает счетчик использований только если это необходимо
// Используется для замены прямых вызовов stateManager.IncrementUsage
func (ih *InlineHandler) incrementUsageIfNeeded(userID int64) {
	err := ih.safeIncrementUsage(userID)
	if err != nil {
		log.Printf("Ошибка увеличения счетчика: %v", err)
	}
}

// handleRewritePostStart обрабатывает начало процесса рерайта поста
func (ih *InlineHandler) handleRewritePostStart(bot *Bot, callback *tgbotapi.CallbackQuery) {
	userID := callback.From.ID

	// Проверяем подписку пользователя (так же как в handleCreatePost)
	subscriptionStatus, canCreate, remainingFree, err := ih.checkUserSubscriptionStatus(userID)
	if err != nil {
		log.Printf("Ошибка проверки подписки для рерайта: %v", err)
		// В случае ошибки разрешаем создание
		subscriptionStatus = "error"
		canCreate = true
	}

	if canCreate {
		// Устанавливаем состояние ожидания текста поста
		ih.stateManager.SetWaitingForPostText(userID, true)
		ih.stateManager.UpdateStep(userID, "waiting_for_post_text")

		msg := tgbotapi.NewEditMessageText(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
			"📝 Отправьте или перешлите пост, который хотите переписать:\n\n"+
				"Просто скопируйте текст поста и отправьте его в чат, или перешлите сообщение с постом.",
		)

		bot.Send(msg)
	} else {
		// Показываем информацию о подписке и предлагаем оформить
		keyboard := ih.createSubscriptionKeyboard(userID, subscriptionStatus, remainingFree)

		var messageText string
		switch subscriptionStatus {
		case "cancelled":
			messageText = "❌ Ваша подписка была отменена.\n\n"
		case "expired":
			messageText = "⏰ Срок действия подписки истек.\n\n"
		case "no_subscription":
			messageText = "💎 У вас нет активной подписки.\n\n"
		default:
			messageText = "💎 Требуется подписка для создания контента.\n\n"
		}

		if remainingFree > 0 {
			messageText += fmt.Sprintf("🎁 У вас осталось %d бесплатных созданий в этом месяце.\n\n", remainingFree)
		} else {
			messageText += "🎁 Бесплатные создания на этот месяц закончились.\n\n"
		}

		messageText += "💳 Оформите подписку для неограниченного создания контента!"

		msg := tgbotapi.NewEditMessageText(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
			messageText,
		)
		msg.ReplyMarkup = &keyboard

		bot.Send(msg)
	}
}

// handleRewritePostDirect обрабатывает прямой рерайт поста без дополнительных указаний
func (ih *InlineHandler) handleRewritePostDirect(bot *Bot, callback *tgbotapi.CallbackQuery) {
	userID := callback.From.ID
	state := ih.stateManager.GetState(userID)

	// Получаем текст поста для рерайта
	originalText := ih.stateManager.GetRewritingPost(userID)
	if originalText == "" {
		msg := tgbotapi.NewEditMessageText(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
			"❌ Ошибка: текст поста не найден. Попробуйте начать заново.",
		)
		bot.Send(msg)
		return
	}

	// Устанавливаем режим рерайта
	ih.stateManager.SetRewriteMode(userID, "direct")

	// Отправляем сообщение о начале обработки
	msg := tgbotapi.NewEditMessageText(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		"⏳ Переписываю пост...",
	)
	bot.Send(msg)

	// Выполняем рерайт через DeepSeek
	rewrittenText, err := ih.voiceHandler.GenerateContent("rewrite_post", originalText, userID, 0)
	if err != nil {
		log.Printf("Ошибка рерайта поста: %v", err)
		msg := tgbotapi.NewMessage(
			callback.Message.Chat.ID,
			"❌ Не удалось переписать пост. Попробуйте еще раз.",
		)
		bot.Send(msg)
		return
	}

	// Форматируем результат
	formatter := NewTelegramPostFormatter(state.PostStyling)
	cleanText, entities := formatter.FormatPost(rewrittenText)

	// Создаем пост
	post := Post{
		ContentType: "telegram_post",
		Content:     cleanText,
		Messages:    []string{originalText},
		Entities:    entities,
		Styling:     state.PostStyling,
	}

	// Сохраняем пост
	ih.stateManager.SetCurrentPost(userID, &post)
	ih.stateManager.SetApprovalStatus(userID, "pending")

	// Отправляем результат с кнопками согласования
	keyboard := bot.CreateApprovalKeyboard()
	messageID, err := bot.SendFormattedMessageWithKeyboard(
		callback.Message.Chat.ID,
		cleanText,
		entities,
		keyboard,
	)
	if err != nil {
		log.Printf("Ошибка отправки форматированного сообщения: %v", err)
		// Отправляем без форматирования в случае ошибки
		resultMsg := tgbotapi.NewMessage(callback.Message.Chat.ID, cleanText)
		resultMsg.ReplyMarkup = keyboard
		bot.Send(resultMsg)
	} else {
		// Сохраняем ID сообщения с готовым постом
		ih.stateManager.SetPostMessageID(userID, messageID)
		log.Printf("Сохранили ID сообщения с рерайтом: %d", messageID)
	}

	// Очищаем состояние рерайта
	ih.stateManager.ClearRewriteState(userID)
}

// handleRewritePostVoice обрабатывает рерайт поста с голосовыми указаниями
func (ih *InlineHandler) handleRewritePostVoice(bot *Bot, callback *tgbotapi.CallbackQuery) {
	userID := callback.From.ID

	// Получаем текст поста для рерайта
	originalText := ih.stateManager.GetRewritingPost(userID)
	if originalText == "" {
		msg := tgbotapi.NewEditMessageText(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
			"❌ Ошибка: текст поста не найден. Попробуйте начать заново.",
		)
		bot.Send(msg)
		return
	}

	// Устанавливаем режим рерайта и состояние ожидания голосовых сообщений
	ih.stateManager.SetRewriteMode(userID, "voice")
	ih.stateManager.UpdateStep(userID, "waiting_for_voice")
	ih.stateManager.SetWaitingForVoice(userID, true)
	ih.stateManager.ClearVoiceMessages(userID)
	ih.stateManager.ClearPendingVoices(userID)

	msg := tgbotapi.NewEditMessageText(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		"🎤 Отправьте голосовые сообщения с указаниями, как переписать пост:\n\n"+
			"Исходный пост:\n"+originalText+"\n\n"+
			"Говорите, что именно нужно изменить, добавить или убрать.",
	)

	bot.Send(msg)
}
