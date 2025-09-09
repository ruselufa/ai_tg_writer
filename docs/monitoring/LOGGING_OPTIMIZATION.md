# 🔧 Оптимизация логирования

## 🎯 Проблема

В приложении было много отладочных логов в командную строку, что засоряло вывод и снижало производительность.

## ✅ Решение

### 1. **Система уровней логирования**

Создана система с 5 уровнями:

```go
// Уровни логирования
const (
    LevelDebug = "DEBUG"  // Отладочная информация
    LevelInfo  = "INFO"   // Информационные сообщения
    LevelWarn  = "WARN"   // Предупреждения
    LevelError = "ERROR"  // Ошибки
    LevelSystem = "SYSTEM" // Системные сообщения (всегда показываются)
)
```

### 2. **Функции логирования**

```go
// Debug - только при LOG_LEVEL=DEBUG
monitoring.Debug("Отладочная информация: %s", data)

// Info - информационные сообщения
monitoring.Info("Пользователь %d отправил сообщение", userID)

// Warn - предупреждения
monitoring.Warn("Недостаточно памяти для обработки")

// Error - ошибки (всегда показываются)
monitoring.Error("Ошибка обработки: %v", err)

// System - системные сообщения (всегда показываются)
monitoring.System("Приложение запущено на порту %s", port)
```

### 3. **Метрики логирования**

Добавлены метрики в Prometheus:

```go
// Количество логов по уровням и компонентам
log_messages_total{level="info", component="bot"}
log_messages_total{level="error", component="payment"}

// Взаимодействия пользователей
user_interactions_total{type="message", user_tariff="premium"}
user_interactions_total{type="voice", user_tariff="free"}

// Шаги обработки
processing_steps_total{step="whisper", status="success"}
processing_steps_total{step="deepseek", status="error"}
```

## 🔧 Настройка

### **Уровень логирования**

```bash
# По умолчанию (только INFO и выше)
export LOG_LEVEL=INFO

# Для отладки (все логи)
export LOG_LEVEL=DEBUG

# Только ошибки
export LOG_LEVEL=ERROR
```

### **В Docker Compose**

```yaml
environment:
  - LOG_LEVEL=INFO
```

## 📊 Мониторинг логов

### **В Grafana**

1. **Количество логов по уровням**:
   ```promql
   rate(log_messages_total[5m]) * 60
   ```

2. **Взаимодействия пользователей**:
   ```promql
   rate(user_interactions_total[5m]) * 60
   ```

3. **Шаги обработки**:
   ```promql
   rate(processing_steps_total[5m]) * 60
   ```

### **Алерты**

```yaml
# Много ошибок
- alert: HighErrorRate
  expr: rate(log_messages_total{level="error"}[5m]) > 0.1
  for: 2m
  labels:
    severity: warning
  annotations:
    summary: "High error rate detected"

# Много взаимодействий
- alert: HighUserActivity
  expr: rate(user_interactions_total[5m]) > 100
  for: 1m
  labels:
    severity: info
  annotations:
    summary: "High user activity detected"
```

## 🎯 Преимущества

### **1. Производительность**
- ✅ Убраны избыточные логи в продакшене
- ✅ Debug логи только при необходимости
- ✅ Системные логи всегда видны

### **2. Мониторинг**
- ✅ Метрики в Prometheus
- ✅ Дашборды в Grafana
- ✅ Алерты при проблемах

### **3. Отладка**
- ✅ Уровни логирования
- ✅ Компонентная разбивка
- ✅ Структурированные логи

## 🔄 Миграция

### **Было:**
```go
log.Printf("Callback от пользователя %d: %s", userID, data)
log.Printf("❌ Ошибка обработки: %v", err)
log.Printf("✅ Успешно обработано")
```

### **Стало:**
```go
monitoring.Debug("Callback от пользователя %d: %s", userID, data)
monitoring.RecordUserInteraction("callback", "premium")

monitoring.Error("Ошибка обработки: %v", err)
monitoring.RecordLogMessage("error", "bot")

monitoring.Info("Успешно обработано")
monitoring.RecordProcessingStep("complete", "success")
```

## 📈 Результат

### **До оптимизации:**
- 354 строки с логированием
- Много отладочной информации в продакшене
- Нет метрик по логам
- Сложно отслеживать активность

### **После оптимизации:**
- Структурированное логирование
- Уровни логирования
- Метрики в Prometheus
- Дашборды в Grafana
- Алерты при проблемах

## 🚀 Использование

### **В коде:**
```go
// Импорт
import "your-project/internal/monitoring"

// Логирование
monitoring.Info("Пользователь %d начал обработку", userID)
monitoring.RecordUserInteraction("voice", "premium")

// Ошибки
monitoring.Error("Ошибка API: %v", err)
monitoring.RecordLogMessage("error", "api")
```

### **В продакшене:**
```bash
# Только важные логи
export LOG_LEVEL=INFO

# Для отладки
export LOG_LEVEL=DEBUG
```

## 🎯 Готово!

Теперь логирование оптимизировано, производительность улучшена, а мониторинг стал более информативным! 🚀
