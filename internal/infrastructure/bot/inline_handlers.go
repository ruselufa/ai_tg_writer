package bot

import (
	"ai_tg_writer/internal/infrastructure/voice"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// InlineHandler обрабатывает inline-команды
type InlineHandler struct {
	stateManager *StateManager
	voiceHandler *voice.VoiceHandler
	prompts      map[string]Prompt
}

// NewInlineHandler создает новый обработчик inline-команд
func NewInlineHandler(stateManager *StateManager, voiceHandler *voice.VoiceHandler) *InlineHandler {
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
		stateManager: stateManager,
		voiceHandler: voiceHandler,
		prompts:      prompts,
	}
}

// HandleCallback обрабатывает callback от инлайн-кнопок
func (ih *InlineHandler) HandleCallback(bot *Bot, callback *tgbotapi.CallbackQuery) {
	log.Printf("Callback от пользователя %d: %s", callback.From.ID, callback.Data)

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
	case "subscription":
		ih.handleSubscription(bot, callback)
	case "styling_settings":
		ih.handleStylingSettings(bot, callback)
	case "test_formatting":
		ih.handleTestFormatting(bot, callback)
	case "no_action":
		// Игнорируем нажатие на пробел-заглушку
		return
	default:
		ih.handleUnknownCallback(bot, callback)
	}
}

// handleCreatePost обрабатывает выбор создания поста
func (ih *InlineHandler) handleCreatePost(bot *Bot, callback *tgbotapi.CallbackQuery) {
	userID := callback.From.ID

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

	// Создаем WaitGroup для ожидания обработки всех сообщений
	var wg sync.WaitGroup
	results := make([]string, 0)
	resultsMutex := &sync.Mutex{}

	// Обрабатываем каждое голосовое сообщение параллельно
	for fileID, voice := range state.PendingVoices {
		wg.Add(1)
		go func(fileID string, voice *VoiceTranscription) {
			defer wg.Done()

			// Транскрибируем файл
			text, err := ih.voiceHandler.TranscribeVoiceFile(voice.FilePath)
			if err != nil {
				log.Printf("Ошибка обработки голосового сообщения: %v", err)
				ih.stateManager.UpdateVoiceTranscription(userID, fileID, "", err)
				return
			}

			// Сохраняем результат
			resultsMutex.Lock()
			results = append(results, text)
			resultsMutex.Unlock()

			ih.stateManager.UpdateVoiceTranscription(userID, fileID, text, nil)

			// Удаляем временный файл
			if err := os.Remove(voice.FilePath); err != nil {
				log.Printf("Ошибка удаления временного файла %s: %v", voice.FilePath, err)
			}
		}(fileID, voice)
	}

	// Ждем завершения обработки всех сообщений
	wg.Wait()

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

	// Генерируем готовый пост через VoiceHandler
	postText, err := ih.voiceHandler.GenerateTelegramPost(allMessages)
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
	}

	// Сохраняем пост
	ih.stateManager.SetCurrentPost(userID, &post)
	ih.stateManager.SetApprovalStatus(userID, "pending")

	// Отправляем результат с кнопками согласования
	keyboard := bot.CreateApprovalKeyboard()
	err = bot.SendFormattedMessageWithKeyboard(
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

	// Полностью очищаем состояние
	ih.stateManager.UpdateStep(userID, "idle")
	ih.stateManager.SetCurrentPost(userID, nil)
	ih.stateManager.ClearVoiceMessages(userID)
	ih.stateManager.ClearPendingVoices(userID)
	ih.stateManager.ClearEditMessages(userID)
	ih.stateManager.ClearPendingEdits(userID)
	ih.stateManager.SetApprovalStatus(userID, "idle")
	ih.stateManager.SetWaitingForVoice(userID, false)

	// Отправляем главное меню
	text := "Привет! Я помогу тебе создать мощный контент из твоих идей. Выбери, что хочешь создать:"
	keyboard := bot.CreateMainKeyboard()

	msg := tgbotapi.NewEditMessageText(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
	)
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

	// Сохраняем пост в БД (заглушка)
	ih.stateManager.SavePost(userID, *state.CurrentPost)
	log.Printf("Пост сохранен в БД (заглушка): %s", state.CurrentPost.ContentType)

	// Очищаем состояние
	ih.stateManager.UpdateStep(userID, "idle")
	ih.stateManager.SetCurrentPost(userID, nil)
	ih.stateManager.ClearVoiceMessages(userID)
	ih.stateManager.ClearPendingVoices(userID)
	ih.stateManager.ClearEditMessages(userID)
	ih.stateManager.ClearPendingEdits(userID)
	ih.stateManager.SetApprovalStatus(userID, "approved")

	// Отправляем сообщение об успехе
	msg := tgbotapi.NewEditMessageText(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		"✅ Пост успешно сохранен! Текст остался в чате.\n\nПривет! Я помогу тебе создать мощный контент из твоих идей. Выбери, что хочешь создать:",
	)
	keyboard := bot.CreateMainKeyboard()
	msg.ReplyMarkup = &keyboard
	bot.Send(msg)
}

// handleApprove обрабатывает согласие с результатом
func (ih *InlineHandler) handleApprove(bot *Bot, callback *tgbotapi.CallbackQuery) {
	userID := callback.From.ID

	// Сохраняем пост в БД (заглушка)
	state := ih.stateManager.GetState(userID)
	if state.CurrentPost != nil {
		ih.stateManager.SavePost(userID, *state.CurrentPost)
		log.Printf("Пост сохранен в БД (заглушка): %s", state.CurrentPost.ContentType)
	}

	// Очищаем состояние
	ih.stateManager.UpdateStep(userID, "idle")
	ih.stateManager.SetCurrentPost(userID, nil)
	ih.stateManager.ClearVoiceMessages(userID)
	ih.stateManager.ClearPendingVoices(userID)
	ih.stateManager.SetApprovalStatus(userID, "approved")

	// Отправляем главное меню
	text := "✅ Пост сохранен! Текст остался в чате.\n\nПривет! Я помогу тебе создать мощный контент из твоих идей. Выбери, что хочешь создать:"
	keyboard := bot.CreateMainKeyboard()

	msg := tgbotapi.NewEditMessageText(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
	)
	msg.ReplyMarkup = &keyboard
	bot.Send(msg)
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

	// Создаем WaitGroup для ожидания обработки всех сообщений
	var wg sync.WaitGroup
	results := make([]string, 0)
	resultsMutex := &sync.Mutex{}

	// Обрабатываем каждое голосовое сообщение с правками параллельно
	for fileID, voice := range state.PendingEdits {
		wg.Add(1)
		go func(fileID string, voice *VoiceTranscription) {
			defer wg.Done()

			// Транскрибируем файл
			text, err := ih.voiceHandler.TranscribeVoiceFile(voice.FilePath)
			if err != nil {
				log.Printf("Ошибка обработки голосового сообщения с правками: %v", err)
				return
			}

			// Сохраняем результат
			resultsMutex.Lock()
			results = append(results, text)
			resultsMutex.Unlock()

			// Удаляем временный файл
			if err := os.Remove(voice.FilePath); err != nil {
				log.Printf("Ошибка удаления временного файла %s: %v", voice.FilePath, err)
			}
		}(fileID, voice)
	}

	// Ждем завершения обработки всех сообщений
	wg.Wait()

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

	// Генерируем обновленный пост через VoiceHandler
	updatedText, err := ih.voiceHandler.GenerateTelegramPost(prompt)
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
	ih.stateManager.SetLastGeneratedText(userID, updatedText)
	ih.stateManager.SetApprovalStatus(userID, "pending")

	// Отправляем обновленный результат с кнопками согласования
	keyboard := bot.CreateEditApprovalKeyboard()
	err = bot.SendFormattedMessageWithKeyboard(
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

	msg := tgbotapi.NewEditMessageText(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
	)

	bot.Send(msg)
}

// handleProfile обрабатывает кнопку профиля
func (ih *InlineHandler) handleProfile(bot *Bot, callback *tgbotapi.CallbackQuery) {
	text := `👤 Ваш профиль

🆔 ID пользователя: ` + strconv.FormatInt(callback.From.ID, 10) + `
📊 Тариф: Бесплатный
📈 Использовано сегодня: 0/5`

	msg := tgbotapi.NewEditMessageText(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
	)

	bot.Send(msg)
}

// handleSubscription обрабатывает кнопку подписки
func (ih *InlineHandler) handleSubscription(bot *Bot, callback *tgbotapi.CallbackQuery) {
	text := `💎 Подписка

📊 Текущий тариф: Бесплатный
⏰ Срок действия: Бессрочно

✨ Премиум тариф:
• Неограниченное количество сообщений
• Приоритетная обработка
• Расширенные возможности редактирования
• Доступ к эксклюзивным функциям

💳 Стоимость: 299₽/месяц`

	msg := tgbotapi.NewEditMessageText(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
	)

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
	err := bot.SendFormattedMessage(callback.Message.Chat.ID, cleanText, entities)
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

// handleUnknownCallback обрабатывает неизвестные callback
func (ih *InlineHandler) handleUnknownCallback(bot *Bot, callback *tgbotapi.CallbackQuery) {
	msg := tgbotapi.NewEditMessageText(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		"❌ Неизвестное действие",
	)
	bot.Send(msg)
}
