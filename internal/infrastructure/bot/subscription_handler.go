package bot

import (
	"ai_tg_writer/internal/service"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type SubscriptionHandler struct {
	subscriptionService *service.SubscriptionService
}

func NewSubscriptionHandler(subscriptionService *service.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{
		subscriptionService: subscriptionService,
	}
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
				tgbotapi.NewInlineKeyboardButtonData("❌ Отменить подписку", "cancel_subscription"),
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
			"Вы уверены, что хотите отменить подписку?\n"+
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

// CheckSubscriptionStatus проверяет статус подписки пользователя
func (h *SubscriptionHandler) CheckSubscriptionStatus(userID int64) (bool, error) {
	return h.subscriptionService.IsUserSubscribed(userID)
}
