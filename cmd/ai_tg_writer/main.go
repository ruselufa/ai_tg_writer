package main

import (
	"log"
	"os"
	"strconv"

	"ai_tg_writer/internal/infrastructure/bot"
	"ai_tg_writer/internal/infrastructure/voice"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func main() {
	// Загружаем переменные окружения
	if err := godotenv.Load(); err != nil {
		log.Println("Файл .env не найден, используем системные переменные")
	}

	// Получаем токен бота из переменных окружения
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN не установлен")
	}

	// Создаем экземпляр бота
	botAPI, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}

	botAPI.Debug = true
	log.Printf("Бот %s запущен", botAPI.Self.UserName)

	// Создаем обработчики
	customBot := bot.NewBot(botAPI)
	voiceHandler := voice.NewVoiceHandler(botAPI)
	stateManager := bot.NewStateManager()
	inlineHandler := bot.NewInlineHandler(stateManager, voiceHandler)
	messageHandler := bot.NewMessageHandler(stateManager, voiceHandler)

	// Настраиваем обновления
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates := botAPI.GetUpdatesChan(updateConfig)

	// Обрабатываем сообщения
	for update := range updates {
		// Обрабатываем callback от инлайн-кнопок
		if update.CallbackQuery != nil {
			inlineHandler.HandleCallback(customBot, update.CallbackQuery)
			continue
		}

		if update.Message == nil {
			continue
		}

		// Обрабатываем голосовые сообщения через MessageHandler
		if update.Message.Voice != nil {
			messageHandler.HandleMessage(customBot, update.Message)
			continue
		}

		// Обрабатываем команды и текстовые сообщения
		handleMessage(customBot, update.Message, voiceHandler, stateManager, inlineHandler)
	}
}

// handleMessage теперь не обрабатывает голосовые сообщения напрямую
func handleMessage(bot *bot.Bot, message *tgbotapi.Message, voiceHandler *voice.VoiceHandler, stateManager *bot.StateManager, inlineHandler *bot.InlineHandler) {
	// Логируем входящие сообщения
	log.Printf("[%s] %s", message.From.UserName, message.Text)

	// Обрабатываем команды
	if message.IsCommand() {
		handleCommand(bot, message)
		return
	}

	// Обрабатываем обычные текстовые сообщения
	handleTextMessage(bot, message, stateManager)
}

func handleCommand(bot *bot.Bot, message *tgbotapi.Message) {
	switch message.Command() {
	case "start":
		sendWelcomeMessage(bot, message.Chat.ID)
	case "help":
		sendHelpMessage(bot, message.Chat.ID)
	case "profile":
		sendProfileMessage(bot, message.Chat.ID, message.From.ID)
	case "subscription":
		sendSubscriptionMessage(bot, message.Chat.ID)
	default:
		sendUnknownCommandMessage(bot, message.Chat.ID)
	}
}

func handleVoiceMessage(bot *bot.Bot, message *tgbotapi.Message, voiceHandler *voice.VoiceHandler, stateManager *bot.StateManager, inlineHandler *bot.InlineHandler) {
	userID := message.From.ID
	state := stateManager.GetState(userID)

	// Проверяем, ожидает ли бот голосовое сообщение
	if !state.WaitingForVoice {
		// Если не ожидаем голосовое, отправляем стандартную обработку
		processingMsg := tgbotapi.NewMessage(message.Chat.ID, "🎵 Обрабатываю ваше голосовое сообщение...")
		processingMsg.ReplyToMessageID = message.MessageID
		bot.Send(processingMsg)

		// Обрабатываем голосовое сообщение
		resultText, err := voiceHandler.ProcessVoiceMessage(message)
		if err != nil {
			log.Printf("Ошибка обработки голосового сообщения: %v", err)
			errorMsg := tgbotapi.NewMessage(message.Chat.ID, "❌ Произошла ошибка при обработке голосового сообщения. Попробуйте еще раз.")
			bot.Send(errorMsg)
			return
		}

		// Отправляем результат
		resultMsg := tgbotapi.NewMessage(message.Chat.ID, resultText)
		bot.Send(resultMsg)
		return
	}

	// Если ожидаем голосовое сообщение в рамках создания контента
	processingMsg := tgbotapi.NewMessage(message.Chat.ID, "🎵 Обрабатываю ваше голосовое сообщение...")
	processingMsg.ReplyToMessageID = message.MessageID
	bot.Send(processingMsg)

	// Обрабатываем голосовое сообщение
	resultText, err := voiceHandler.ProcessVoiceMessage(message)
	if err != nil {
		log.Printf("Ошибка обработки голосового сообщения: %v", err)
		errorMsg := tgbotapi.NewMessage(message.Chat.ID, "❌ Произошла ошибка при обработке голосового сообщения. Попробуйте еще раз.")
		bot.Send(errorMsg)
		return
	}

	// Добавляем голосовое сообщение к состоянию пользователя
	stateManager.AddVoiceMessage(userID, resultText)

	// Отправляем сообщение с кнопками для продолжения
	keyboard := bot.CreateContinueKeyboard()
	msg := tgbotapi.NewMessage(message.Chat.ID, "✅ Принято. Хотите продолжить диктовку или уже начинать создание текста?")
	msg.ReplyMarkup = &keyboard
	bot.Send(msg)

	// Обновляем состояние
	stateManager.SetWaitingForVoice(userID, false)
}

func handleTextMessage(bot *bot.Bot, message *tgbotapi.Message, stateManager *bot.StateManager) {
	userID := message.From.ID
	state := stateManager.GetState(userID)

	// Если пользователь в процессе создания контента, отправляем подсказку
	if state.CurrentStep != "idle" {
		response := tgbotapi.NewMessage(message.Chat.ID,
			"🎤 Отправьте голосовое сообщение с вашими идеями для создания контента.")
		bot.Send(response)
		return
	}

	// Отправляем информацию о том, что бот принимает голосовые сообщения
	response := tgbotapi.NewMessage(message.Chat.ID,
		"👋 Отправьте мне голосовое сообщение, и я перепишу его в красивый текст!\n\n"+
			"Доступные команды:\n"+
			"/start - Начать работу\n"+
			"/help - Помощь\n"+
			"/profile - Ваш профиль\n"+
			"/subscription - Подписка")

	bot.Send(response)
}

func sendWelcomeMessage(bot *bot.Bot, chatID int64) {
	text := `Привет! Я помогу тебе создать мощный контент из твоих идей. Выбери, что хочешь создать:`

	keyboard := bot.CreateMainKeyboard()
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = &keyboard
	bot.Send(msg)
}

func sendHelpMessage(bot *bot.Bot, chatID int64) {
	text := `📚 Справка по использованию бота

🎤 Голосовые сообщения:
• Отправьте голосовое сообщение любой длины
• Я распознаю речь и перепишу её красиво
• Поддерживаются все основные языки

📊 Лимиты использования:
• Бесплатный тариф: 5 сообщений в день
• Премиум тариф: неограниченно

👤 Профиль (/profile):
• Просмотр текущего тарифа
• Остаток использований
• Статистика использования`

	msg := tgbotapi.NewMessage(chatID, text)
	bot.Send(msg)
}

func sendProfileMessage(bot *bot.Bot, chatID int64, userID int64) {
	text := `👤 Ваш профиль

🆔 ID пользователя: ` + strconv.FormatInt(userID, 10) + `
📊 Тариф: Бесплатный
📈 Использовано сегодня: 0/5`

	msg := tgbotapi.NewMessage(chatID, text)
	bot.Send(msg)
}

func sendSubscriptionMessage(bot *bot.Bot, chatID int64) {
	text := `💎 Подписка

📊 Текущий тариф: Бесплатный
⏰ Срок действия: Бессрочно

✨ Премиум тариф:
• Неограниченное количество сообщений
• Приоритетная обработка
• Расширенные возможности редактирования
• Доступ к эксклюзивным функциям

💳 Стоимость: 299₽/месяц`

	msg := tgbotapi.NewMessage(chatID, text)
	bot.Send(msg)
}

func sendUnknownCommandMessage(bot *bot.Bot, chatID int64) {
	text := "❌ Неизвестная команда. Используйте /start для начала работы."

	msg := tgbotapi.NewMessage(chatID, text)
	bot.Send(msg)
}
