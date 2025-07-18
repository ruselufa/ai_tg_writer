# Настройка транскрипции голосовых сообщений

## Обзор

Бот настроен для работы с асинхронным API транскрипции, который поддерживает следующие эндпоинты:

- `POST /transcribe` - отправка аудио файла на транскрипцию
- `GET /status/{file_id}` - получение статуса транскрипции
- `GET /download/{file_id}` - скачивание результата
- `GET /metrics` - получение метрик сервиса

## Настройка

### 1. Переменные окружения

Создайте файл `.env` на основе `env.example`:

```bash
cp env.example .env
```

Обязательные переменные:
- `TELEGRAM_BOT_TOKEN` - токен вашего Telegram бота
- `WHISPER_API_URL` - URL вашего API транскрипции (по умолчанию: http://localhost:8000)

Опциональные переменные:
- `DEEPSEEK_API_KEY` - для переписывания текста с помощью ИИ

### 2. Запуск бота

```bash
go run cmd/ai_tg_writer/main.go
```

### 3. Тестирование API

Для тестирования API транскрипции используйте:

```bash
go run test_api.go path/to/audio/file.mp3
```

## Как это работает

1. **Получение голосового сообщения**: Бот получает голосовое сообщение от пользователя
2. **Скачивание файла**: Голосовое сообщение скачивается во временную папку
3. **Отправка на транскрипцию**: Файл отправляется на ваш API с параметром `language=ru`
4. **Ожидание результата**: Бот ждет завершения транскрипции (максимум 5 минут)
5. **Переписывание текста**: Если настроен DeepSeek API, текст переписывается для улучшения качества
6. **Отправка результата**: Пользователь получает готовый текст

## Структура ответов API

### POST /transcribe
```json
{
  "file_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "queued",
  "queue_position": 1,
  "metrics": {
    "queue_size": 1,
    "avg_processing_time": 2.34,
    "avg_processing_speed": 0.43,
    "files_processed": 5
  }
}
```

### GET /status/{file_id}
```json
{
  "status": "completed",
  "created_at": "2024-01-01T12:00:00",
  "input_path": "/temp/audio.mp3",
  "output_path": "/temp/result.txt",
  "completed_at": "2024-01-01T12:02:34",
  "processing_time": 2.34,
  "file_size": 1024000,
  "error": null,
  "queue_position": 0,
  "metrics": {...}
}
```

### GET /download/{file_id}
```json
{
  "text": "Распознанный текст из аудио"
}
```

## Логирование

Бот ведет подробные логи процесса транскрипции:
- Скачивание файла
- Отправка на транскрипцию
- Статус обработки
- Время обработки
- Ошибки (если есть)

## Обработка ошибок

- Если API недоступен - пользователь получает сообщение об ошибке
- Если транскрипция не удалась - возвращается исходный текст
- Если переписывание не работает - возвращается транскрибированный текст
- Временные файлы автоматически удаляются после обработки

## Производительность

- Максимальное время ожидания транскрипции: 5 минут
- Интервал проверки статуса: 2 секунды
- Автоматическая очистка старых файлов: каждый час 