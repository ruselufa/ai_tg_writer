#!/bin/bash

# Скрипт для запуска AI TG Writer
# Использование: ./start_app.sh

echo "🚀 Запуск AI TG Writer..."

# Проверяем, не запущено ли уже приложение
if lsof -i :8080 >/dev/null 2>&1; then
    echo "❌ Порт 8080 уже занят. Остановите приложение сначала."
    echo "Для остановки используйте: pkill -f 'go run.*ai_tg_writer'"
    exit 1
fi

# Проверяем наличие .env файла
if [ ! -f .env ]; then
    echo "❌ Файл .env не найден!"
    echo "Создайте файл .env с необходимыми переменными окружения."
    exit 1
fi

# Запускаем приложение
echo "✅ Запускаем приложение..."
go run ./cmd/ai_tg_writer
