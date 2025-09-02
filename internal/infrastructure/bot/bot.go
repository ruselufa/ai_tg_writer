package bot

import (
	"ai_tg_writer/internal/infrastructure/database"
	"ai_tg_writer/internal/service"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Bot представляет собой обертку над tgbotapi.BotAPI с дополнительной функциональностью
type Bot struct {
	API                 *tgbotapi.BotAPI
	StateManager        *StateManager
	DB                  *database.DB
	SubscriptionService *service.SubscriptionService
}

func NewBot(api *tgbotapi.BotAPI, db *database.DB) *Bot {
	return &Bot{
		API: api,
		DB:  db,
	}
}

func NewBotWithSubscriptionService(api *tgbotapi.BotAPI, db *database.DB, subscriptionService *service.SubscriptionService) *Bot {
	return &Bot{
		API:                 api,
		DB:                  db,
		SubscriptionService: subscriptionService,
	}
}

// Send отправляет сообщение через API бота
func (b *Bot) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	return b.API.Send(c)
}

// CreateApprovalKeyboard создает клавиатуру для согласования результата
func (b *Bot) CreateApprovalKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Согласен", "approve"),
			tgbotapi.NewInlineKeyboardButtonData("✏️ Внести правки", "edit_post"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏠 Главное меню", "main_menu"),
		),
	)
}

// CreateEditApprovalKeyboard создает клавиатуру для согласования после правок
func (b *Bot) CreateEditApprovalKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✏️ Внести изменения", "edit_post"),
			tgbotapi.NewInlineKeyboardButtonData("💾 Сохранить ответ", "save_post"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏠 Главное меню", "main_menu"),
		),
	)
}

// CreateContinueKeyboard создает клавиатуру для продолжения диктовки
func (b *Bot) CreateContinueKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✍️ Продолжить диктовку", "continue_dictation"),
			tgbotapi.NewInlineKeyboardButtonData("🚀 Начать создание", "start_creation"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏠 Главное меню", "main_menu"),
		),
	)
}

// CreateEditContinueKeyboard создает клавиатуру для продолжения правок
func (b *Bot) CreateEditContinueKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✍️ Добавить правки", "continue_dictation"),
			tgbotapi.NewInlineKeyboardButtonData("✅ Применить изменения", "edit_start_creation"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏠 Главное меню", "main_menu"),
		),
	)
}

// CreateMainKeyboard создает главное меню с пробелом-заглушкой
func (b *Bot) CreateMainKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📝 Создать пост/сценарий", "create_post"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👤 Мой профиль", "profile"),
			tgbotapi.NewInlineKeyboardButtonData("💎 Моя подписка", "subscription"),
		),
		tgbotapi.NewInlineKeyboardRow(
			// tgbotapi.NewInlineKeyboardButtonData("🎨 Настройки стилизации", "styling_settings"),
			tgbotapi.NewInlineKeyboardButtonData("❓ Помощь", "help"),
		),
		// tgbotapi.NewInlineKeyboardRow(
		// 	tgbotapi.NewInlineKeyboardButtonData("🧪 Тест форматирования", "test_formatting"),
		// ),
	)
}

// CreateContentTypeKeyboard создает клавиатуру выбора типа контента
func (b *Bot) CreateContentTypeKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Создать пост в Телеграм", "create_telegram_post"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Создать сценарий для видео на YouTube", "create_script_youtube"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Создать сценарий для Reels в Instagram", "create_script_reels"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Создать пост в Instagram", "create_post_instagram"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏠 Главное меню", "main_menu"),
		),
	)
}

// CreatePostActionKeyboard создает клавиатуру с действиями для поста
func (b *Bot) CreatePostActionKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✏️ Редактировать", "edit_post"),
			tgbotapi.NewInlineKeyboardButtonData("✅ Сохранить", "save_post"),
		),
	)
}

// SendFormattedMessage отправляет сообщение с форматированием
// Автоматически разбивает длинные сообщения на части
func (b *Bot) SendFormattedMessage(chatID int64, text string, entities []MessageEntity) (int, error) {
	// Проверяем длину текста и разбиваем на части если нужно
	if len(text) > 3900 {
		return b.sendSplitFormattedMessage(chatID, text, entities)
	}

	msg := tgbotapi.NewMessage(chatID, text)

	// Конвертируем наши entities в формат tgbotapi
	var tgbotEntities []tgbotapi.MessageEntity
	for _, entity := range entities {
		tgbotEntity := tgbotapi.MessageEntity{
			Type:   entity.Type,
			Offset: entity.Offset,
			Length: entity.Length,
		}

		if entity.URL != "" {
			tgbotEntity.URL = entity.URL
		}

		if entity.User != nil {
			tgbotEntity.User = &tgbotapi.User{
				ID:           entity.User.ID,
				IsBot:        entity.User.IsBot,
				FirstName:    entity.User.FirstName,
				LastName:     entity.User.LastName,
				UserName:     entity.User.Username,
				LanguageCode: entity.User.LanguageCode,
			}
		}

		if entity.Language != "" {
			tgbotEntity.Language = entity.Language
		}

		tgbotEntities = append(tgbotEntities, tgbotEntity)
	}

	msg.Entities = tgbotEntities

	message, err := b.Send(msg)
	if err != nil {
		// Если ошибка связана с длиной сообщения, пробуем разбить на части
		if err.Error() == "Bad Request: message is too long" {
			return b.sendSplitFormattedMessage(chatID, text, entities)
		}
		return 0, err
	}
	return message.MessageID, nil
}

// SendFormattedMessageWithKeyboard отправляет форматированное сообщение с клавиатурой
func (b *Bot) SendFormattedMessageWithKeyboard(chatID int64, text string, entities []MessageEntity, keyboard tgbotapi.InlineKeyboardMarkup) (int, error) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard

	// Конвертируем наши entities в формат tgbotapi
	var tgbotEntities []tgbotapi.MessageEntity
	for _, entity := range entities {
		tgbotEntity := tgbotapi.MessageEntity{
			Type:   entity.Type,
			Offset: entity.Offset,
			Length: entity.Length,
		}

		if entity.URL != "" {
			tgbotEntity.URL = entity.URL
		}

		if entity.User != nil {
			tgbotEntity.User = &tgbotapi.User{
				ID:           entity.User.ID,
				IsBot:        entity.User.IsBot,
				FirstName:    entity.User.FirstName,
				LastName:     entity.User.LastName,
				UserName:     entity.User.Username,
				LanguageCode: entity.User.LanguageCode,
			}
		}

		if entity.Language != "" {
			tgbotEntity.Language = entity.Language
		}

		tgbotEntities = append(tgbotEntities, tgbotEntity)
	}

	msg.Entities = tgbotEntities

	message, err := b.Send(msg)
	if err != nil {
		return 0, err
	}
	return message.MessageID, nil
}

// CreateStylingSettingsKeyboard создает клавиатуру для настроек стилизации
func (b *Bot) CreateStylingSettingsKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔤 Жирный текст", "toggle_bold"),
			tgbotapi.NewInlineKeyboardButtonData("📝 Курсив", "toggle_italic"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ Зачеркивание", "toggle_strikethrough"),
			tgbotapi.NewInlineKeyboardButtonData("💻 Код", "toggle_code"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔗 Ссылки", "toggle_links"),
			tgbotapi.NewInlineKeyboardButtonData("# Хештеги", "toggle_hashtags"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("@ Упоминания", "toggle_mentions"),
			tgbotapi.NewInlineKeyboardButtonData("📋 Подчеркивание", "toggle_underline"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📦 Блоки кода", "toggle_pre"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏠 Главное меню", "main_menu"),
		),
	)
}

// sendSplitFormattedMessage разбивает длинное сообщение на части и отправляет их
func (b *Bot) sendSplitFormattedMessage(chatID int64, text string, entities []MessageEntity) (int, error) {
	// Разбиваем текст на части по 3800 символов (меньше лимита Telegram в 4096)
	parts := splitText(text, 3800)
	var lastMessageID int

	for i, part := range parts {
		// Создаем сообщение для части текста
		msg := tgbotapi.NewMessage(chatID, part)

		// Отправляем без форматирования для разбитых сообщений
		message, err := b.Send(msg)
		if err != nil {
			return 0, err
		}

		// Сохраняем ID последнего сообщения
		lastMessageID = message.MessageID

		// Если это не последняя часть, добавляем индикатор продолжения
		if i < len(parts)-1 {
			continueMsg := tgbotapi.NewMessage(chatID, "⏩ Продолжение следует...")
			_, err := b.Send(continueMsg)
			if err != nil {
				return 0, err
			}
		}
	}

	return lastMessageID, nil
}

// splitText разбивает текст на части указанного размера
func splitText(text string, maxLength int) []string {
	var parts []string
	runes := []rune(text)

	for len(runes) > 0 {
		if len(runes) <= maxLength {
			parts = append(parts, string(runes))
			break
		}

		// Ищем место для разбивки (предпочтительно по абзацам)
		splitIndex := findSplitIndex(runes, maxLength)
		parts = append(parts, string(runes[:splitIndex]))
		runes = runes[splitIndex:]
	}

	return parts
}

// findSplitIndex находит оптимальное место для разбивки текста
func findSplitIndex(runes []rune, maxLength int) int {
	if len(runes) <= maxLength {
		return len(runes)
	}

	// Пытаемся найти конец абзаца
	for i := maxLength; i > maxLength-100 && i > 0; i-- {
		if runes[i] == '\n' && (i+1 >= len(runes) || runes[i+1] == '\n') {
			return i + 1
		}
	}

	// Пытаемся найти конец предложения
	for i := maxLength; i > maxLength-50 && i > 0; i-- {
		if runes[i] == '.' || runes[i] == '!' || runes[i] == '?' {
			if i+1 < len(runes) && runes[i+1] == ' ' {
				return i + 2
			}
			return i + 1
		}
	}

	// Если не нашли хорошее место, разбиваем по максимальной длине
	return maxLength
}

// SendHTMLMessage отправляет сообщение с HTML разметкой, конвертируя ее в Telegram Entities
func (b *Bot) SendHTMLMessage(chatID int64, htmlText string) (int, error) {
	// Создаем форматтер с настройками по умолчанию
	formatter := NewTelegramPostFormatter(DefaultPostStyling())
	cleanText, entities := formatter.ParseHTMLToEntities(htmlText)
	return b.SendFormattedMessage(chatID, cleanText, entities)
}

// CreateSubscriptionLink создает ссылку на оплату подписки
func (b *Bot) CreateSubscriptionLink(userID int64, tariff string, amount float64) (string, error) {
	if b.SubscriptionService == nil {
		return "", fmt.Errorf("subscription service not initialized")
	}

	return b.SubscriptionService.CreateSubscriptionLink(userID, tariff, amount)
}
