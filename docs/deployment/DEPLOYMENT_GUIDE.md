# 🚀 Руководство по развертыванию на сервере

## 🎯 Домен: aiwhisper.ru

### 1. **Подготовка сервера**

```bash
# Обновляем систему
sudo apt update && sudo apt upgrade -y

# Устанавливаем Docker и Docker Compose
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER

# Устанавливаем Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Перезагружаемся для применения изменений
sudo reboot
```

### 2. **Загрузка проекта**

```bash
# Клонируем репозиторий (или загружаем файлы)
git clone <your-repo-url> ai_tg_writer
cd ai_tg_writer

# Или загружаем файлы через SCP/SFTP
```

### 3. **Настройка переменных окружения**

```bash
# Создаем .env файл
cat > .env << EOF
# Telegram Bot
TELEGRAM_BOT_TOKEN=your_bot_token_here

# Database
DB_HOST=postgres
DB_PORT=5432
DB_NAME=ai_tg_writer
DB_USER=postgres
DB_PASSWORD=your_secure_password_here

# YooKassa
YOOKASSA_SHOP_ID=your_shop_id
YOOKASSA_SECRET_KEY=your_secret_key

# DeepSeek
DEEPSEEK_API_KEY=your_deepseek_api_key

# Whisper
WHISPER_API_KEY=your_whisper_api_key

# Redis
REDIS_URL=redis://redis:6379
EOF
```

### 4. **Настройка SSL и Nginx**

```bash
# Запускаем скрипт настройки SSL
./setup_ssl.sh
```

### 5. **Запуск приложения**

```bash
# Запускаем все сервисы
docker-compose up -d

# Запускаем приложение
./start_app.sh

# Проверяем статус
docker ps
./stop_app.sh && ./start_app.sh  # если нужно перезапустить
```

### 6. **Настройка автозапуска**

```bash
# Создаем systemd сервис для приложения
sudo tee /etc/systemd/system/ai-tg-writer.service > /dev/null << EOF
[Unit]
Description=AI TG Writer Bot
After=network.target

[Service]
Type=simple
User=$USER
WorkingDirectory=$(pwd)
ExecStart=$(pwd)/start_app.sh
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

# Включаем автозапуск
sudo systemctl enable ai-tg-writer.service
sudo systemctl start ai-tg-writer.service
```

## 🌐 Ваши адреса

### **Основное приложение**
- **URL**: https://aiwhisper.ru
- **API**: https://aiwhisper.ru/api/
- **Метрики**: https://aiwhisper.ru/metrics
- **Здоровье**: https://aiwhisper.ru/health

### **Мониторинг**
- **Grafana**: https://aiwhisper.ru/grafana/
- **Prometheus**: https://aiwhisper.ru/prometheus/
- **Jaeger**: https://aiwhisper.ru/jaeger/

### **Админ-панель мониторинга**
- **URL**: https://monitor.aiwhisper.ru
- **Логин**: admin / (пароль который вы создали)

## 🔧 Настройка YooKassa

### 1. **Webhook URL**
```
https://aiwhisper.ru/api/yookassa/callback
```

### 2. **Настройки в YooKassa**
- Зайдите в личный кабинет YooKassa
- Перейдите в раздел "Настройки" → "Webhook"
- Добавьте URL: `https://aiwhisper.ru/api/yookassa/callback`
- Выберите события: `payment.succeeded`, `payment.canceled`, `payment.waiting_for_capture`

## 🔒 Безопасность

### 1. **Смена паролей**
```bash
# В Grafana: Settings → Users → Admin → Change Password
# Новый пароль: ваш_очень_безопасный_пароль
```

### 2. **Настройка файрвола**
```bash
# Разрешаем только необходимые порты
sudo ufw allow 22    # SSH
sudo ufw allow 80    # HTTP
sudo ufw allow 443   # HTTPS
sudo ufw enable
```

### 3. **Регулярные обновления**
```bash
# Обновление системы
sudo apt update && sudo apt upgrade -y

# Обновление Docker образов
docker-compose pull
docker-compose up -d
```

## 📊 Мониторинг

### 1. **Проверка статуса**
```bash
# Статус сервисов
docker ps
sudo systemctl status ai-tg-writer

# Логи
docker logs ai_tg_writer_grafana
docker logs ai_tg_writer_prometheus
tail -f app.log
```

### 2. **Резервное копирование**
```bash
# Бэкап базы данных
docker exec ai_tg_writer_postgres pg_dump -U postgres ai_tg_writer > backup_$(date +%Y%m%d).sql

# Бэкап конфигурации
tar -czf config_backup_$(date +%Y%m%d).tar.gz monitoring/ docker-compose.yml .env
```

## 🚨 Troubleshooting

### 1. **Приложение не запускается**
```bash
# Проверяем логи
tail -f app.log

# Проверяем порты
netstat -tlnp | grep 8080

# Перезапускаем
./stop_app.sh && ./start_app.sh
```

### 2. **Grafana не открывается**
```bash
# Проверяем контейнер
docker ps | grep grafana

# Проверяем логи
docker logs ai_tg_writer_grafana

# Перезапускаем
docker-compose restart grafana
```

### 3. **SSL не работает**
```bash
# Проверяем сертификаты
sudo certbot certificates

# Обновляем сертификаты
sudo certbot renew

# Перезапускаем nginx
sudo systemctl reload nginx
```

## 🎯 Готово!

Теперь ваш бот доступен по адресу **https://aiwhisper.ru** с полным мониторингом! 🚀

**Не забудьте**:
- Сменить пароли по умолчанию
- Настроить webhook в YooKassa
- Настроить автозапуск
- Регулярно делать бэкапы
