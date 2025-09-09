#!/bin/bash

# Скрипт для настройки SSL для aiwhisper.ru

echo "🔐 Настройка SSL для aiwhisper.ru"

# Обновляем систему
sudo apt update

# Устанавливаем nginx и certbot
sudo apt install -y nginx certbot python3-certbot-nginx

# Останавливаем nginx
sudo systemctl stop nginx

# Копируем конфигурацию
sudo cp nginx.conf /etc/nginx/sites-available/aiwhisper.ru
sudo ln -sf /etc/nginx/sites-available/aiwhisper.ru /etc/nginx/sites-enabled/
sudo rm -f /etc/nginx/sites-enabled/default

# Создаем базовую аутентификацию для мониторинга
echo "Создайте пароль для доступа к мониторингу:"
sudo htpasswd -c /etc/nginx/.htpasswd admin

# Тестируем конфигурацию
sudo nginx -t

if [ $? -eq 0 ]; then
    echo "✅ Конфигурация nginx корректна"
    
    # Запускаем nginx
    sudo systemctl start nginx
    sudo systemctl enable nginx
    
    echo "🌐 Получаем SSL сертификат..."
    sudo certbot --nginx -d aiwhisper.ru -d www.aiwhisper.ru -d monitor.aiwhisper.ru
    
    echo "🔄 Перезапускаем nginx с SSL"
    sudo systemctl reload nginx
    
    echo "✅ SSL настроен!"
    echo ""
    echo "🌐 Ваши адреса:"
    echo "   Основное приложение: https://aiwhisper.ru"
    echo "   Мониторинг: https://monitor.aiwhisper.ru"
    echo "   Grafana: https://aiwhisper.ru/grafana/"
    echo "   Prometheus: https://aiwhisper.ru/prometheus/"
    echo ""
    echo "🔐 Логин для мониторинга: admin / (пароль который вы создали)"
    
else
    echo "❌ Ошибка в конфигурации nginx"
    exit 1
fi
