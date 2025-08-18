package bot

import (
	"ai_tg_writer/internal/service"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type SubscriptionHandler struct {
	subscriptionService *service.SubscriptionService
	bot                 *Bot // Экземпляр бота для отправки сообщений
}

func NewSubscriptionHandler(subscriptionService *service.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{
		subscriptionService: subscriptionService,
		bot:                 nil, // Будет установлен позже
	}
}

// SetBot устанавливает экземпляр бота для отправки сообщений
func (h *SubscriptionHandler) SetBot(bot *Bot) {
	h.bot = bot
}

// HandleSubscriptionCommand обрабатывает команду /subscription
func (h *SubscriptionHandler) HandleSubscriptionCommand(bot *Bot, message *tgbotapi.Message) {
	userID := message.From.ID

	// Получаем доступные тарифы
	tariffs := h.subscriptionService.GetAvailableTariffs()

	if len(tariffs) == 0 {
		msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Тарифы временно недоступны")
		bot.Send(msg)
		return
	}

	tariff := tariffs[0] // Берем первый тариф

	// Проверяем текущую подписку пользователя
	subscription, err := h.subscriptionService.GetUserSubscription(userID)
	if err != nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка при получении информации о подписке")
		bot.Send(msg)
		return
	}

	var messageText string

	if subscription != nil && subscription.Active {
		// У пользователя есть активная подписка
		messageText = fmt.Sprintf(
			"🎉 *У вас активна подписка %s*\n\n"+
				"💰 Стоимость: %.0f₽/месяц\n"+
				"📅 Следующий платеж: %s\n"+
				"✅ Статус: Активна\n\n"+
				"Хотите отменить подписку?",
			tariff.Name,
			tariff.Price,
			subscription.NextPayment.Format("02.01.2006"),
		)

		// Создаем кнопки для управления подпиской
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("❌ Отменить подписку и отвязать карту", "cancel_subscription"),
			),
		)

		msg := tgbotapi.NewMessage(message.Chat.ID, messageText)
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = keyboard
		bot.Send(msg)

	} else {
		// У пользователя нет подписки
		messageText = fmt.Sprintf(
			"💎 *Подписка %s*\n\n"+
				"💰 Стоимость: %.0f₽/месяц\n"+
				"📝 Описание: %s\n\n"+
				"✨ Преимущества:\n",
			tariff.Name,
			tariff.Price,
			tariff.Description,
		)

		for _, feature := range tariff.Features {
			messageText += fmt.Sprintf("• %s\n", feature)
		}

		messageText += "\nНажмите кнопку ниже для оформления подписки:"

		// Создаем ссылку на оплату
		paymentURL, err := h.subscriptionService.CreateSubscriptionLink(userID, tariff.ID, tariff.Price)
		if err != nil {
			msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Ошибка создания ссылки на оплату")
			bot.Send(msg)
			return
		}

		// Создаем кнопку для оплаты
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonURL("💳 Оформить подписку", paymentURL),
			),
		)

		msg := tgbotapi.NewMessage(message.Chat.ID, messageText)
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = keyboard
		bot.Send(msg)
	}
}

// HandleSubscriptionCallback обрабатывает callback от кнопок подписки
func (h *SubscriptionHandler) HandleSubscriptionCallback(bot *Bot, callback *tgbotapi.CallbackQuery) {
	switch callback.Data {
	case "cancel_subscription":
		h.handleCancelSubscription(bot, callback)
	case "confirm_cancel_subscription":
		h.handleConfirmCancelSubscription(bot, callback)
	case "retry_payment":
		h.handleRetryPayment(bot, callback)
	case "change_payment_method":
		h.handleChangePaymentMethod(bot, callback)
	default:
		// Неизвестный callback - игнорируем
		return
	}
}

// handleCancelSubscription обрабатывает отмену подписки
func (h *SubscriptionHandler) handleCancelSubscription(bot *Bot, callback *tgbotapi.CallbackQuery) {
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
		"⚠️ *Подтверждение отмены подписки*\n\n"+
			"Вы уверены, что хотите отменить подписку и отвязать карту?\n"+
			"После отмены вы потеряете доступ к премиум функциям.",
	)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = &keyboard

	bot.Send(msg)
}

// handleConfirmCancelSubscription подтверждает отмену подписки
func (h *SubscriptionHandler) handleConfirmCancelSubscription(bot *Bot, callback *tgbotapi.CallbackQuery) {
	userID := callback.From.ID

	// Отменяем подписку
	err := h.subscriptionService.CancelSubscription(userID)
	if err != nil {
		msg := tgbotapi.NewEditMessageText(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
			"❌ Ошибка при отмене подписки. Попробуйте позже.",
		)
		bot.Send(msg)
		return
	}

	// Удаляем клавиатуру
	msg := tgbotapi.NewEditMessageText(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		"✅ *Подписка отменена*\n\n"+
			"Ваша подписка была успешно отменена.\n"+
			"Для возобновления доступа оформите подписку заново.",
	)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = &tgbotapi.InlineKeyboardMarkup{}

	bot.Send(msg)
}

// handleRetryPayment обрабатывает повторную попытку списания
func (h *SubscriptionHandler) handleRetryPayment(bot *Bot, callback *tgbotapi.CallbackQuery) {
	userID := callback.From.ID

	// Пытаемся списать деньги снова
	err := h.subscriptionService.RetryPayment(userID)
	if err != nil {
		msg := tgbotapi.NewEditMessageText(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
			"❌ *Ошибка при повторной попытке*\n\n"+
				"Не удалось списать деньги: "+err.Error()+"\n\n"+
				"Попробуйте позже или используйте другую карту.",
		)
		msg.ParseMode = "Markdown"
		bot.Send(msg)
		return
	}

	// Успешное списание
	msg := tgbotapi.NewEditMessageText(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		"✅ *Платеж успешен!*\n\n"+
			"Деньги успешно списаны с вашей карты.\n"+
			"Подписка продлена.",
	)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = &tgbotapi.InlineKeyboardMarkup{}

	bot.Send(msg)
}

// handleChangePaymentMethod обрабатывает смену метода оплаты
func (h *SubscriptionHandler) handleChangePaymentMethod(bot *Bot, callback *tgbotapi.CallbackQuery) {
	userID := callback.From.ID

	// Создаем новую ссылку для оплаты
	paymentURL, err := h.subscriptionService.ChangePaymentMethod(userID)
	if err != nil {
		msg := tgbotapi.NewEditMessageText(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
			"❌ *Ошибка создания ссылки*\n\n"+
				"Не удалось создать ссылку для оплаты: "+err.Error()+"\n\n"+
				"Попробуйте позже.",
		)
		msg.ParseMode = "Markdown"
		bot.Send(msg)
		return
	}

	// Создаем кнопку для оплаты новой картой
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("💳 Оплатить новой картой", paymentURL),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ Отменить подписку и отвязать карту", "cancel_subscription"),
		),
	)

	msg := tgbotapi.NewEditMessageText(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		"💳 *Смена метода оплаты*\n\n"+
			"Нажмите кнопку ниже для оплаты новой картой.\n\n"+
			"После успешной оплаты подписка будет восстановлена.",
	)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = &keyboard

	bot.Send(msg)
}

// CheckSubscriptionStatus проверяет статус подписки пользователя
func (h *SubscriptionHandler) CheckSubscriptionStatus(userID int64) (bool, error) {
	return h.subscriptionService.IsUserSubscribed(userID)
}

// SendPaymentFailedMessage отправляет сообщение о неудачной попытке оплаты
func (h *SubscriptionHandler) SendPaymentFailedMessage(userID int64, attempt int) error {
	if h.bot == nil {
		log.Printf("📨 [BOT] Cannot send message - bot not set for user %d (attempt %d)", userID, attempt)
		return fmt.Errorf("bot not set")
	}

	messageText := fmt.Sprintf(
		"❌ *Не удалось списать деньги*\n\n"+
			"Попытка %d из 3\n\n"+
			"Возможные причины:\n"+
			"• Недостаточно средств на карте\n"+
			"• Карта заблокирована\n"+
			"• Истек срок действия карты\n\n"+
			"Выберите действие:",
		attempt,
	)

	// Создаем кнопки для управления подпиской
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 Попробовать снова", "retry_payment"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💳 Использовать новую карту", "change_payment_method"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ Отменить подписку и отвязать карту", "cancel_subscription"),
		),
	)

	msg := tgbotapi.NewMessage(userID, messageText)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard

	_, err := h.bot.Send(msg)
	if err != nil {
		log.Printf("❌ [BOT] Failed to send payment failed message to user %d: %v", userID, err)
		return err
	}

	log.Printf("📨 [BOT] Payment failed message sent to user %d (attempt %d)", userID, attempt)
	return nil
}

// SendSubscriptionSuspendedMessage отправляет сообщение о приостановке подписки
func (h *SubscriptionHandler) SendSubscriptionSuspendedMessage(userID int64) error {
	if h.bot == nil {
		log.Printf("📨 [BOT] Cannot send message - bot not set for user %d", userID)
		return fmt.Errorf("bot not set")
	}

	messageText := "🚫 *Подписка приостановлена*\n\n" +
		"После 3 неудачных попыток списания ваша подписка была приостановлена.\n\n" +
		"Для восстановления доступа:\n" +
		"• Пополните баланс карты\n" +
		"• Используйте другую карту\n" +
		"• Обратитесь в поддержку\n\n" +
		"Выберите действие:"

	// Создаем кнопки для восстановления подписки
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 Попробовать снова", "retry_payment"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💳 Использовать новый способ оплаты", "change_payment_method"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ Отменить подписку и отвязать карту", "cancel_subscription"),
		),
	)

	msg := tgbotapi.NewMessage(userID, messageText)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard

	_, err := h.bot.Send(msg)
	if err != nil {
		log.Printf("❌ [BOT] Failed to send subscription suspended message to user %d: %v", userID, err)
		return err
	}

	log.Printf("📨 [BOT] Subscription suspended message sent to user %d", userID)
	return nil
}
