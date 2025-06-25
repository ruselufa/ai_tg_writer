# AI Voice Writer - Telegram Bot

Telegram-бот для переписывания голосовых сообщений в красивый текст с помощью ИИ.

## 🎯 Возможности

- 🎤 Прием голосовых сообщений
- 🧠 Распознавание речи
- ✨ Переписывание текста с помощью ИИ
- 👤 Личный кабинет с лимитами
- 💳 Система подписок
- 🤝 Готовность к партнёрской программе

## 📋 Требования

- Go 1.21+
- PostgreSQL
- Telegram Bot Token
- OpenAI API Key (опционально)

## 🚀 Установка и настройка

### 1. Клонирование репозитория

```bash
git clone <repository-url>
cd ai_tg_writer
```

### 2. Установка зависимостей

```bash
go mod tidy
```

### 3. Настройка базы данных

Создайте базу данных PostgreSQL:

```sql
CREATE DATABASE ai_tg_writer;
```

### 4. Настройка переменных окружения

Скопируйте файл конфигурации:

```bash
cp env.example .env
```

Отредактируйте `.env` файл:

```env
# Telegram Bot Token (получите у @BotFather)
TELEGRAM_BOT_TOKEN=your_bot_token_here

# OpenAI API Key (для переписывания текста)
OPENAI_API_KEY=your_openai_api_key_here

# Database configuration
DB_HOST=localhost
DB_PORT=5432
DB_NAME=ai_tg_writer
DB_USER=postgres
DB_PASSWORD=your_password_here

# Admin username for support
ADMIN_USERNAME=admin_username
```

### 5. Создание Telegram бота

1. Найдите @BotFather в Telegram
2. Отправьте команду `/newbot`
3. Следуйте инструкциям для создания бота
4. Скопируйте полученный токен в `.env` файл

### 6. Получение OpenAI API Key

1. Зарегистрируйтесь на [OpenAI](https://platform.openai.com/)
2. Создайте API ключ
3. Добавьте ключ в `.env` файл

## 🏃‍♂️ Запуск

### Разработка

```bash
go run main.go
```

### Продакшн

```bash
go build -o ai_tg_writer main.go
./ai_tg_writer
```

## 📁 Структура проекта

```
ai_tg_writer/
├── main.go                 # Основной файл приложения
├── go.mod                  # Go модуль
├── go.sum                  # Зависимости
├── .env                    # Переменные окружения
├── .gitignore             # Исключения Git
├── README.md              # Документация
├── database/              # Работа с базой данных
│   └── database.go
├── voice/                 # Обработка голосовых сообщений
│   └── voice_handler.go
├── ai/                    # Работа с ИИ
│   └── openai_handler.go
└── audio/                 # Временные аудио файлы (создается автоматически)
```

## 🎮 Использование

### Команды бота

- `/start` - Начать работу с ботом
- `/help` - Получить справку
- `/profile` - Просмотр профиля и лимитов
- `/subscription` - Информация о подписке

### Работа с голосовыми сообщениями

1. Отправьте голосовое сообщение боту
2. Бот скачает и обработает файл
3. Получите переписанный текст

## 🔧 Разработка

### План разработки (2 недели)

#### Дни 1-2: Базовая настройка
- ✅ Настройка бота
- ✅ Обработка команд
- ✅ Приветственные сообщения

#### Дни 3-4: Работа с голосовыми
- 🔄 Прием голосовых сообщений
- 🔄 Сохранение файлов

#### Дни 5-6: Распознавание речи
- ⏳ Интеграция с сервисом распознавания
- ⏳ Преобразование аудио в текст

#### Дни 7-8: ИИ для переписывания
- ⏳ Интеграция с OpenAI
- ⏳ Переписывание текста

#### Дни 9-10: Личный кабинет
- ⏳ Профиль пользователя
- ⏳ Система лимитов
- ⏳ Статистика использования

#### Дни 11-12: Подписка
- ⏳ Информация о тарифах
- ⏳ Ручное подключение подписок

#### Дни 13-14: Тестирование и деплой
- ⏳ Тестирование всех функций
- ⏳ Деплой на сервер
- ⏳ Первые пользователи

## 🗄️ База данных

### Таблицы

#### users
- `id` - ID пользователя Telegram
- `username` - Имя пользователя
- `first_name` - Имя
- `last_name` - Фамилия
- `tariff` - Тариф (free/premium)
- `usage_count` - Количество использований
- `last_usage` - Последнее использование
- `created_at` - Дата регистрации
- `referral_code` - Реферальный код
- `referred_by` - Кто пригласил

#### voice_messages
- `id` - Уникальный ID
- `user_id` - ID пользователя
- `file_id` - ID файла в Telegram
- `duration` - Длительность (секунды)
- `file_size` - Размер файла
- `text` - Распознанный текст
- `rewritten` - Переписанный текст
- `created_at` - Дата создания

#### usage_stats
- `id` - Уникальный ID
- `user_id` - ID пользователя
- `date` - Дата
- `usage_count` - Количество использований

## 🔒 Безопасность

- Все токены хранятся в переменных окружения
- Файл `.env` исключен из Git
- Временные аудио файлы автоматически удаляются
- Проверка лимитов использования

## 🚀 Деплой

### Docker (рекомендуется)

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o ai_tg_writer main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/ai_tg_writer .
CMD ["./ai_tg_writer"]
```

### Docker Compose

```yaml
version: '3.8'
services:
  bot:
    build: .
    environment:
      - TELEGRAM_BOT_TOKEN=${TELEGRAM_BOT_TOKEN}
      - OPENAI_API_KEY=${OPENAI_API_KEY}
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_NAME=ai_tg_writer
      - DB_USER=postgres
      - DB_PASSWORD=${DB_PASSWORD}
    depends_on:
      - postgres
    restart: unless-stopped

  postgres:
    image: postgres:15
    environment:
      - POSTGRES_DB=ai_tg_writer
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    restart: unless-stopped

volumes:
  postgres_data:
```

## 📞 Поддержка

При возникновении проблем:

1. Проверьте логи бота
2. Убедитесь в правильности настроек в `.env`
3. Проверьте подключение к базе данных
4. Обратитесь к администратору: @admin_username

## 📄 Лицензия

MIT License

## 🤝 Вклад в проект

1. Форкните репозиторий
2. Создайте ветку для новой функции
3. Внесите изменения
4. Создайте Pull Request

---

**Статус разработки**: В процессе (дни 1-2 из 14) 