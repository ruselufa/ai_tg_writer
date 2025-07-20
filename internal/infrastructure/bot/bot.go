package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Bot представляет собой обертку над tgbotapi.BotAPI с дополнительной функциональностью
type Bot struct {
	*tgbotapi.BotAPI
}

// NewBot создает новый экземпляр Bot
func NewBot(api *tgbotapi.BotAPI) *Bot {
	return &Bot{
		BotAPI: api,
	}
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
			tgbotapi.NewInlineKeyboardButtonData("🎨 Настройки стилизации", "styling_settings"),
			tgbotapi.NewInlineKeyboardButtonData("❓ Помощь", "help"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👤 Мой профиль", "profile"),
			tgbotapi.NewInlineKeyboardButtonData("💎 Моя подписка", "subscription"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🧪 Тест форматирования", "test_formatting"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(" ", "no_action"),
		),
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
func (b *Bot) SendFormattedMessage(chatID int64, text string, entities []MessageEntity) error {
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

	_, err := b.Send(msg)
	return err
}

// SendFormattedMessageWithKeyboard отправляет форматированное сообщение с клавиатурой
func (b *Bot) SendFormattedMessageWithKeyboard(chatID int64, text string, entities []MessageEntity, keyboard tgbotapi.InlineKeyboardMarkup) error {
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

	_, err := b.Send(msg)
	return err
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
