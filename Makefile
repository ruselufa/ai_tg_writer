.PHONY: help build run test clean docker-build docker-run docker-stop install-deps lint format

# Переменные
BINARY_NAME=ai_tg_writer
DOCKER_IMAGE=ai_tg_writer
DOCKER_CONTAINER=ai_tg_writer_bot

# Цвета для вывода
GREEN=\033[0;32m
YELLOW=\033[1;33m
RED=\033[0;31m
NC=\033[0m # No Color

help: ## Показать справку
	@echo "$(GREEN)Доступные команды:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(YELLOW)%-15s$(NC) %s\n", $$1, $$2}'

install-deps: ## Установить зависимости
	@echo "$(GREEN)Устанавливаем зависимости...$(NC)"
	go mod download
	go mod tidy

build: ## Собрать приложение
	@echo "$(GREEN)Собираем приложение...$(NC)"
	go build -o $(BINARY_NAME) main.go

run: ## Запустить приложение
	@echo "$(GREEN)Запускаем приложение...$(NC)"
	go run main.go

test: ## Запустить тесты
	@echo "$(GREEN)Запускаем тесты...$(NC)"
	go test ./...

clean: ## Очистить собранные файлы
	@echo "$(GREEN)Очищаем файлы...$(NC)"
	rm -f $(BINARY_NAME)
	rm -rf audio/
	go clean

lint: ## Проверить код линтером
	@echo "$(GREEN)Проверяем код...$(NC)"
	golangci-lint run

format: ## Форматировать код
	@echo "$(GREEN)Форматируем код...$(NC)"
	go fmt ./...
	go vet ./...

docker-build: ## Собрать Docker образ
	@echo "$(GREEN)Собираем Docker образ...$(NC)"
	docker build -t $(DOCKER_IMAGE) .

docker-run: ## Запустить с Docker Compose
	@echo "$(GREEN)Запускаем с Docker Compose...$(NC)"
	docker-compose up -d

docker-stop: ## Остановить Docker Compose
	@echo "$(GREEN)Останавливаем Docker Compose...$(NC)"
	docker-compose down

docker-logs: ## Показать логи Docker
	@echo "$(GREEN)Показываем логи...$(NC)"
	docker-compose logs -f

docker-clean: ## Очистить Docker
	@echo "$(GREEN)Очищаем Docker...$(NC)"
	docker-compose down -v
	docker rmi $(DOCKER_IMAGE)

dev: ## Запустить в режиме разработки
	@echo "$(GREEN)Запускаем в режиме разработки...$(NC)"
	@if [ ! -f .env ]; then \
		echo "$(RED)Файл .env не найден! Скопируйте env.example в .env и настройте переменные.$(NC)"; \
		exit 1; \
	fi
	go run main.go

setup: ## Настройка проекта
	@echo "$(GREEN)Настраиваем проект...$(NC)"
	@if [ ! -f .env ]; then \
		cp env.example .env; \
		echo "$(YELLOW)Создан файл .env. Отредактируйте его и добавьте необходимые переменные.$(NC)"; \
	else \
		echo "$(GREEN)Файл .env уже существует.$(NC)"; \
	fi
	go mod tidy
	@echo "$(GREEN)Настройка завершена!$(NC)"

check-env: ## Проверить переменные окружения
	@echo "$(GREEN)Проверяем переменные окружения...$(NC)"
	@if [ ! -f .env ]; then \
		echo "$(RED)Файл .env не найден!$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)Переменные окружения:$(NC)"
	@source .env && echo "TELEGRAM_BOT_TOKEN: $${TELEGRAM_BOT_TOKEN:+***}"
	@source .env && echo "OPENAI_API_KEY: $${OPENAI_API_KEY:+***}"
	@source .env && echo "DB_HOST: $${DB_HOST:-не установлен}"
	@source .env && echo "DB_NAME: $${DB_NAME:-не установлен}"

# Команды для работы с базой данных
db-create: ## Создать базу данных (требует PostgreSQL)
	@echo "$(GREEN)Создаем базу данных...$(NC)"
	@source .env && psql -h $${DB_HOST:-localhost} -U $${DB_USER:-postgres} -c "CREATE DATABASE $${DB_NAME:-ai_tg_writer};" || echo "$(YELLOW)База данных уже существует или ошибка подключения$(NC)"

db-init: ## Инициализировать таблицы базы данных
	@echo "$(GREEN)Инициализируем таблицы...$(NC)"
	@source .env && psql -h $${DB_HOST:-localhost} -U $${DB_USER:-postgres} -d $${DB_NAME:-ai_tg_writer} -f database/init.sql

db-reset: ## Сбросить базу данных (ОПАСНО!)
	@echo "$(RED)ВНИМАНИЕ: Это удалит все данные!$(NC)"
	@read -p "Продолжить? (y/N): " confirm && [ "$$confirm" = "y" ] || exit 1
	@source .env && psql -h $${DB_HOST:-localhost} -U $${DB_USER:-postgres} -c "DROP DATABASE IF EXISTS $${DB_NAME:-ai_tg_writer};"
	@source .env && psql -h $${DB_HOST:-localhost} -U $${DB_USER:-postgres} -c "CREATE DATABASE $${DB_NAME:-ai_tg_writer};"
	@source .env && psql -h $${DB_HOST:-localhost} -U $${DB_USER:-postgres} -d $${DB_NAME:-ai_tg_writer} -f database/init.sql

# Команды для мониторинга
logs: ## Показать логи приложения
	@echo "$(GREEN)Логи приложения:$(NC)"
	@if [ -f ai_tg_writer.log ]; then \
		tail -f ai_tg_writer.log; \
	else \
		echo "$(YELLOW)Файл логов не найден$(NC)"; \
	fi

status: ## Показать статус сервисов
	@echo "$(GREEN)Статус сервисов:$(NC)"
	@docker-compose ps

# Команды для деплоя
deploy: ## Деплой на сервер (требует настройки)
	@echo "$(GREEN)Деплой на сервер...$(NC)"
	@echo "$(YELLOW)Эта команда требует дополнительной настройки$(NC)"

# Команды для разработки
watch: ## Запустить с автоперезагрузкой (требует air)
	@echo "$(GREEN)Запускаем с автоперезагрузкой...$(NC)"
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "$(YELLOW)Установите air: go install github.com/cosmtrek/air@latest$(NC)"; \
		go run main.go; \
	fi

# Информационные команды
info: ## Показать информацию о проекте
	@echo "$(GREEN)Информация о проекте:$(NC)"
	@echo "Версия Go: $(shell go version)"
	@echo "Путь к Go: $(shell which go)"
	@echo "Рабочая директория: $(shell pwd)"
	@echo "Размер проекта: $(shell du -sh . | cut -f1)"
	@echo "Количество Go файлов: $(shell find . -name "*.go" | wc -l)"

version: ## Показать версию
	@echo "$(GREEN)Версия:$(NC)"
	@echo "AI Voice Writer Bot v1.0.0"
	@echo "Go $(shell go version | cut -d' ' -f3)" 