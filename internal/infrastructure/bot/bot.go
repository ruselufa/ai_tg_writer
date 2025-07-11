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
			tgbotapi.NewInlineKeyboardButtonData("❓ Помощь", "help"),
			tgbotapi.NewInlineKeyboardButtonData("👤 Мой профиль", "profile"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💎 Моя подписка", "subscription"),
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
