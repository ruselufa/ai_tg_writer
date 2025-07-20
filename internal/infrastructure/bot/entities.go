package bot

import (
	"fmt"
	"regexp"
	"strings"
)

// MessageEntity представляет entity в сообщении Telegram
type MessageEntity struct {
	Type     string `json:"type"`
	Offset   int    `json:"offset"`
	Length   int    `json:"length"`
	URL      string `json:"url,omitempty"`
	User     *User  `json:"user,omitempty"`
	Language string `json:"language,omitempty"`
}

// User представляет пользователя в entity
type User struct {
	ID           int64  `json:"id"`
	IsBot        bool   `json:"is_bot"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name,omitempty"`
	Username     string `json:"username,omitempty"`
	LanguageCode string `json:"language_code,omitempty"`
}

// PostStyling содержит настройки стилизации для постов
type PostStyling struct {
	UseBold          bool `json:"use_bold"`          // Использовать жирный текст для заголовков
	UseItalic        bool `json:"use_italic"`        // Использовать курсив для акцентов
	UseStrikethrough bool `json:"use_strikethrough"` // Использовать зачеркивание
	UseCode          bool `json:"use_code"`          // Использовать код
	UseLinks         bool `json:"use_links"`         // Использовать ссылки
	UseHashtags      bool `json:"use_hashtags"`      // Использовать хештеги
	UseMentions      bool `json:"use_mentions"`      // Использовать упоминания
	UseUnderline     bool `json:"use_underline"`     // Использовать подчеркивание
	UsePre           bool `json:"use_pre"`           // Использовать pre для блоков кода
}

// DefaultPostStyling возвращает настройки стилизации по умолчанию
func DefaultPostStyling() PostStyling {
	return PostStyling{
		UseBold:          true,
		UseItalic:        true,
		UseStrikethrough: false,
		UseCode:          false,
		UseLinks:         true,
		UseHashtags:      true,
		UseMentions:      false,
		UseUnderline:     false,
		UsePre:           false,
	}
}

// TelegramPostFormatter форматирует текст для Telegram с entities
type TelegramPostFormatter struct {
	styling PostStyling
}

// NewTelegramPostFormatter создает новый форматтер
func NewTelegramPostFormatter(styling PostStyling) *TelegramPostFormatter {
	return &TelegramPostFormatter{
		styling: styling,
	}
}

// FormatPost форматирует пост с учетом настроек стилизации
func (tf *TelegramPostFormatter) FormatPost(text string) (string, []MessageEntity) {
	// Очищаем MarkdownV2 разметку перед парсингом
	cleanMarkdownText := tf.cleanMarkdownV2(text)

	// Парсим entities из очищенного текста
	cleanText, entities := tf.ParseMarkdownToEntities(cleanMarkdownText)

	// Отладочная информация
	fmt.Printf("Форматирование поста:\n")
	fmt.Printf("Исходный текст: %s\n", text[:min(100, len(text))])
	fmt.Printf("После очистки MarkdownV2: %s\n", cleanMarkdownText[:min(100, len(cleanMarkdownText))])
	fmt.Printf("Очищенный текст: %s\n", cleanText[:min(100, len(cleanText))])
	fmt.Printf("Найдено entities: %d\n", len(entities))

	// Валидируем entities
	if err := tf.validateEntities(cleanText, entities); err != nil {
		fmt.Printf("Ошибка валидации entities: %v\n", err)
		// Если есть ошибки, возвращаем исходный текст без entities
		return text, []MessageEntity{}
	}

	return cleanText, entities
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// applyBasicFormatting применяет базовое форматирование к тексту
func (tf *TelegramPostFormatter) applyBasicFormatting(text string) string {
	result := text

	// Форматируем заголовки (строки, начинающиеся с #)
	if tf.styling.UseBold {
		result = tf.formatHeadings(result)
	}

	// Форматируем важные слова (в кавычках или с восклицательными знаками)
	if tf.styling.UseItalic {
		result = tf.formatEmphasis(result)
	}

	// Форматируем хештеги
	if tf.styling.UseHashtags {
		result = tf.formatHashtags(result)
	}

	// Форматируем упоминания
	if tf.styling.UseMentions {
		result = tf.formatMentions(result)
	}

	// Форматируем ссылки
	if tf.styling.UseLinks {
		result = tf.formatLinks(result)
	}

	return result
}

// formatHeadings форматирует заголовки
func (tf *TelegramPostFormatter) formatHeadings(text string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			// Убираем # и делаем текст жирным
			content := strings.TrimSpace(strings.TrimPrefix(trimmed, "#"))
			if content != "" {
				lines[i] = strings.Replace(line, trimmed, "*"+content+"*", 1)
			}
		}
	}
	return strings.Join(lines, "\n")
}

// formatEmphasis форматирует акценты
func (tf *TelegramPostFormatter) formatEmphasis(text string) string {
	// Находим слова в кавычках и делаем их курсивными
	quoteRegex := regexp.MustCompile(`"([^"]+)"`)
	result := quoteRegex.ReplaceAllString(text, "_$1_")

	// Находим слова с восклицательными знаками
	emphasisRegex := regexp.MustCompile(`(\w+!)`)
	result = emphasisRegex.ReplaceAllString(result, "_$1_")

	return result
}

// formatHashtags форматирует хештеги
func (tf *TelegramPostFormatter) formatHashtags(text string) string {
	// Находим хештеги и добавляем # если его нет
	hashtagRegex := regexp.MustCompile(`\b([А-Яа-яA-Za-z]+[А-Яа-яA-Za-z0-9]*)\b`)
	result := hashtagRegex.ReplaceAllStringFunc(text, func(match string) string {
		// Проверяем, не является ли это уже хештегом
		if strings.HasPrefix(match, "#") {
			return match
		}

		// Проверяем, является ли это ключевым словом для хештега
		keywords := []string{"обучение", "продуктивность", "советы", "тренировка", "развитие", "навыки", "техники", "методы"}
		for _, keyword := range keywords {
			if strings.ToLower(match) == keyword {
				return "#" + match
			}
		}
		return match
	})

	return result
}

// formatMentions форматирует упоминания
func (tf *TelegramPostFormatter) formatMentions(text string) string {
	// Находим упоминания и добавляем @ если его нет
	mentionRegex := regexp.MustCompile(`\b([А-Яа-яA-Za-z]+[А-Яа-яA-Za-z0-9_]*)\b`)
	result := mentionRegex.ReplaceAllStringFunc(text, func(match string) string {
		// Проверяем, не является ли это уже упоминанием
		if strings.HasPrefix(match, "@") {
			return match
		}

		// Проверяем, является ли это именем пользователя
		if len(match) > 3 && len(match) < 20 {
			return "@" + match
		}
		return match
	})

	return result
}

// formatLinks форматирует ссылки
func (tf *TelegramPostFormatter) formatLinks(text string) string {
	// Находим URL и оборачиваем их в markdown ссылки
	urlRegex := regexp.MustCompile(`(https?://[^\s]+)`)
	result := urlRegex.ReplaceAllString(text, "[Ссылка]($1)")

	return result
}

// ParseMarkdownToEntities парсит markdown в Telegram entities
func (tf *TelegramPostFormatter) ParseMarkdownToEntities(text string) (string, []MessageEntity) {
	fmt.Printf("=== ParseMarkdownToEntities вызвана ===\n")
	fmt.Printf("Входной текст: %s\n", text[:min(100, len(text))])

	var entities []MessageEntity
	cleanText := text

	// Паттерны для различных типов разметки
	patterns := []struct {
		regex      *regexp.Regexp
		entityType string
		groupIndex int
	}{
		{regexp.MustCompile(`\*\*([^*]+)\*\*`), "bold", 1},                         // **жирный**
		{regexp.MustCompile(`\*([^*]+)\*`), "bold", 1},                             // *жирный*
		{regexp.MustCompile(`_([^_]+)_`), "italic", 1},                             // _курсив_
		{regexp.MustCompile(`~([^~]+)~`), "strikethrough", 1},                      // ~зачеркнутый~
		{regexp.MustCompile(`` + "`" + `([^` + "`" + `]+)` + "`" + ``), "code", 1}, // `код`
		{regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`), "text_link", 1},            // [текст](ссылка)
		{regexp.MustCompile(`@(\w+)`), "mention", 1},                               // @username
		{regexp.MustCompile(`#([А-Яа-яA-Za-z0-9_]+)`), "hashtag", 1},               // #hashtag (поддержка русских букв)
	}

	// Собираем все совпадения с их позициями
	type matchInfo struct {
		start       int
		end         int
		entityType  string
		url         string
		markupStart int
		markupEnd   int
	}

	var allMatches []matchInfo

	for _, pattern := range patterns {
		matches := pattern.regex.FindAllStringSubmatchIndex(text, -1)

		for _, match := range matches {
			start := match[pattern.groupIndex*2]
			end := match[pattern.groupIndex*2+1]
			markupStart := match[0]
			markupEnd := match[1]

			matchInfo := matchInfo{
				start:       start,
				end:         end,
				entityType:  pattern.entityType,
				markupStart: markupStart,
				markupEnd:   markupEnd,
			}

			// Для ссылок добавляем URL
			if pattern.entityType == "text_link" {
				urlStart := match[4]
				urlEnd := match[5]
				matchInfo.url = text[urlStart:urlEnd]
			}

			allMatches = append(allMatches, matchInfo)
		}
	}

	// Сортируем совпадения по позиции начала разметки (в прямом порядке)
	for i := 0; i < len(allMatches)-1; i++ {
		for j := 0; j < len(allMatches)-i-1; j++ {
			if allMatches[j].markupStart > allMatches[j+1].markupStart {
				allMatches[j], allMatches[j+1] = allMatches[j+1], allMatches[j]
			}
		}
	}

	// Создаем новый текст без разметки
	var newText strings.Builder
	lastPos := 0

	// Обрабатываем совпадения в прямом порядке
	for _, match := range allMatches {
		// Добавляем текст до разметки
		if match.markupStart > lastPos {
			newText.WriteString(text[lastPos:match.markupStart])
		}

		// Добавляем содержимое разметки (без самой разметки)
		newText.WriteString(text[match.start:match.end])

		// Создаем entity с правильными UTF-16 смещениями
		textBeforeEntity := newText.String()
		entityStart := tf.getUTF16Length(textBeforeEntity) - tf.getUTF16Length(text[match.start:match.end])
		entityLength := tf.getUTF16Length(text[match.start:match.end])

		entity := MessageEntity{
			Type:   match.entityType,
			Offset: entityStart,
			Length: entityLength,
		}

		if match.url != "" {
			entity.URL = match.url
		}

		entities = append(entities, entity)

		lastPos = match.markupEnd
	}

	// Добавляем оставшийся текст
	if lastPos < len(text) {
		newText.WriteString(text[lastPos:])
	}

	cleanText = newText.String()

	// Отладочная информация
	fmt.Printf("Обработано совпадений: %d\n", len(allMatches))
	fmt.Printf("Длина исходного текста: %d байт, %d UTF-16\n", len(text), tf.getUTF16Length(text))
	fmt.Printf("Длина очищенного текста: %d байт, %d UTF-16\n", len(cleanText), tf.getUTF16Length(cleanText))

	// Показываем все найденные совпадения
	for i, match := range allMatches {
		fmt.Printf("Совпадение %d: Type=%s, Start=%d, End=%d, MarkupStart=%d, MarkupEnd=%d\n",
			i+1, match.entityType, match.start, match.end, match.markupStart, match.markupEnd)
		fmt.Printf("  Исходный текст: '%s'\n", text[match.start:match.end])
	}

	// Показываем созданные entities
	for i, entity := range entities {
		fmt.Printf("Entity %d: Type=%s, Offset=%d, Length=%d\n", i+1, entity.Type, entity.Offset, entity.Length)
		// Находим текст entity в очищенном тексте
		entityText := tf.getTextAtUTF16Position(cleanText, entity.Offset, entity.Length)
		fmt.Printf("  Текст entity: '%s'\n", entityText)
	}

	// Сортируем entities по offset для правильного порядка
	tf.sortEntities(entities)

	return cleanText, entities
}

// validateEntities проверяет корректность entities
func (tf *TelegramPostFormatter) validateEntities(text string, entities []MessageEntity) error {
	textUTF16Length := tf.getUTF16Length(text)

	for _, entity := range entities {
		if entity.Offset < 0 || entity.Offset >= textUTF16Length {
			return fmt.Errorf("некорректный offset для entity: %d (максимум: %d)", entity.Offset, textUTF16Length)
		}

		if entity.Length <= 0 || entity.Offset+entity.Length > textUTF16Length {
			return fmt.Errorf("некорректная длина для entity: %d (максимум: %d)", entity.Length, textUTF16Length-entity.Offset)
		}
	}

	return nil
}

// getUTF16Length вычисляет длину строки в UTF-16 кодовых единицах
func (tf *TelegramPostFormatter) getUTF16Length(text string) int {
	length := 0
	runes := []rune(text)

	for _, r := range runes {
		if r >= 0x10000 {
			length += 2 // Суррогатные пары UTF-16
		} else {
			length += 1 // Обычные символы
		}
	}
	return length
}

// getTextAtUTF16Position получает текст по UTF-16 позиции и длине
func (tf *TelegramPostFormatter) getTextAtUTF16Position(text string, offset, length int) string {
	runes := []rune(text)
	utf16Pos := 0
	runePos := 0

	// Находим начальную позицию в рунах
	for runePos < len(runes) && utf16Pos < offset {
		r := runes[runePos]
		if r >= 0x10000 {
			utf16Pos += 2
		} else {
			utf16Pos += 1
		}
		runePos++
	}

	startRunePos := runePos

	// Находим конечную позицию в рунах
	for runePos < len(runes) && utf16Pos < offset+length {
		r := runes[runePos]
		if r >= 0x10000 {
			utf16Pos += 2
		} else {
			utf16Pos += 1
		}
		runePos++
	}

	if startRunePos < len(runes) && runePos <= len(runes) {
		return string(runes[startRunePos:runePos])
	}

	return ""
}

// sortEntities сортирует entities по offset
func (tf *TelegramPostFormatter) sortEntities(entities []MessageEntity) {
	// Простая сортировка пузырьком для небольшого количества entities
	for i := 0; i < len(entities)-1; i++ {
		for j := 0; j < len(entities)-i-1; j++ {
			if entities[j].Offset > entities[j+1].Offset {
				entities[j], entities[j+1] = entities[j+1], entities[j]
			}
		}
	}
}

// EscapeMarkdownV2 экранирует специальные символы для MarkdownV2
func (tf *TelegramPostFormatter) EscapeMarkdownV2(text string) string {
	specialChars := []string{"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}
	result := text

	for _, char := range specialChars {
		result = strings.ReplaceAll(result, char, "\\"+char)
	}

	return result
}

// UnescapeMarkdownV2 убирает экранирование из MarkdownV2
func (tf *TelegramPostFormatter) UnescapeMarkdownV2(text string) string {
	specialChars := []string{"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}
	result := text

	for _, char := range specialChars {
		result = strings.ReplaceAll(result, "\\"+char, char)
	}

	return result
}

// cleanMarkdownV2 очищает MarkdownV2 разметку, убирая экранирование
func (tf *TelegramPostFormatter) cleanMarkdownV2(text string) string {
	// Убираем экранирование
	result := tf.UnescapeMarkdownV2(text)

	// Убираем лишние комментарии в начале и конце
	result = strings.TrimSpace(result)

	// Убираем "---" в начале и конце
	result = strings.TrimPrefix(result, "---")
	result = strings.TrimSuffix(result, "---")
	result = strings.TrimSpace(result)

	// Убираем комментарии типа "Вот привлекательный пост для Telegram с правильной разметкой:"
	lines := strings.Split(result, "\n")
	var cleanLines []string

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Пропускаем лишние комментарии
		if strings.Contains(trimmedLine, "Вот привлекательный пост") ||
			strings.Contains(trimmedLine, "---") ||
			strings.Contains(trimmedLine, "Вот готовый пост") ||
			strings.Contains(trimmedLine, "Вот пост") {
			continue
		}

		cleanLines = append(cleanLines, line)
	}

	result = strings.Join(cleanLines, "\n")

	// Убираем дублирование текста в конце
	// Ищем паттерн: хештеги + дублированный контент
	lines = strings.Split(result, "\n")
	var finalLines []string

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Если находим строку с хештегами, добавляем только её
		if strings.Contains(trimmedLine, "#") && strings.Contains(trimmedLine, "Тест") && strings.Contains(trimmedLine, "Форматирование") {
			finalLines = append(finalLines, line)
			// Останавливаемся здесь, не добавляем ничего после хештегов
			break
		} else {
			finalLines = append(finalLines, line)
		}
	}

	result = strings.Join(finalLines, "\n")
	return strings.TrimSpace(result)
}
