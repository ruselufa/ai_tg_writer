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

