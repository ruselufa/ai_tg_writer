# 🚀 Расширенные метрики мониторинга

## 📊 Обзор

Добавлены **12 новых категорий метрик** для полного мониторинга AI TG Writer:

### 1. 🏢 Бизнес-метрики
- **Конверсия пользователей** - переходы между тарифами (free → basic → premium)
- **Retention Rate** - удержание пользователей (дневное, недельное, месячное)
- **ARPU** - средний доход с пользователя
- **Churn Rate** - отток пользователей
- **Session Duration** - длительность сессий по тарифам

### 2. 🎯 Метрики качества
- **Voice Message Success Rate** - успешность обработки голосовых сообщений
- **API Response Time** - время ответа API с перцентилями (50th, 95th)
- **Endpoint Error Rate** - частота ошибок по эндпоинтам
- **Queue Length** - длина очередей обработки
- **Processing Success Rate** - общая успешность обработки

### 3. 🌐 Метрики внешних сервисов
- **External API Latency** - задержка внешних API (Whisper, DeepSeek, YooKassa)
- **External API Success Rate** - процент успешных вызовов
- **External API Quota Usage** - использование квот сервисов
- **Third-party Service Health** - состояние внешних сервисов

### 4. ⚡ Метрики производительности
- **Goroutine Count** - количество горутин
- **Memory Allocations** - аллокации памяти по размерам
- **GC Pause Duration** - длительность пауз сборщика мусора
- **Channel Buffer Usage** - использование буферов каналов
- **Worker Pool Utilization** - загрузка пулов воркеров

### 5. 👤 Метрики пользовательского опыта
- **User Journey Steps** - шаги пользовательского пути
- **Feature Usage** - использование функций
- **User Satisfaction Score** - оценка удовлетворенности (1-5)
- **Support Ticket Volume** - объем обращений в поддержку
- **User Feedback Sentiment** - тональность отзывов (-1 до 1)

## 🔧 Реализация

### Новые метрики в `internal/monitoring/metrics.go`:

```go
// Бизнес-метрики
userConversions          // Конверсии пользователей
userRetention            // Retention Rate
userARPU                 // ARPU
userChurnRate            // Churn Rate
userSessionDuration      // Длительность сессий

// Метрики качества
voiceMessageSuccessRate  // Успешность голосовых сообщений
apiResponseTime          // Время ответа API
endpointErrorRate        // Частота ошибок
queueLength              // Длина очередей
processingSuccessRate    // Успешность обработки

// Внешние сервисы
externalAPILatency       // Задержка внешних API
externalAPISuccessRate   // Успешность внешних API
externalAPIQuotaUsage    // Использование квот
externalServiceHealth    // Состояние сервисов

// Производительность
goroutineCount           // Количество горутин
memoryAllocations        // Аллокации памяти
gcPauseDuration          // Паузы GC
channelBufferUsage       // Использование буферов
workerPoolUtilization    // Загрузка пулов

// Пользовательский опыт
userJourneySteps         // Шаги пользователя
featureUsage             // Использование функций
userSatisfactionScore    // Удовлетворенность
supportTicketVolume      // Обращения в поддержку
userFeedbackSentiment    // Тональность отзывов
```

### Функции для записи метрик:

```go
// Бизнес-метрики
RecordUserConversion(fromTariff, toTariff string)
SetUserRetentionRate(period string, rate float64)
SetUserARPU(period string, arpu float64)
SetUserChurnRate(period string, rate float64)
RecordUserSessionDuration(userTariff string, duration time.Duration)

// Метрики качества
SetVoiceMessageSuccessRate(userTariff string, rate float64)
RecordAPIResponseTime(endpoint, method string, duration time.Duration)
SetEndpointErrorRate(endpoint, method string, rate float64)
SetQueueLength(queueType string, length float64)
SetProcessingSuccessRate(processType string, rate float64)

// Внешние сервисы
RecordExternalAPILatency(service, endpoint string, duration time.Duration)
SetExternalAPISuccessRate(service, endpoint string, rate float64)
SetExternalAPIQuotaUsage(service, quotaType string, usage float64)
SetExternalServiceHealth(service string, healthy bool)

// Производительность
SetGoroutineCount(count int)
RecordMemoryAllocation(sizeClass string, count int)
RecordGCPauseDuration(gcType string, duration time.Duration)
SetChannelBufferUsage(channelName string, usage float64)
SetWorkerPoolUtilization(poolName string, utilization float64)

// Пользовательский опыт
RecordUserJourneyStep(step, userTariff string)
RecordFeatureUsage(feature, userTariff string)
SetUserSatisfactionScore(userTariff, feature string, score float64)
RecordSupportTicket(category, priority string)
SetUserFeedbackSentiment(feature, userTariff string, sentiment float64)
```

## 📈 Дашборд Grafana

Добавлены **12 новых панелей** в дашборд:

1. **Business Metrics - Conversions** - конверсии пользователей
2. **Business Metrics - Retention & ARPU** - удержание и ARPU
3. **Quality Metrics - Success Rates** - показатели успешности
4. **Quality Metrics - API Response Time** - время ответа API
5. **External Services - Latency** - задержка внешних сервисов
6. **External Services - Success Rate & Health** - успешность и состояние
7. **Performance - Goroutines & Memory** - горутины и память
8. **Performance - GC & Worker Pools** - GC и пулы воркеров
9. **User Experience - Journey Steps** - шаги пользователя
10. **User Experience - Satisfaction & Feedback** - удовлетворенность
11. **Support & Queue Metrics** - поддержка и очереди
12. **External API Quota Usage** - использование квот

## 🧪 Тестирование

Обновлен endpoint `/test-metrics` для генерации всех новых метрик:

```bash
curl http://localhost:8080/test-metrics
```

Генерирует тестовые данные для всех категорий метрик.

## 🎯 Применение в коде

### Примеры интеграции:

```go
// В voice_handler.go
monitoring.RecordUserJourneyStep("voice_sent", userTariff)
monitoring.RecordFeatureUsage("voice_processing", userTariff)
monitoring.SetVoiceMessageSuccessRate(userTariff, successRate)

// В yookassa_handler.go
monitoring.RecordUserConversion("free", "basic")
monitoring.RecordSupportTicket("billing", "medium")
monitoring.SetUserSatisfactionScore("basic", "payment", 4.2)

// В deepseek_handler.go
monitoring.RecordExternalAPILatency("deepseek", "chat", duration)
monitoring.SetExternalAPISuccessRate("deepseek", "chat", successRate)
monitoring.SetExternalAPIQuotaUsage("deepseek", "tokens", quotaUsage)
```

## 📊 Мониторинг в реальном времени

Все метрики доступны в Prometheus:
- **Endpoint**: `http://localhost:8080/metrics`
- **Grafana**: `http://localhost:3000` (admin/admin)
- **Prometheus**: `http://localhost:9090`

## 🔍 Полезные запросы Prometheus

```promql
# Конверсии пользователей
rate(user_conversions_total[5m]) * 60

# Retention Rate
user_retention_rate

# ARPU
user_arpu

# Успешность голосовых сообщений
voice_message_success_rate

# Время ответа API (95-й перцентиль)
histogram_quantile(0.95, rate(api_response_time_seconds_bucket[5m]))

# Задержка внешних API
histogram_quantile(0.95, rate(external_api_latency_seconds_bucket[5m]))

# Количество горутин
goroutine_count

# Шаги пользовательского пути
rate(user_journey_steps_total[5m]) * 60

# Удовлетворенность пользователей
user_satisfaction_score
```

## 🚀 Следующие шаги

1. **Интеграция в реальный код** - добавить вызовы метрик в обработчики
2. **Автоматический сбор** - настроить периодический сбор системных метрик
3. **Алерты** - настроить уведомления при критических значениях
4. **Аналитика** - создать дополнительные дашборды для анализа
5. **A/B тестирование** - использовать метрики для сравнения версий

## 📝 Примечания

- Все метрики используют **rate()** для отображения активности в минуту
- **Histogram** метрики автоматически создают перцентили
- **Gauge** метрики показывают текущее состояние
- **Counter** метрики накапливают значения
- Все метрики имеют соответствующие лейблы для детализации
