# 📚 Документация AI TG Writer

## 🎯 Обзор

Полная документация по системе мониторинга, развертыванию и управлению AI TG Writer ботом.

## 📁 Структура документации

### 🚀 **Быстрый старт**
- [QUICKSTART.md](./QUICKSTART.md) - Быстрый старт для разработки
- [QUICKSTART_PRODAMUS.md](./QUICKSTART_PRODAMUS.md) - Настройка Prodamus
- [TRANSCRIPTION_SETUP.md](./TRANSCRIPTION_SETUP.md) - Настройка транскрипции

### 📊 **Мониторинг**
- [ADVANCED_METRICS.md](./monitoring/ADVANCED_METRICS.md) - Расширенные метрики
- [BUSINESS_DASHBOARD.md](./monitoring/BUSINESS_DASHBOARD.md) - Бизнес-дашборд
- [ACTIVE_USERS_SYSTEM.md](./monitoring/ACTIVE_USERS_SYSTEM.md) - Система активных пользователей
- [DATABASE_MONITORING.md](./monitoring/DATABASE_MONITORING.md) - Мониторинг базы данных
- [SYSTEM_MONITORING.md](./monitoring/SYSTEM_MONITORING.md) - Системный мониторинг

### 🚀 **Развертывание**
- [DEPLOYMENT_GUIDE.md](./deployment/DEPLOYMENT_GUIDE.md) - Полное руководство по развертыванию
- [DOMAIN_ACCESS.md](./deployment/DOMAIN_ACCESS.md) - Доступ через домен aiwhisper.ru
- [EXTERNAL_ACCESS_SETUP.md](./deployment/EXTERNAL_ACCESS_SETUP.md) - Настройка внешнего доступа
- [EXTERNAL_QUICK_ACCESS.md](./deployment/EXTERNAL_QUICK_ACCESS.md) - Быстрый доступ извне
- [NGINX_ADDITIONS.md](./deployment/NGINX_ADDITIONS.md) - Добавления в nginx.conf
- [QUICK_ACCESS.md](./deployment/QUICK_ACCESS.md) - Быстрый доступ к мониторингу
- [REMOTE_ACCESS_SETUP.md](./deployment/REMOTE_ACCESS_SETUP.md) - Настройка удаленного доступа

### 🔧 **Устранение неполадок**
- [APP_CONTROL.md](./troubleshooting/APP_CONTROL.md) - Управление приложением

### 📋 **История изменений**
- [CHANGELOG_ASYNC.md](./CHANGELOG_ASYNC.md) - Асинхронная обработка
- [CHANGELOG_POST_FIX.md](./CHANGELOG_POST_FIX.md) - Исправления постов
- [ASYNC_PROCESSING.md](./ASYNC_PROCESSING.md) - Асинхронная обработка
- [DEEPSEEK_TIMEOUT_FIX.md](./DEEPSEEK_TIMEOUT_FIX.md) - Исправление таймаутов DeepSeek
- [MESSAGE_ORDER_FIX.md](./MESSAGE_ORDER_FIX.md) - Исправление порядка сообщений
- [POST_HISTORY_FIXES.md](./POST_HISTORY_FIXES.md) - Исправления истории постов
- [POST_HISTORY_SYSTEM.md](./POST_HISTORY_SYSTEM.md) - Система истории постов
- [POST_MESSAGE_PRESERVATION.md](./POST_MESSAGE_PRESERVATION.md) - Сохранение сообщений
- [SUBSCRIPTION_ASYNC_SAFETY.md](./SUBSCRIPTION_ASYNC_SAFETY.md) - Безопасность подписок

### 🔌 **Интеграции**
- [PRODAMUS_INTEGRATION.md](./PRODAMUS_INTEGRATION.md) - Интеграция с Prodamus

### 📖 **Техническая документация**
- [IMPLEMENTATION_SUMMARY.md](./IMPLEMENTATION_SUMMARY.md) - Сводка по реализации

## 🌐 Основные адреса

### **Производство (aiwhisper.ru)**
- **Приложение**: https://aiwhisper.ru
- **Grafana**: https://aiwhisper.ru/grafana/
- **Prometheus**: https://aiwhisper.ru/prometheus/
- **Админ-панель**: https://monitor.aiwhisper.ru

### **Разработка (localhost)**
- **Приложение**: http://localhost:8080
- **Grafana**: http://localhost:3000
- **Prometheus**: http://localhost:9090

## 🔧 Быстрые команды

### **Управление приложением**
```bash
# Запуск
./start_app.sh

# Остановка
./stop_app.sh

# Статус
docker ps
```

### **Управление мониторингом**
```bash
# Запуск всех сервисов
docker-compose up -d

# Остановка всех сервисов
docker-compose down

# Перезапуск Grafana
docker-compose restart grafana
```

### **Проверка здоровья**
```bash
# Приложение
curl http://localhost:8080/health

# Метрики
curl http://localhost:8080/metrics

# Тест метрик
curl http://localhost:8080/test-metrics
```

## 🚨 Экстренные контакты

- **Логи приложения**: `tail -f app.log`
- **Логи Grafana**: `docker logs ai_tg_writer_grafana`
- **Логи Prometheus**: `docker logs ai_tg_writer_prometheus`

## 📝 Примечания

- Все пароли по умолчанию: `admin`
- **ОБЯЗАТЕЛЬНО** смените пароли в продакшене
- Регулярно обновляйте SSL сертификаты
- Делайте резервные копии базы данных