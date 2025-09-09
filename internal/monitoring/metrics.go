package monitoring

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP метрики
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status_code"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// Бизнес метрики
	voiceMessagesProcessed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "voice_messages_processed_total",
			Help: "Total number of voice messages processed",
		},
		[]string{"status", "user_tariff"},
	)

	voiceProcessingDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "voice_processing_duration_seconds",
			Help:    "Voice message processing duration in seconds",
			Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30, 60},
		},
		[]string{"stage"}, // whisper, deepseek, total
	)

	activeUsers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_users_total",
			Help: "Number of active users",
		},
	)

	subscriptionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "subscriptions_total",
			Help: "Total number of subscriptions",
		},
		[]string{"action", "tariff"}, // created, cancelled, renewed
	)

	// Системные метрики
	databaseConnections = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "database_connections_active",
			Help: "Number of active database connections",
		},
		[]string{"state"}, // idle, active
	)

	externalAPICalls = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "external_api_calls_total",
			Help: "Total number of external API calls",
		},
		[]string{"service", "status"}, // whisper, deepseek, yookassa
	)

	// Telegram бот метрики
	telegramMessagesReceived = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "telegram_messages_received_total",
			Help: "Total number of Telegram messages received",
		},
		[]string{"message_type", "user_tariff"}, // voice, text, command
	)

	telegramMessagesSent = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "telegram_messages_sent_total",
			Help: "Total number of Telegram messages sent",
		},
		[]string{"message_type"}, // response, error, notification
	)

	// Активные пользователи (текущие сессии)
	activeTelegramUsers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "telegram_active_users",
			Help: "Number of currently active Telegram users",
		},
	)

	// DeepSeek токены
	deepseekTokensUsed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "deepseek_tokens_used_total",
			Help: "Total number of DeepSeek tokens used",
		},
		[]string{"type"}, // input, output, total
	)

	// Платежи
	paymentsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "payments_total",
			Help: "Total number of payments",
		},
		[]string{"status", "provider"}, // success, failed, pending; yookassa, prodamus
	)

	paymentAmount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "payment_amount_total",
			Help: "Total payment amount in rubles",
		},
		[]string{"status", "provider"},
	)

	paymentDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "payment_duration_seconds",
			Help:    "Payment processing duration in seconds",
			Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30, 60},
		},
		[]string{"provider"},
	)

	// Ошибки
	errorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "errors_total",
			Help: "Total number of errors",
		},
		[]string{"type", "component"}, // payment, voice, api; yookassa, whisper, deepseek
	)
)

// HTTP метрики
func RecordHTTPRequest(method, endpoint, statusCode string, duration time.Duration) {
	httpRequestsTotal.WithLabelValues(method, endpoint, statusCode).Inc()
	httpRequestDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())
}

// Голосовые сообщения
func RecordVoiceMessageProcessed(status, userTariff string) {
	voiceMessagesProcessed.WithLabelValues(status, userTariff).Inc()
}

func RecordVoiceProcessingDuration(stage string, duration time.Duration) {
	voiceProcessingDuration.WithLabelValues(stage).Observe(duration.Seconds())
}

// Пользователи
func SetActiveUsers(count int) {
	activeUsers.Set(float64(count))
}

// Подписки
func RecordSubscription(action, tariff string) {
	subscriptionsTotal.WithLabelValues(action, tariff).Inc()
}

// База данных
func SetDatabaseConnections(state string, count int) {
	databaseConnections.WithLabelValues(state).Set(float64(count))
}

// Внешние API
func RecordExternalAPICall(service, status string) {
	externalAPICalls.WithLabelValues(service, status).Inc()
}

// Telegram бот
func RecordTelegramMessageReceived(messageType, userTariff string) {
	telegramMessagesReceived.WithLabelValues(messageType, userTariff).Inc()
}

func RecordTelegramMessageSent(messageType string) {
	telegramMessagesSent.WithLabelValues(messageType).Inc()
}

// Активные пользователи Telegram
func SetActiveTelegramUsers(count int) {
	activeTelegramUsers.Set(float64(count))
}

// DeepSeek токены
func RecordDeepSeekTokens(tokenType string, count int) {
	deepseekTokensUsed.WithLabelValues(tokenType).Add(float64(count))
}

// Платежи
func RecordPayment(status, provider string, amount float64, duration time.Duration) {
	paymentsTotal.WithLabelValues(status, provider).Inc()
	paymentAmount.WithLabelValues(status, provider).Add(amount)
	paymentDuration.WithLabelValues(provider).Observe(duration.Seconds())
}

// Ошибки
func RecordError(errorType, component string) {
	errorsTotal.WithLabelValues(errorType, component).Inc()
}

// ===== БИЗНЕС-МЕТРИКИ =====

// Конверсия пользователей
var (
	userConversions = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_conversions_total",
			Help: "Total number of user conversions",
		},
		[]string{"from_tariff", "to_tariff"}, // free -> basic, basic -> premium
	)

	// Retention Rate
	userRetention = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "user_retention_rate",
			Help: "User retention rate by period",
		},
		[]string{"period"}, // daily, weekly, monthly
	)

	// ARPU (Average Revenue Per User)
	userARPU = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "user_arpu",
			Help: "Average Revenue Per User",
		},
		[]string{"period"}, // daily, weekly, monthly
	)

	// Churn Rate
	userChurnRate = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "user_churn_rate",
			Help: "User churn rate by period",
		},
		[]string{"period"}, // daily, weekly, monthly
	)

	// Session Duration
	userSessionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "user_session_duration_seconds",
			Help:    "User session duration in seconds",
			Buckets: []float64{60, 300, 600, 1800, 3600, 7200, 14400}, // 1min, 5min, 10min, 30min, 1h, 2h, 4h
		},
		[]string{"user_tariff"},
	)
)

// ===== МЕТРИКИ КАЧЕСТВА =====

var (
	// Voice Message Success Rate
	voiceMessageSuccessRate = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "voice_message_success_rate",
			Help: "Voice message processing success rate",
		},
		[]string{"user_tariff"},
	)

	// API Response Time Percentiles
	apiResponseTime = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "api_response_time_seconds",
			Help:    "API response time in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"endpoint", "method"},
	)

	// Error Rate by Endpoint
	endpointErrorRate = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "endpoint_error_rate",
			Help: "Error rate by endpoint",
		},
		[]string{"endpoint", "method"},
	)

	// Queue Length
	queueLength = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "queue_length",
			Help: "Current queue length",
		},
		[]string{"queue_type"}, // voice_processing, payment_processing
	)

	// Processing Success Rate
	processingSuccessRate = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "processing_success_rate",
			Help: "Overall processing success rate",
		},
		[]string{"process_type"}, // voice, payment, api
	)
)

// ===== МЕТРИКИ ВНЕШНИХ СЕРВИСОВ =====

var (
	// External API Latency
	externalAPILatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "external_api_latency_seconds",
			Help:    "External API latency in seconds",
			Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30, 60, 120},
		},
		[]string{"service", "endpoint"}, // whisper, deepseek, yookassa
	)

	// External API Success Rate
	externalAPISuccessRate = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "external_api_success_rate",
			Help: "External API success rate",
		},
		[]string{"service", "endpoint"},
	)

	// External API Quota Usage
	externalAPIQuotaUsage = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "external_api_quota_usage",
			Help: "External API quota usage percentage",
		},
		[]string{"service", "quota_type"}, // whisper, deepseek, yookassa; requests, tokens, calls
	)

	// Third-party Service Health
	externalServiceHealth = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "external_service_health",
			Help: "External service health status (1=healthy, 0=unhealthy)",
		},
		[]string{"service"},
	)
)

// ===== МЕТРИКИ ПРОИЗВОДИТЕЛЬНОСТИ =====

var (
	// Goroutine Count
	goroutineCount = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "goroutine_count",
			Help: "Current number of goroutines",
		},
	)

	// Memory Allocations
	memoryAllocations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "memory_allocations_total",
			Help: "Total number of memory allocations",
		},
		[]string{"size_class"}, // small, medium, large
	)

	// GC Pause Duration
	gcPauseDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "gc_pause_duration_seconds",
			Help:    "GC pause duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		},
		[]string{"gc_type"}, // minor, major
	)

	// Channel Buffer Usage
	channelBufferUsage = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "channel_buffer_usage",
			Help: "Channel buffer usage percentage",
		},
		[]string{"channel_name"},
	)

	// Worker Pool Utilization
	workerPoolUtilization = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "worker_pool_utilization",
			Help: "Worker pool utilization percentage",
		},
		[]string{"pool_name"},
	)
)

// ===== МЕТРИКИ ПОЛЬЗОВАТЕЛЬСКОГО ОПЫТА =====

var (
	// User Journey Steps
	userJourneySteps = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_journey_steps_total",
			Help: "Total number of user journey steps",
		},
		[]string{"step", "user_tariff"}, // start, voice_sent, processed, paid, completed
	)

	// Feature Usage
	featureUsage = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "feature_usage_total",
			Help: "Total number of feature usage",
		},
		[]string{"feature", "user_tariff"}, // voice_processing, payment, subscription, support
	)

	// User Satisfaction Score
	userSatisfactionScore = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "user_satisfaction_score",
			Help: "User satisfaction score (1-5)",
		},
		[]string{"user_tariff", "feature"},
	)

	// Support Ticket Volume
	supportTicketVolume = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "support_ticket_volume_total",
			Help: "Total number of support tickets",
		},
		[]string{"category", "priority"}, // technical, billing, feature; low, medium, high
	)

	// User Feedback Sentiment
	userFeedbackSentiment = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "user_feedback_sentiment",
			Help: "User feedback sentiment score (-1 to 1)",
		},
		[]string{"feature", "user_tariff"},
	)
)

// ===== МЕТРИКИ ИСПОЛЬЗОВАНИЯ БОТА (БЕЗ НАКОПЛЕНИЯ) =====

var (
	// Bot Usage Rate - использование бота по тарифам (rate, не накапливается)
	botUsageRate = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "bot_usage_rate",
			Help: "Bot usage rate per minute by tariff",
		},
		[]string{"user_tariff", "action_type"}, // free, premium; message, voice, button_click
	)

	// Feature Usage Rate - использование функций по тарифам (rate, не накапливается)
	featureUsageRate = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "feature_usage_rate",
			Help: "Feature usage rate per minute by tariff",
		},
		[]string{"user_tariff", "feature"}, // free, premium; voice_processing, payment, subscription
	)

	// User Engagement Rate - вовлеченность пользователей (rate, не накапливается)
	userEngagementRate = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "user_engagement_rate",
			Help: "User engagement rate per minute by tariff",
		},
		[]string{"user_tariff", "engagement_type"}, // free, premium; active, returning, new
	)

	// Bot Response Time by Tariff - время ответа бота по тарифам
	botResponseTimeByTariff = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "bot_response_time_seconds_by_tariff",
			Help:    "Bot response time in seconds by tariff",
			Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30},
		},
		[]string{"user_tariff", "message_type"}, // free, premium; text, voice, button
	)

	// Bot Error Rate by Tariff - частота ошибок бота по тарифам
	botErrorRateByTariff = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "bot_error_rate_by_tariff",
			Help: "Bot error rate by tariff",
		},
		[]string{"user_tariff", "error_type"}, // free, premium; api_error, processing_error, timeout
	)
)

// ===== МЕТРИКИ ЛОГИРОВАНИЯ =====

var (
	// Log Messages - количество логов по уровням
	logMessages = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "log_messages_total",
			Help: "Total number of log messages by level",
		},
		[]string{"level", "component"}, // debug, info, warn, error; bot, voice, api, payment
	)

	// User Interactions - взаимодействия пользователей
	userInteractions = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_interactions_total",
			Help: "Total number of user interactions",
		},
		[]string{"type", "user_tariff"}, // message, voice, button, callback; free, premium
	)

	// Processing Steps - шаги обработки
	processingSteps = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "processing_steps_total",
			Help: "Total number of processing steps",
		},
		[]string{"step", "status"}, // start, whisper, deepseek, complete; success, error
	)
)

// ===== ФУНКЦИИ ДЛЯ БИЗНЕС-МЕТРИК =====

func RecordUserConversion(fromTariff, toTariff string) {
	userConversions.WithLabelValues(fromTariff, toTariff).Inc()
}

func SetUserRetentionRate(period string, rate float64) {
	userRetention.WithLabelValues(period).Set(rate)
}

func SetUserARPU(period string, arpu float64) {
	userARPU.WithLabelValues(period).Set(arpu)
}

func SetUserChurnRate(period string, rate float64) {
	userChurnRate.WithLabelValues(period).Set(rate)
}

func RecordUserSessionDuration(userTariff string, duration time.Duration) {
	userSessionDuration.WithLabelValues(userTariff).Observe(duration.Seconds())
}

// ===== ФУНКЦИИ ДЛЯ МЕТРИК КАЧЕСТВА =====

func SetVoiceMessageSuccessRate(userTariff string, rate float64) {
	voiceMessageSuccessRate.WithLabelValues(userTariff).Set(rate)
}

func RecordAPIResponseTime(endpoint, method string, duration time.Duration) {
	apiResponseTime.WithLabelValues(endpoint, method).Observe(duration.Seconds())
}

func SetEndpointErrorRate(endpoint, method string, rate float64) {
	endpointErrorRate.WithLabelValues(endpoint, method).Set(rate)
}

func SetQueueLength(queueType string, length float64) {
	queueLength.WithLabelValues(queueType).Set(length)
}

func SetProcessingSuccessRate(processType string, rate float64) {
	processingSuccessRate.WithLabelValues(processType).Set(rate)
}

// ===== ФУНКЦИИ ДЛЯ ВНЕШНИХ СЕРВИСОВ =====

func RecordExternalAPILatency(service, endpoint string, duration time.Duration) {
	externalAPILatency.WithLabelValues(service, endpoint).Observe(duration.Seconds())
}

func SetExternalAPISuccessRate(service, endpoint string, rate float64) {
	externalAPISuccessRate.WithLabelValues(service, endpoint).Set(rate)
}

func SetExternalAPIQuotaUsage(service, quotaType string, usage float64) {
	externalAPIQuotaUsage.WithLabelValues(service, quotaType).Set(usage)
}

func SetExternalServiceHealth(service string, healthy bool) {
	health := 0.0
	if healthy {
		health = 1.0
	}
	externalServiceHealth.WithLabelValues(service).Set(health)
}

// ===== ФУНКЦИИ ДЛЯ ПРОИЗВОДИТЕЛЬНОСТИ =====

func SetGoroutineCount(count int) {
	goroutineCount.Set(float64(count))
}

func RecordMemoryAllocation(sizeClass string, count int) {
	memoryAllocations.WithLabelValues(sizeClass).Add(float64(count))
}

func RecordGCPauseDuration(gcType string, duration time.Duration) {
	gcPauseDuration.WithLabelValues(gcType).Observe(duration.Seconds())
}

func SetChannelBufferUsage(channelName string, usage float64) {
	channelBufferUsage.WithLabelValues(channelName).Set(usage)
}

func SetWorkerPoolUtilization(poolName string, utilization float64) {
	workerPoolUtilization.WithLabelValues(poolName).Set(utilization)
}

// ===== ФУНКЦИИ ДЛЯ ПОЛЬЗОВАТЕЛЬСКОГО ОПЫТА =====

func RecordUserJourneyStep(step, userTariff string) {
	userJourneySteps.WithLabelValues(step, userTariff).Inc()
}

func RecordFeatureUsage(feature, userTariff string) {
	featureUsage.WithLabelValues(feature, userTariff).Inc()
}

func SetUserSatisfactionScore(userTariff, feature string, score float64) {
	userSatisfactionScore.WithLabelValues(userTariff, feature).Set(score)
}

func RecordSupportTicket(category, priority string) {
	supportTicketVolume.WithLabelValues(category, priority).Inc()
}

func SetUserFeedbackSentiment(feature, userTariff string, sentiment float64) {
	userFeedbackSentiment.WithLabelValues(feature, userTariff).Set(sentiment)
}

// ===== ФУНКЦИИ ДЛЯ МЕТРИК ИСПОЛЬЗОВАНИЯ БОТА =====

func SetBotUsageRate(userTariff, actionType string, rate float64) {
	botUsageRate.WithLabelValues(userTariff, actionType).Set(rate)
}

func SetFeatureUsageRate(userTariff, feature string, rate float64) {
	featureUsageRate.WithLabelValues(userTariff, feature).Set(rate)
}

func SetUserEngagementRate(userTariff, engagementType string, rate float64) {
	userEngagementRate.WithLabelValues(userTariff, engagementType).Set(rate)
}

func RecordBotResponseTimeByTariff(userTariff, messageType string, duration time.Duration) {
	botResponseTimeByTariff.WithLabelValues(userTariff, messageType).Observe(duration.Seconds())
}

func SetBotErrorRateByTariff(userTariff, errorType string, rate float64) {
	botErrorRateByTariff.WithLabelValues(userTariff, errorType).Set(rate)
}

// ===== ФУНКЦИИ ДЛЯ МЕТРИК ЛОГИРОВАНИЯ =====

func RecordLogMessage(level, component string) {
	logMessages.WithLabelValues(level, component).Inc()
}

func RecordUserInteraction(interactionType, userTariff string) {
	userInteractions.WithLabelValues(interactionType, userTariff).Inc()
}

func RecordProcessingStep(step, status string) {
	processingSteps.WithLabelValues(step, status).Inc()
}

// InitMetrics инициализирует все метрики
func InitMetrics() {
	// Метрики уже зарегистрированы при импорте пакета благодаря promauto
	// Эта функция нужна только для явной инициализации
}
