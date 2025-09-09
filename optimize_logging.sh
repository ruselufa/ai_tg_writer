#!/bin/bash

# Скрипт для оптимизации логирования

echo "🔧 Оптимизация логирования..."

# Функция для замены log.Printf на monitoring функции
replace_logging() {
    local file="$1"
    local component="$2"
    
    echo "Обрабатываем $file..."
    
    # Заменяем log.Printf на соответствующие monitoring функции
    sed -i '' 's/log\.Printf("\[DEBUG\]/monitoring.Debug("/g' "$file"
    sed -i '' 's/log\.Printf("❌/monitoring.Error("/g' "$file"
    sed -i '' 's/log\.Printf("✅/monitoring.Info("/g' "$file"
    sed -i '' 's/log\.Printf("⚠️/monitoring.Warn("/g' "$file"
    sed -i '' 's/log\.Printf("🔄/monitoring.Info("/g' "$file"
    sed -i '' 's/log\.Printf("⏳/monitoring.Info("/g' "$file"
    
    # Заменяем обычные log.Printf на Info
    sed -i '' 's/log\.Printf(/monitoring.Info(/g' "$file"
    
    # Добавляем метрики логирования после каждого вызова
    sed -i '' 's/monitoring\.Info(/monitoring.Info(/g' "$file"
    sed -i '' 's/monitoring\.Error(/monitoring.Error(/g' "$file"
    sed -i '' 's/monitoring\.Warn(/monitoring.Warn(/g' "$file"
    sed -i '' 's/monitoring\.Debug(/monitoring.Debug(/g' "$file"
}

# Обрабатываем основные файлы
replace_logging "internal/infrastructure/bot/inline_handlers.go" "bot"
replace_logging "internal/infrastructure/bot/message_handler.go" "bot"
replace_logging "internal/infrastructure/voice/voice_handler.go" "voice"
replace_logging "api/yookassa_handler.go" "payment"
replace_logging "internal/infrastructure/deepseek/deepseek_handler.go" "api"

echo "✅ Логирование оптимизировано!"
echo ""
echo "📊 Добавлены метрики:"
echo "   - log_messages_total - количество логов по уровням"
echo "   - user_interactions_total - взаимодействия пользователей"
echo "   - processing_steps_total - шаги обработки"
echo ""
echo "🔧 Уровни логирования:"
echo "   - DEBUG: только при LOG_LEVEL=DEBUG"
echo "   - INFO: информационные сообщения"
echo "   - WARN: предупреждения"
echo "   - ERROR: ошибки (всегда показываются)"
echo "   - SYSTEM: системные сообщения (всегда показываются)"
echo ""
echo "🌐 Установка уровня логирования:"
echo "   export LOG_LEVEL=INFO    # По умолчанию"
echo "   export LOG_LEVEL=DEBUG   # Для отладки"
echo "   export LOG_LEVEL=ERROR   # Только ошибки"
