# Telegram Bot с поддержкой Entities

Пакет для работы с Telegram ботом, включающий автоматическое форматирование постов с использованием Telegram API entities.

## Возможности

- ✅ Автоматическое форматирование постов
- ✅ Настраиваемые стили для разных типов контента
- ✅ Поддержка всех типов Telegram entities
- ✅ Интерактивные настройки стилизации
- ✅ Валидация entities
- ✅ Интеграция с существующим ботом

## Структура

### Основные компоненты

- **Bot** - обертка над tgbotapi.BotAPI с дополнительными методами
- **InlineHandler** - обработчик inline-команд
- **MessageHandler** - обработчик сообщений
- **StateManager** - управление состоянием пользователей
- **TelegramPostFormatter** - форматирование постов с entities

### Типы данных

- **PostStyling** - настройки стилизации
- **MessageEntity** - Telegram entity
- **Post** - структура поста с entities
- **UserState** - состояние пользователя

## Использование

### Базовое форматирование поста

```go
// Создаем настройки стилизации
styling := bot.DefaultPostStyling()

// Создаем форматтер
formatter := bot.NewTelegramPostFormatter(styling)

// Форматируем пост
text := "Ваш текст поста"
cleanText, entities := formatter.FormatPost(text)

// Отправляем с форматированием
err := bot.SendFormattedMessage(chatID, cleanText, entities)
```

### Настройка стилизации

```go
// Создаем кастомные настройки
styling := bot.PostStyling{
    UseBold:         true,
    UseItalic:       true,
    UseHashtags:     true,
    UseLinks:        true,
    UseStrikethrough: false,
    UseCode:         false,
    UseMentions:     false,
    UseUnderline:    false,
    UsePre:          false,
}

// Применяем к пользователю
stateManager.SetPostStyling(userID, styling)
```

### Отправка форматированного сообщения с клавиатурой

```go
keyboard := bot.CreateApprovalKeyboard()
err := bot.SendFormattedMessageWithKeyboard(
    chatID,
    cleanText,
    entities,
    keyboard,
)
```

## Настройки стилизации

### Доступные опции

- **UseBold** - жирный текст для заголовков
- **UseItalic** - курсив для акцентов
- **UseStrikethrough** - зачеркивание
- **UseCode** - код в строке
- **UseLinks** - автоматическое создание ссылок
- **UseHashtags** - автоматическое добавление хештегов
- **UseMentions** - упоминания пользователей
- **UseUnderline** - подчеркивание
- **UsePre** - блоки кода

### Настройки по умолчанию

```go
func DefaultPostStyling() PostStyling {
    return PostStyling{
        UseBold:         true,
        UseItalic:       true,
        UseStrikethrough: false,
        UseCode:         false,
        UseLinks:        true,
        UseHashtags:     true,
        UseMentions:     false,
        UseUnderline:    false,
        UsePre:          false,
    }
}
```

## Интеграция с ботом

### Обработка команд

```go
// В HandleCallback добавляем новые команды
case "styling_settings":
    ih.handleStylingSettings(bot, callback)
case "toggle_bold":
    ih.handleToggleBold(bot, callback)
// ... другие команды
```

### Автоматическое форматирование

При создании поста бот автоматически:

1. Получает настройки стилизации пользователя
2. Форматирует текст с помощью `TelegramPostFormatter`
3. Создает entities для Telegram API
4. Отправляет сообщение с форматированием

### Сохранение постов

Посты сохраняются с entities и настройками стилизации:

```go
post := Post{
    ContentType: contentType,
    Content:     cleanText,
    Messages:    voiceMessages,
    Entities:    entities,
    Styling:     styling,
}
```

## Интерфейс пользователя

### Главное меню

Добавлена кнопка "🎨 Настройки стилизации" в главное меню.

### Настройки стилизации

Интерактивное меню с кнопками для включения/выключения различных типов форматирования:

- 🔤 Жирный текст
- 📝 Курсив
- ❌ Зачеркивание
- 💻 Код
- 🔗 Ссылки
- # Хештеги
- @ Упоминания
- 📋 Подчеркивание
- 📦 Блоки кода

## Примеры

### Пример 1: Создание поста с форматированием

```go
// Пользователь отправляет голосовое сообщение
// Бот транскрибирует и генерирует пост
postText := "Как освоить слепую печать быстро и правильно?"

// Автоматическое форматирование
formatter := NewTelegramPostFormatter(userState.PostStyling)
cleanText, entities := formatter.FormatPost(postText)

// Отправка с форматированием
bot.SendFormattedMessageWithKeyboard(chatID, cleanText, entities, keyboard)
```

### Пример 2: Изменение настроек стилизации

```go
// Пользователь нажимает "toggle_bold"
func (ih *InlineHandler) handleToggleBold(bot *Bot, callback *tgbotapi.CallbackQuery) {
    userID := callback.From.ID
    styling := ih.stateManager.GetPostStyling(userID)
    
    // Переключаем настройку
    styling.UseBold = !styling.UseBold
    ih.stateManager.SetPostStyling(userID, styling)
    
    // Показываем обновленные настройки
    ih.handleStylingSettings(bot, callback)
}
```

### Пример 3: Кастомные настройки для разных типов контента

```go
// Для YouTube сценариев - больше акцентов
youtubeStyling := PostStyling{
    UseBold:         true,
    UseItalic:       true,
    UseHashtags:     false, // Не нужны для YouTube
    UseLinks:        true,
    UseStrikethrough: true, // Для зачеркивания ошибок
}

// Для Instagram постов - больше хештегов
instagramStyling := PostStyling{
    UseBold:         true,
    UseItalic:       true,
    UseHashtags:     true, // Важно для Instagram
    UseLinks:        false,
    UseMentions:     true, // Упоминания брендов
}
```

## Обработка ошибок

### Валидация entities

```go
err := formatter.validateEntities(cleanText, entities)
if err != nil {
    // Отправляем без форматирования
    bot.Send(tgbotapi.NewMessage(chatID, cleanText))
    return
}
```

### Fallback при ошибках

Если отправка с entities не удалась, бот автоматически отправляет текст без форматирования:

```go
err := bot.SendFormattedMessageWithKeyboard(chatID, cleanText, entities, keyboard)
if err != nil {
    log.Printf("Ошибка отправки форматированного сообщения: %v", err)
    // Отправляем без форматирования
    resultMsg := tgbotapi.NewMessage(chatID, cleanText)
    resultMsg.ReplyMarkup = keyboard
    bot.Send(resultMsg)
}
```

## Производительность

- Форматирование происходит только при создании/редактировании поста
- Entities кэшируются в структуре Post
- Валидация выполняется только при необходимости
- Минимальные накладные расходы на форматирование

## Расширение функциональности

### Добавление новых типов entities

1. Добавить поле в `PostStyling`
2. Реализовать логику форматирования в `applyBasicFormatting`
3. Добавить паттерн в `parseMarkdownToEntities`
4. Добавить кнопку в `CreateStylingSettingsKeyboard`

### Кастомные стили для разных типов контента

```go
func getStylingForContentType(contentType string) PostStyling {
    switch contentType {
    case "telegram_post":
        return DefaultPostStyling()
    case "youtube_script":
        return getYouTubeStyling()
    case "instagram_post":
        return getInstagramStyling()
    default:
        return DefaultPostStyling()
    }
}
```

## Тестирование

Запуск примера:

```bash
cd examples
go run bot_entities_example.go
```

## Ссылки

- [Telegram Bot API Documentation](https://core.telegram.org/bots/api)
- [MessageEntity Documentation](https://core.telegram.org/type/MessageEntity)
- [Entities API Documentation](https://core.telegram.org/api/entities) 