package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"ai_tg_writer/api"
	"ai_tg_writer/internal/config"
	"ai_tg_writer/internal/infrastructure/bot"
	"ai_tg_writer/internal/infrastructure/database"
	"ai_tg_writer/internal/infrastructure/voice"
	"ai_tg_writer/internal/infrastructure/yookassa"
	"ai_tg_writer/internal/service"
	"ai_tg_writer/internal/worker"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func main() {
	// Загружаем переменные окружения
	fmt.Println("Загружаем переменные окружения")
	if err := godotenv.Load(); err != nil {
		log.Println("Файл .env не найден, используем системные переменные")
	}

	// Получаем токен бота из переменных окружения
	fmt.Println("Получаем токен бота из переменных окружения")
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN не установлен")
	}
	fmt.Println("Токен бота: ", token)
	// Создаем экземпляр бота
	botAPI, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Экземпляр бота создан")
	botAPI.Debug = true
	log.Printf("Бот %s запущен", botAPI.Self.UserName)

	// Инициализируем подключение к базе данных
	fmt.Println("Инициализируем подключение к базе данных")
	db, err := database.NewConnection()
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	fmt.Println("Подключение к базе данных успешно")
	// Инициализируем таблицы
	if err := db.InitTables(); err != nil {
		log.Fatalf("Ошибка инициализации таблиц: %v", err)
	}
	fmt.Println("Таблицы инициализированы")

	// Создаем репозиторий подписок
	subscriptionRepo := database.NewSubscriptionRepository(db)

	// Создаем сервис подписок (временно без платежного модуля)
	// Загружаем конфигурацию
	cfg := config.NewConfig()
	log.Printf("📋 Configuration loaded: Mode=%s, SubscriptionInterval=%s, WorkerCheckInterval=%s",
		cfg.Mode, cfg.SubscriptionInterval, cfg.WorkerCheckInterval)

	// Инициализируем клиента YooKassa
	ykClient := yookassa.New()

	// Создаем временный сервис подписок для создания SubscriptionHandler
	tempSubscriptionService := service.NewSubscriptionService(subscriptionRepo, ykClient, cfg)

	// Создаем SubscriptionHandler для отправки сообщений
	subscriptionHandler := bot.NewSubscriptionHandler(tempSubscriptionService)

	// Создаем сервис подписок с ботом для отправки сообщений
	subscriptionService := service.NewSubscriptionServiceWithBot(subscriptionRepo, ykClient, cfg, subscriptionHandler)

	fmt.Println("Сервис подписок инициализирован")

	// Создаем обработчики
	customBot := bot.NewBotWithSubscriptionService(botAPI, db, subscriptionService)

	// Устанавливаем бота в SubscriptionHandler для отправки сообщений
	subscriptionHandler.SetBot(customBot)

	// Создаем HTTP-сервер для обработки платежей
	httpServer := api.NewServer("8080")
	httpServer.SetupRoutes(subscriptionService, nil, db, customBot)

	// Запускаем HTTP-сервер в горутине
	go func() {
		if err := httpServer.Start(); err != nil {
			log.Fatalf("Ошибка запуска HTTP-сервера: %v", err)
		}
	}()
	fmt.Println("HTTP-сервер запущен на порту 8080")

	// Создаем контекст для graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Запускаем воркер для рекуррентных платежей
	subscriptionWorker := worker.NewSubscriptionWorker(subscriptionService, cfg)
	subscriptionWorker.Start(ctx)

	// Настраиваем graceful shutdown
	setupGracefulShutdown(cancel)

	// Создаем репозиторий для истории постов
	postHistoryRepo := database.NewPostHistoryRepository(db.DB)
	
	voiceHandler := voice.NewVoiceHandler(botAPI, postHistoryRepo)
	stateManager := bot.NewStateManager(db)
	inlineHandler := bot.NewInlineHandler(stateManager, voiceHandler)
	messageHandler := bot.NewMessageHandler(stateManager, voiceHandler, inlineHandler)
	fmt.Println("Обработчики созданы")
	// Настраиваем обновления
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60
	fmt.Println("Настройки обновлений установлены")
	updates := botAPI.GetUpdatesChan(updateConfig)
	fmt.Println("Обновления получены")

	// Создаем семафор для ограничения одновременных обработок
	const maxConcurrentHandlers = 10
	semaphore := make(chan struct{}, maxConcurrentHandlers)
	fmt.Printf("🚦 Семафор создан с лимитом %d одновременных обработок\n", maxConcurrentHandlers)

	// Счетчик активных обработок
	var activeHandlers int32

	// Статистика производительности
	var totalProcessed int64
	var totalProcessingTime time.Duration
	var timeMutex sync.Mutex

	// Запускаем мониторинг производительности
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			processed := atomic.LoadInt64(&totalProcessed)
			active := atomic.LoadInt32(&activeHandlers)
			avgTime := time.Duration(0)
			if processed > 0 {
				avgTime = totalProcessingTime / time.Duration(processed)
			}

			log.Printf("📊 [Статистика] Обработано: %d, Активных: %d/%d, Среднее время: %v",
				processed, active, maxConcurrentHandlers, avgTime)
		}
	}()

	// Обрабатываем сообщения
	for update := range updates {
		// Получаем слот для обработки
		semaphore <- struct{}{}
		atomic.AddInt32(&activeHandlers, 1)
		currentActive := atomic.LoadInt32(&activeHandlers)

		log.Printf("🚦 [Семафор] Получен слот. Активных обработок: %d/%d", currentActive, maxConcurrentHandlers)

		go func(update tgbotapi.Update, handlerID int32) {
			startTime := time.Now()
			defer func() {
				<-semaphore // Освобождаем слот после обработки
				atomic.AddInt32(&activeHandlers, -1)
				duration := time.Since(startTime)

				// Обновляем статистику
				atomic.AddInt64(&totalProcessed, 1)
				// Примечание: totalProcessingTime нужно обновлять безопасно
				timeMutex.Lock()
				totalProcessingTime += duration
				timeMutex.Unlock()

				log.Printf("🚦 [Семафор] Обработчик %d завершен за %v. Активных обработок: %d/%d",
					handlerID, duration, atomic.LoadInt32(&activeHandlers), maxConcurrentHandlers)
			}()

			log.Printf("🚦 [Семафор] Обработчик %d начал работу", handlerID)

			// Добавляем информацию о пользователе
			if update.Message != nil && update.Message.From != nil {
				log.Printf("👤 [Семафор] Обработчик %d работает с пользователем %d (@%s)",
					handlerID, update.Message.From.ID, update.Message.From.UserName)
			} else if update.CallbackQuery != nil && update.CallbackQuery.From != nil {
				log.Printf("👤 [Семафор] Обработчик %d работает с пользователем %d (@%s)",
					handlerID, update.CallbackQuery.From.ID, update.CallbackQuery.From.UserName)
			}

			// Обрабатываем callback от инлайн-кнопок
			if update.CallbackQuery != nil {
				inlineHandler.HandleCallback(customBot, update.CallbackQuery)
				return
			}

			fmt.Println("Обработка сообщения")
			if update.Message == nil {
				return
			}

			fmt.Println("Обработка голосового сообщения")
			// Обрабатываем голосовые сообщения через MessageHandler
			if update.Message.Voice != nil {
				messageHandler.HandleMessage(customBot, update.Message)
				return
			}

			fmt.Println("Обработка текстового сообщения")
			// Сначала обрабатываем команды (они имеют приоритет)
			if update.Message.IsCommand() {
				// Если пользователь ввёл команду, сбрасываем специальные состояния
				userID := update.Message.From.ID
				state := stateManager.GetState(userID)
				if state.WaitingForEmail {
					state.WaitingForEmail = false
				}
				handleMessage(customBot, update.Message, voiceHandler, stateManager, inlineHandler)
				return
			}

			// Затем проверяем специальные состояния (email, etc) через MessageHandler
			if handled := messageHandler.HandleMessage(customBot, update.Message); handled {
				return // сообщение уже обработано, не продолжаем
			}
			// Обрабатываем обычные текстовые сообщения
			handleMessage(customBot, update.Message, voiceHandler, stateManager, inlineHandler)
		}(update, currentActive)
	}
}

// setupGracefulShutdown настраивает graceful shutdown для приложения
func setupGracefulShutdown(cancel context.CancelFunc) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("🛑 Получен сигнал остановки, завершаем работу...")
		cancel()
		time.Sleep(2 * time.Second) // Даем время воркерам завершиться
		os.Exit(0)
	}()
}

// handleMessage теперь не обрабатывает голосовые сообщения напрямую
func handleMessage(bot *bot.Bot, message *tgbotapi.Message, voiceHandler *voice.VoiceHandler, stateManager *bot.StateManager, inlineHandler *bot.InlineHandler) {
	// Логируем входящие сообщения
	log.Printf("[%s] %s", message.From.UserName, message.Text)
	fmt.Println("Обработка команды")
	// Обрабатываем команды
	if message.IsCommand() {
		handleCommand(bot, message)
		return
	}
	fmt.Println("Обработка текстового сообщения")
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
	case "admin":
		handleAdminCommand(bot, message)
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
	// Получаем информацию о подписке пользователя
	userID := chatID // В Telegram chatID обычно равен userID для личных чатов

	subscription, err := bot.SubscriptionService.GetUserSubscription(userID)
	if err != nil {
		log.Printf("Ошибка получения подписки для пользователя %d: %v", userID, err)
		subscription = nil
	}

	var text string
	var keyboard tgbotapi.InlineKeyboardMarkup

	if subscription != nil && subscription.Active {
		// У пользователя есть активная подписка
		statusText := "Активна"
		if subscription.Status == "cancelled" {
			statusText = "Отменена (работает до конца периода)"
		}

		nextPaymentText := "Не указана"
		if subscription.NextPayment != (time.Time{}) {
			nextPaymentText = subscription.NextPayment.Format("02.01.2006 15:04")
		}

		text = fmt.Sprintf(`💎 *Ваша подписка*

📊 Тариф: %s
✅ Статус: %s
⏰ Следующий платеж: %s
💰 Стоимость: %.0f₽/месяц

✨ Ваши возможности:
• Неограниченное количество сообщений
• Приоритетная обработка
• Расширенные возможности редактирования
• Доступ к эксклюзивным функциям`,
			subscription.Tariff, statusText, nextPaymentText, subscription.Amount)

		keyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("❌ Отменить подписку", "cancel_subscription"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🔙 Назад в меню", "main_menu"),
			),
		)
	} else {
		// У пользователя нет активной подписки
		text = `💎 *Подписка*

📊 Текущий тариф: Бесплатный
⏰ Срок действия: Бессрочно

✨ Премиум тариф:
• Неограниченное количество сообщений
• Приоритетная обработка
• Расширенные возможности редактирования
• Доступ к эксклюзивным функциям

💳 Стоимость: 990₽/месяц`

		keyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("💰 Купить подписку", "buy_premium"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🔙 Назад в меню", "main_menu"),
			),
		)
	}

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = &keyboard
	bot.Send(msg)
}

func sendUnknownCommandMessage(bot *bot.Bot, chatID int64) {
	text := "❌ Неизвестная команда. Используйте /start для начала работы."

	msg := tgbotapi.NewMessage(chatID, text)
	bot.Send(msg)
}

func handleAdminCommand(bot *bot.Bot, message *tgbotapi.Message) {
	// Проверяем права администратора
	isAdmin, err := bot.DB.IsAdmin(message.From.ID)
	if err != nil {
		log.Printf("Ошибка проверки прав администратора: %v", err)
		msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Произошла ошибка при проверке прав доступа")
		bot.Send(msg)
		return
	}

	if !isAdmin {
		msg := tgbotapi.NewMessage(message.Chat.ID, "⛔ У вас нет прав администратора")
		bot.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, "🛠 Админ-панель\n\nДоступные команды:\n/reset_limits [user_id] - Сбросить лимиты пользователя\n/add_admin [user_id] - Добавить администратора")
	bot.Send(msg)
}
