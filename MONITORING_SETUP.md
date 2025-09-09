# 📊 Настройка мониторинга и логирования

## 🚀 Быстрый старт

### 1. Запуск системы мониторинга

```bash
# Запускаем все сервисы мониторинга
docker-compose up -d prometheus grafana jaeger

# Или запускаем все сервисы включая основное приложение
docker-compose up -d
```

### 2. Доступ к интерфейсам

- **Grafana**: http://localhost:3000 (admin/admin)
- **Prometheus**: http://localhost:9090
- **Jaeger**: http://localhost:16686
- **Приложение**: http://localhost:8080
- **Health Check**: http://localhost:8080/health
- **Метрики**: http://localhost:8080/metrics

## 📈 Что мониторится

### HTTP метрики
- Количество запросов по методам и эндпоинтам
- Время ответа (гистограмма)
- Коды ответов

### Бизнес метрики
- Обработка голосовых сообщений
- Время транскрипции (Whisper)
- Время генерации AI (DeepSeek)
- Активные пользователи
- Подписки (создание, отмена, продление)

### Системные метрики
- Подключения к базе данных
- Вызовы внешних API
- Telegram сообщения (входящие/исходящие)

## 🔧 Конфигурация

### Переменные окружения

```bash
# Режим работы (production/development)
ENV=production

# Jaeger endpoint (опционально)
JAEGER_ENDPOINT=http://localhost:14268/api/traces
```

### Prometheus конфигурация

Файл: `monitoring/prometheus.yml`

```yaml
scrape_configs:
  - job_name: 'ai_tg_writer'
    static_configs:
      - targets: ['host.docker.internal:8080']
    metrics_path: '/metrics'
    scrape_interval: 5s
```

### Grafana дашборды

Автоматически загружаются из:
- `monitoring/grafana/dashboards/ai_tg_writer.json`

## 📊 Дашборды

### Основной дашборд "AI TG Writer Monitoring"

1. **HTTP Requests Rate** - частота HTTP запросов
2. **Voice Messages Processed** - обработка голосовых сообщений
3. **Active Users** - количество активных пользователей
4. **Voice Processing Duration** - время обработки голосовых сообщений

### Создание собственных дашбордов

1. Откройте Grafana: http://localhost:3000
2. Войдите (admin/admin)
3. Создайте новый дашборд
4. Добавьте панели с метриками Prometheus

## 🔍 Логирование

### Структурированные логи

В production режиме логи выводятся в JSON формате:

```json
{
  "level": "info",
  "msg": "HTTP request",
  "method": "GET",
  "path": "/ping",
  "ip": "127.0.0.1",
  "trace_id": "1234567890",
  "time": "2024-01-01T12:00:00Z"
}
```

### Уровни логирования

- **DEBUG** - детальная отладочная информация
- **INFO** - общая информация о работе
- **WARN** - предупреждения
- **ERROR** - ошибки
- **FATAL** - критические ошибки

## 🚨 Алерты (будущая функция)

Можно настроить алерты в Prometheus:

```yaml
# monitoring/rules/alerts.yml
groups:
  - name: ai_tg_writer_alerts
    rules:
      - alert: HighErrorRate
        expr: rate(http_requests_total{status_code=~"5.."}[5m]) > 0.1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High error rate detected"
```

## 🛠 Устранение неполадок

### Prometheus не видит метрики

1. Проверьте, что приложение запущено на порту 8080
2. Убедитесь, что эндпоинт `/metrics` доступен
3. Проверьте конфигурацию в `monitoring/prometheus.yml`

### Grafana не загружает дашборды

1. Проверьте права доступа к папке `monitoring/grafana/`
2. Убедитесь, что provisioning настроен правильно
3. Перезапустите Grafana: `docker-compose restart grafana`

### Jaeger не показывает трейсы

1. Проверьте, что приложение отправляет трейсы
2. Убедитесь, что Jaeger запущен: `docker-compose ps jaeger`
3. Проверьте логи приложения на наличие ошибок трейсинга

## 📚 Полезные команды

```bash
# Просмотр логов приложения
docker-compose logs -f ai_tg_writer

# Просмотр логов Prometheus
docker-compose logs -f prometheus

# Просмотр логов Grafana
docker-compose logs -f grafana

# Перезапуск всех сервисов
docker-compose restart

# Остановка всех сервисов
docker-compose down

# Очистка данных (ОСТОРОЖНО!)
docker-compose down -v
```

## 🔄 Обновление

При изменении конфигурации мониторинга:

```bash
# Перезапустите соответствующие сервисы
docker-compose restart prometheus grafana

# Или пересоберите и перезапустите все
docker-compose down
docker-compose up -d --build
```
