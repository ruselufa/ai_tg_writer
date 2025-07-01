package main

import (
	"log"
	"os"

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
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = true
	log.Printf("Бот %s запущен", bot.Self.UserName)

	// Создаем обработчик голосовых сообщений
	voiceHandler := voice.NewVoiceHandler(bot)

	// Настраиваем обновления
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates := bot.GetUpdatesChan(updateConfig)

	// Обрабатываем сообщения
	for update := range updates {
		if update.Message == nil {
			continue
		}

		// Обрабатываем команды и сообщения
		handleMessage(bot, update.Message, voiceHandler)
	}
}

func handleMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message, voiceHandler *voice.VoiceHandler) {
	// Логируем входящие сообщения
	log.Printf("[%s] %s", message.From.UserName, message.Text)

	// Обрабатываем команды
	if message.IsCommand() {
		handleCommand(bot, message)
		return
	}

	// Обрабатываем голосовые сообщения
	if message.Voice != nil {
		handleVoiceMessage(bot, message, voiceHandler)
		return
	}

	// Обрабатываем обычные текстовые сообщения
	handleTextMessage(bot, message)
}

func handleCommand(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
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

func handleVoiceMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message, voiceHandler *voice.VoiceHandler) {
	// Отправляем сообщение о том, что обрабатываем голосовое
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
}

func handleTextMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
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

func sendWelcomeMessage(bot *tgbotapi.BotAPI, chatID int64) {
	text := `🎉 Добро пожаловать в AI Voice Writer!

Я помогу вам превратить голосовые сообщения в красивый, структурированный текст.

📝 Как использовать:
• Отправьте мне голосовое сообщение
• Я распознаю речь и перепишу её с помощью ИИ
• Получите готовый текст

💡 Доступные команды:
/help - Подробная справка
/profile - Ваш профиль и лимиты
/subscription - Информация о подписке

🎤 Отправьте голосовое сообщение прямо сейчас!`

	msg := tgbotapi.NewMessage(chatID, text)
	bot.Send(msg)
}

func sendHelpMessage(bot *tgbotapi.BotAPI, chatID int64) {
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
• Статистика использования

💳 Подписка (/subscription):
• Информация о тарифах
• Способы оплаты
• Партнёрская программа

❓ Если у вас есть вопросы, обратитесь к администратору.`

	msg := tgbotapi.NewMessage(chatID, text)
	bot.Send(msg)
}

func sendProfileMessage(bot *tgbotapi.BotAPI, chatID int64, userID int64) {
	// TODO: Получать данные из базы данных
	text := `👤 Ваш профиль

🆔 ID пользователя: ` + string(rune(userID)) + `

📊 Тариф: Бесплатный
📈 Использовано сегодня: 0/5
📅 Сброс лимита: каждый день в 00:00

💡 Хотите больше возможностей?
Используйте /subscription для перехода на премиум!`

	msg := tgbotapi.NewMessage(chatID, text)
	bot.Send(msg)
}

func sendSubscriptionMessage(bot *tgbotapi.BotAPI, chatID int64) {
	text := `💳 Информация о подписке

🎯 Доступные тарифы:

🆓 Бесплатный:
• 5 голосовых сообщений в день
• Базовое распознавание речи
• Стандартное переписывание

⭐ Премиум (скоро):
• Неограниченное количество сообщений
• Высокое качество распознавания
• Продвинутое переписывание с ИИ
• Приоритетная поддержка

🤝 Партнёрская программа (скоро):
• Реферальные ссылки
• Комиссия за привлечённых пользователей
• Статистика и аналитика

📞 Для подключения премиум тарифа свяжитесь с администратором:
@admin_username

💡 Следите за обновлениями!`

	msg := tgbotapi.NewMessage(chatID, text)
	bot.Send(msg)
}

func sendUnknownCommandMessage(bot *tgbotapi.BotAPI, chatID int64) {
	text := `❓ Неизвестная команда

Доступные команды:
/start - Начать работу
/help - Помощь
/profile - Ваш профиль
/subscription - Подписка

🎤 Или просто отправьте голосовое сообщение!`

	msg := tgbotapi.NewMessage(chatID, text)
	bot.Send(msg)
}
