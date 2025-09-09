# 📚 Система логирования истории постов

## 🎯 Обзор

Система логирования истории постов (`post_history`) предназначена для отслеживания всего процесса создания постов от голосового сообщения до готового результата. Это позволяет анализировать производительность, отслеживать использование и предоставлять детальную статистику пользователям.

## 🗄️ Структура базы данных

### Таблица `post_history`

```sql
CREATE TABLE post_history (
    id SERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    
    -- Голосовое сообщение
    voice_text TEXT NOT NULL,
    voice_file_id VARCHAR(255),
    voice_duration INTEGER,
    voice_file_size INTEGER,
    
    -- Тайминги
    voice_sent_at TIMESTAMP NOT NULL,           -- Время отправки в Whisper
    voice_received_at TIMESTAMP,                -- Время получения от Whisper
    ai_sent_at TIMESTAMP,                       -- Время отправки в AI LLM
    ai_received_at TIMESTAMP,                   -- Время получения от AI LLM
    
    -- AI ответ
    ai_response TEXT,                           -- Полный ответ от AI
    ai_model VARCHAR(100) DEFAULT 'deepseek',   -- Тип нейросети
    ai_tokens_used INTEGER,                     -- Количество токенов
    ai_cost DECIMAL(10,6),                     -- Стоимость запроса
    
    -- Статус
    is_saved BOOLEAN DEFAULT FALSE,             -- Сохранил ли пользователь пост
    saved_at TIMESTAMP,                         -- Когда сохранил
    
    -- Метаданные
    processing_duration_ms INTEGER,             -- Общее время обработки
    whisper_duration_ms INTEGER,                -- Время транскрипции
    ai_generation_duration_ms INTEGER,          -- Время генерации AI
    
    -- Системные поля
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Индексы

- `idx_post_history_user_id` - для быстрого поиска по пользователю
- `idx_post_history_created_at` - для сортировки по времени создания
- `idx_post_history_is_saved` - для фильтрации сохраненных постов
- `idx_post_history_ai_model` - для анализа по типам нейросетей
- `idx_post_history_voice_sent_at` - для анализа по времени отправки

## 🔧 Архитектура

### Компоненты

1. **PostHistoryRepository** - репозиторий для работы с базой данных
2. **VoiceHandler** - обработчик голосовых сообщений с логированием
3. **StateManager** - управление состоянием пользователя с historyID
4. **InlineHandler** - обработка действий пользователя с отметкой сохранения

### Поток данных

```
Голосовое сообщение → Создание записи в истории → Транскрипция → 
Обновление истории → Генерация AI → Обновление истории → 
Отметка сохранения (если пользователь сохранил)
```

## 📊 Возможности

### Отслеживание времени

- **Общее время обработки** - от получения голосового до готового поста
- **Время транскрипции** - только процесс распознавания речи
- **Время генерации AI** - только процесс создания поста

### Аналитика

- **Количество токенов** - для расчета стоимости
- **Стоимость запросов** - для финансового учета
- **Статистика по пользователям** - общее количество, сохраненные посты
- **Производительность** - средние времена обработки

### Статусы

- **Сохраненные посты** - какие посты пользователь принял
- **Время сохранения** - когда именно пользователь сохранил пост

## 🚀 Использование

### Создание записи

```go
history := &database.PostHistory{
    UserID:        userID,
    VoiceText:     "",
    VoiceFileID:   fileID,
    VoiceDuration: duration,
    VoiceFileSize: fileSize,
    VoiceSentAt:   time.Now().UTC(),
    AIModel:       "deepseek",
}

err := postHistoryRepo.CreatePostHistory(history)
```

### Обновление записи

```go
history.VoiceText = transcribedText
history.VoiceReceivedAt = &voiceReceivedAt
history.WhisperDurationMs = &whisperDurationMs

err := postHistoryRepo.UpdatePostHistory(history)
```

### Отметка как сохраненный

```go
err := postHistoryRepo.MarkAsSaved(historyID)
```

### Получение статистики

```go
stats, err := postHistoryRepo.GetPostHistoryStats(userID)
```

## 📈 Аналитика и отчеты

### Статистика пользователя

- Общее количество постов
- Количество сохраненных постов
- Среднее время обработки
- Общая стоимость использования
- Количество использованных токенов

### Производительность системы

- Время транскрипции по типам аудио
- Время генерации по типам контента
- Загрузка по времени суток
- Эффективность использования ресурсов

## 🔒 Безопасность

- **Каскадное удаление** - при удалении пользователя удаляется вся его история
- **Изоляция данных** - пользователи видят только свою историю
- **Автоматическое обновление** - `updated_at` обновляется автоматически

## 🚧 Ограничения

- **Размер ответа AI** - ограничен типом `TEXT` в PostgreSQL
- **Время хранения** - пока не реализована автоматическая очистка старых записей
- **Масштабируемость** - для больших объемов может потребоваться партиционирование

## 🔮 Планы развития

1. **Автоматическая очистка** - удаление записей старше определенного возраста
2. **Экспорт данных** - возможность выгрузки истории в CSV/JSON
3. **Уведомления** - оповещения о превышении лимитов
4. **Дашборд аналитики** - веб-интерфейс для просмотра статистики
5. **Интеграция с внешними системами** - экспорт в CRM, аналитические платформы

## 📝 Примеры запросов

### Топ пользователей по активности

```sql
SELECT user_id, COUNT(*) as post_count
FROM post_history 
GROUP BY user_id 
ORDER BY post_count DESC 
LIMIT 10;
```

### Среднее время обработки по дням

```sql
SELECT 
    DATE(created_at) as date,
    AVG(processing_duration_ms) as avg_duration
FROM post_history 
GROUP BY DATE(created_at)
ORDER BY date DESC;
```

### Стоимость использования по пользователям

```sql
SELECT 
    user_id,
    SUM(ai_cost) as total_cost,
    COUNT(*) as post_count
FROM post_history 
WHERE ai_cost IS NOT NULL
GROUP BY user_id
ORDER BY total_cost DESC;
```

## 🎉 Заключение

Система логирования истории постов предоставляет мощный инструмент для анализа использования, оптимизации производительности и улучшения пользовательского опыта. Она позволяет точно отслеживать все этапы создания контента и предоставлять детальную аналитику как пользователям, так и администраторам системы.
