# 🚀 Быстрый старт - Оптимизированное логирование

## 🎯 Что изменилось

### **До:**
- 354 строки с `log.Printf()` в командную строку
- Много отладочной информации в продакшене
- Нет метрик по логам

### **После:**
- Система уровней логирования
- Метрики в Prometheus
- Дашборды в Grafana
- Оптимизированная производительность

## 🔧 Быстрая настройка

### **1. Уровень логирования**

```bash
# По умолчанию (только важные логи)
export LOG_LEVEL=INFO

# Для отладки (все логи)
export LOG_LEVEL=DEBUG

# Только ошибки
export LOG_LEVEL=ERROR
```

### **2. В Docker Compose**

```yaml
environment:
  - LOG_LEVEL=INFO
```

## 📊 Новые метрики

### **В Prometheus:**
- `log_messages_total` - количество логов по уровням
- `user_interactions_total` - взаимодействия пользователей
- `processing_steps_total` - шаги обработки

### **В Grafana:**
- **Log Messages by Level** - логи по уровням
- **User Interactions** - взаимодействия пользователей
- **Processing Steps** - шаги обработки

## 🎯 Использование в коде

### **Импорт:**
```go
import "your-project/internal/monitoring"
```

### **Логирование:**
```go
// Отладочная информация (только при DEBUG)
monitoring.Debug("Отладочная информация: %s", data)

// Информационные сообщения
monitoring.Info("Пользователь %d отправил сообщение", userID)

// Предупреждения
monitoring.Warn("Недостаточно памяти")

// Ошибки (всегда показываются)
monitoring.Error("Ошибка обработки: %v", err)

// Системные сообщения (всегда показываются)
monitoring.System("Приложение запущено")
```

### **Метрики:**
```go
// Запись метрик
monitoring.RecordLogMessage("info", "bot")
monitoring.RecordUserInteraction("voice", "premium")
monitoring.RecordProcessingStep("whisper", "success")
```

## 🚀 Запуск

### **Разработка:**
```bash
# Все логи
export LOG_LEVEL=DEBUG
./start_app.sh

# Только важные
export LOG_LEVEL=INFO
./start_app.sh
```

### **Продакшн:**
```bash
# Только ошибки и системные
export LOG_LEVEL=ERROR
./start_app.sh
```

## 📈 Мониторинг

### **Grafana дашборды:**
- http://localhost:3000 - основной дашборд
- http://localhost:3000/grafana/ - мониторинг

### **Полезные запросы:**
```promql
# Количество ошибок в минуту
rate(log_messages_total{level="error"}[5m]) * 60

# Взаимодействия пользователей
rate(user_interactions_total[5m]) * 60

# Успешные шаги обработки
rate(processing_steps_total{status="success"}[5m]) * 60
```

## 🎯 Результат

- ✅ **Производительность** - убраны избыточные логи
- ✅ **Мониторинг** - метрики в Prometheus
- ✅ **Отладка** - уровни логирования
- ✅ **Алерты** - уведомления при проблемах

## 🔧 Troubleshooting

### **Если не видно логов:**
```bash
# Проверьте уровень
echo $LOG_LEVEL

# Установите DEBUG для отладки
export LOG_LEVEL=DEBUG
```

### **Если не работают метрики:**
```bash
# Проверьте Prometheus
curl http://localhost:9090/api/v1/targets

# Проверьте метрики
curl http://localhost:8080/metrics | grep log_messages
```

## 🎉 Готово!

Теперь логирование оптимизировано и интегрировано с мониторингом! 🚀
