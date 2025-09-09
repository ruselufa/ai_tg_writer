# 🎮 Управление приложением AI TG Writer

## 🚀 Запуск приложения

### Способ 1: Использование скриптов (рекомендуется)
```bash
# Запуск
./start_app.sh

# Остановка
./stop_app.sh
```

### Способ 2: Ручной запуск
```bash
# Запуск
go run ./cmd/ai_tg_writer

# Остановка (Ctrl+C)
# Или в другом терминале:
pkill -f "go run.*ai_tg_writer"
```

## 🔍 Проверка статуса

### Проверка запущенных процессов
```bash
ps aux | grep ai_tg_writer | grep -v grep
```

### Проверка занятости порта 8080
```bash
lsof -i :8080
```

### Проверка метрик
```bash
curl http://localhost:8080/metrics
```

## 🧪 Тестирование метрик

### Генерация тестовых метрик
```bash
go run test_metrics_simple.go
```

### Проверка конкретных метрик
```bash
# Голосовые сообщения
curl -s http://localhost:8080/metrics | grep voice_messages

# Платежи
curl -s http://localhost:8080/metrics | grep payment

# Ошибки
curl -s http://localhost:8080/metrics | grep errors

# Telegram активность
curl -s http://localhost:8080/metrics | grep telegram
```

## 📊 Мониторинг

### Prometheus
- URL: http://localhost:9090
- Метрики: http://localhost:8080/metrics

### Grafana
- URL: http://localhost:3000
- Логин: admin
- Пароль: admin

### Jaeger (трейсинг)
- URL: http://localhost:16686

## 🛠️ Устранение проблем

### Порт 8080 занят
```bash
# Найти процесс
lsof -i :8080

# Остановить процесс
kill -9 <PID>

# Или принудительно освободить порт
lsof -ti :8080 | xargs kill -9
```

### Приложение не запускается
1. Проверьте файл .env
2. Проверьте подключение к базе данных
3. Проверьте токен Telegram бота
4. Посмотрите логи приложения

### Метрики не отображаются
1. Убедитесь, что приложение запущено
2. Проверьте, что Prometheus работает
3. Запустите тестовые метрики: `go run test_metrics_simple.go`

## 📝 Полезные команды

```bash
# Очистка всех процессов ai_tg_writer
pkill -f ai_tg_writer

# Проверка всех Go процессов
ps aux | grep go | grep -v grep

# Просмотр логов в реальном времени
tail -f /var/log/ai_tg_writer.log  # если настроено логирование в файл

# Перезапуск мониторинга
docker-compose restart prometheus grafana jaeger
```

## ⚠️ Важные замечания

1. **Не используйте `&` в конце команды запуска** - это запускает процесс в фоне и может вызвать проблемы
2. **Всегда останавливайте приложение перед повторным запуском**
3. **Проверяйте свободность порта 8080 перед запуском**
4. **Используйте скрипты start_app.sh и stop_app.sh для удобства**
