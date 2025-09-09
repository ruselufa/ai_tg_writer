#!/bin/bash

# Скрипт для остановки AI TG Writer
# Использование: ./stop_app.sh

echo "🛑 Остановка AI TG Writer..."

# Ищем и останавливаем все процессы ai_tg_writer
PIDS=$(pgrep -f "ai_tg_writer")
if [ -n "$PIDS" ]; then
    echo "Найдены процессы: $PIDS"
    kill -9 $PIDS
    echo "✅ Процессы остановлены"
else
    echo "ℹ️  Процессы ai_tg_writer не найдены"
fi

# Проверяем порт 8080
if lsof -i :8080 >/dev/null 2>&1; then
    echo "⚠️  Порт 8080 все еще занят. Принудительная очистка..."
    lsof -ti :8080 | xargs kill -9
    echo "✅ Порт 8080 освобожден"
else
    echo "✅ Порт 8080 свободен"
fi

echo "🎉 Остановка завершена"
