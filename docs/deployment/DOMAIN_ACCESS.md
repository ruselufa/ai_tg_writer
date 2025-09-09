# 🌐 Доступ через домен aiwhisper.ru

## 🎯 Ваши адреса

**Домен**: `aiwhisper.ru`

## 🌐 Основные адреса

### **🤖 Основное приложение**
- **URL**: https://aiwhisper.ru
- **API**: https://aiwhisper.ru/api/
- **Метрики**: https://aiwhisper.ru/metrics
- **Здоровье**: https://aiwhisper.ru/health
- **Тест метрик**: https://aiwhisper.ru/test-metrics

### **📊 Мониторинг (публичный)**
- **Grafana**: https://aiwhisper.ru/grafana/
- **Prometheus**: https://aiwhisper.ru/prometheus/
- **Jaeger**: https://aiwhisper.ru/jaeger/

### **🔐 Админ-панель мониторинга**
- **URL**: https://monitor.aiwhisper.ru
- **Логин**: `admin`
- **Пароль**: (пароль который вы создали при настройке)

## 🔧 Настройка YooKassa

### **Webhook URL для колбеков**
```
https://aiwhisper.ru/api/yookassa/callback
```

### **Настройки в YooKassa**
1. Зайдите в личный кабинет YooKassa
2. Перейдите в "Настройки" → "Webhook"
3. Добавьте URL: `https://aiwhisper.ru/api/yookassa/callback`
4. Выберите события:
   - `payment.succeeded`
   - `payment.canceled`
   - `payment.waiting_for_capture`

## 📱 Доступ с мобильных устройств

### **Grafana Mobile App**
- **iOS**: Grafana Mobile
- **Android**: Grafana Mobile
- **URL**: https://aiwhisper.ru/grafana/
- **Логин**: `admin` / `ваш_пароль`

### **Веб-браузер**
Любой современный браузер поддерживает все интерфейсы.

## 🔒 Безопасность

### **Обязательно сделайте:**

1. **Смените пароль admin в Grafana**
   - Зайдите в https://aiwhisper.ru/grafana/
   - Settings → Users → Admin → Change Password

2. **Настройте базовую аутентификацию**
   - Админ-панель: https://monitor.aiwhisper.ru
   - Защищена паролем

3. **Ограничьте доступ по IP (рекомендуется)**
   - В nginx конфигурации добавьте ваши IP

## 🚀 Развертывание на сервере

### **Быстрый старт:**
```bash
# 1. Загрузите проект на сервер
# 2. Настройте .env файл с вашими ключами
# 3. Запустите настройку SSL
./setup_ssl.sh

# 4. Запустите приложение
docker-compose up -d
./start_app.sh
```

### **Подробная инструкция:**
Смотрите `DEPLOYMENT_GUIDE.md`

## 📊 Дашборды

### **1. Основной дашборд**
- **URL**: https://aiwhisper.ru/grafana/d/ai-tg-writer-monitoring
- **Описание**: Технические метрики, производительность, система

### **2. Бизнес-дашборд**
- **URL**: https://aiwhisper.ru/grafana/d/ai-tg-writer-business-metrics
- **Описание**: Конверсии, ARPU, удержание, сравнение тарифов

## 🔧 Управление

### **Команды для управления:**
```bash
# Остановить все сервисы
docker-compose down

# Запустить все сервисы
docker-compose up -d

# Перезапустить приложение
./stop_app.sh && ./start_app.sh

# Посмотреть логи
docker logs ai_tg_writer_grafana
tail -f app.log
```

### **Проверка статуса:**
```bash
# Статус контейнеров
docker ps

# Статус приложения
curl https://aiwhisper.ru/health

# Статус метрик
curl https://aiwhisper.ru/metrics
```

## 🚨 Если что-то не работает

### **1. Проверьте DNS**
```bash
# Проверка разрешения домена
nslookup aiwhisper.ru
ping aiwhisper.ru
```

### **2. Проверьте SSL**
```bash
# Проверка сертификата
openssl s_client -connect aiwhisper.ru:443 -servername aiwhisper.ru
```

### **3. Проверьте сервисы**
```bash
# Статус nginx
sudo systemctl status nginx

# Статус Docker
docker ps

# Логи nginx
sudo tail -f /var/log/nginx/error.log
```

## 🎯 Готово!

Теперь ваш бот доступен по адресу **https://aiwhisper.ru** с полным мониторингом! 🚀

**Преимущества домена:**
- ✅ Красивые URL вместо IP адресов
- ✅ SSL сертификаты для безопасности
- ✅ Легко запомнить и использовать
- ✅ Профессиональный вид
- ✅ Работает из любой точки мира
