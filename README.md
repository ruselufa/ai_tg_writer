# 🤖 AI TG Writer

Telegram бот для создания постов с использованием AI (Whisper + DeepSeek).

## 🚀 Быстрый старт

### Разработка
```bash
# Запуск приложения
./start_app.sh

# Запуск мониторинга
docker-compose up -d

# Остановка
./stop_app.sh
docker-compose down
```

### Продакшн
Смотрите [docs/DEPLOYMENT_GUIDE.md](docs/deployment/DEPLOYMENT_GUIDE.md)

## 📊 Мониторинг

- **Grafana**: http://localhost:3000 (admin/admin)
- **Prometheus**: http://localhost:9090
- **Метрики**: http://localhost:8080/metrics

## 📚 Документация

Полная документация в папке [docs/](docs/README.md)

## 🔧 Основные команды

```bash
# Управление приложением
./start_app.sh    # Запуск
./stop_app.sh     # Остановка

# Управление мониторингом
docker-compose up -d     # Запуск
docker-compose down      # Остановка
docker-compose restart   # Перезапуск

# Проверка здоровья
curl http://localhost:8080/health
curl http://localhost:8080/metrics
```

## 🌐 Продакшн адреса

- **Приложение**: https://aiwhisper.ru
- **Мониторинг**: https://aiwhisper.ru/grafana/
- **Админ-панель**: https://monitor.aiwhisper.ru

## 📁 Структура проекта

```
├── api/                    # HTTP API
├── cmd/ai_tg_writer/       # Главный файл приложения
├── internal/               # Внутренняя логика
│   ├── monitoring/         # Система мониторинга
│   ├── infrastructure/     # Инфраструктурные компоненты
│   └── service/           # Бизнес-логика
├── monitoring/             # Конфигурация мониторинга
├── docs/                  # Документация
├── migrations/            # Миграции БД
└── docker-compose.yml     # Docker сервисы
```

## 🔒 Безопасность

- Смените пароли по умолчанию
- Настройте HTTPS в продакшене
- Ограничьте доступ к мониторингу
- Регулярно обновляйте систему
