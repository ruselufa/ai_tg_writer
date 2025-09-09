package api

import (
	"ai_tg_writer/internal/infrastructure/bot"
	"ai_tg_writer/internal/infrastructure/database"
	"ai_tg_writer/internal/monitoring"
	"ai_tg_writer/internal/service"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type Server struct {
	router        *mux.Router
	port          string
	healthChecker *monitoring.HealthChecker
}

func NewServer(port string) *Server {
	return &Server{
		router: mux.NewRouter(),
		port:   port,
	}
}

func (s *Server) AddHealthCheck(healthChecker *monitoring.HealthChecker) {
	s.healthChecker = healthChecker
	s.router.HandleFunc("/health", s.healthChecker.HealthHandler).Methods("GET")
}

// SetupRoutes настраивает все маршруты сервера
func (s *Server) SetupRoutes(
	subscriptionService *service.SubscriptionService,
	prodamusHandler interface{},
	db *database.DB,
	bot *bot.Bot,
) {
	// Создаем обработчик платежей
	paymentHandler := NewPaymentHandler(subscriptionService, prodamusHandler, db)

	// Настраиваем маршруты для платежей
	paymentHandler.SetupRoutes(s.router)

	// Тестовый эндпоинт для проверки доступности через localtunnel
	s.router.HandleFunc("/ping", s.handlePing).Methods("GET")

	// Тестовый эндпоинт для генерации метрик
	s.router.HandleFunc("/test-metrics", s.handleTestMetrics).Methods("GET")

	// Добавляем эндпоинт для метрик Prometheus
	s.router.Handle("/metrics", promhttp.Handler()).Methods("GET")

	// Добавляем middleware для мониторинга
	s.router.Use(monitoringMiddleware)
	s.router.Use(otelhttp.NewMiddleware("ai_tg_writer"))

	yk := NewYooKassaHandler(subscriptionService, db, bot)
	yk.SetupRoutes(s.router)
}

// handlePing — простой эндпоинт для проверки доступности сервера
func (s *Server) handlePing(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	response := map[string]any{
		"status": "ok",
		"ts":     time.Now().UTC().Format(time.RFC3339),
		"method": r.Method,
		"path":   r.URL.Path,
	}
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		response["forwarded_for"] = fwd
	}
	if ua := r.Header.Get("User-Agent"); ua != "" {
		response["user_agent"] = ua
	}
	_ = json.NewEncoder(w).Encode(response)
}

// handleTestMetrics — тестовый эндпоинт для генерации метрик
func (s *Server) handleTestMetrics(w http.ResponseWriter, r *http.Request) {
	// Генерируем тестовые метрики
	monitoring.RecordVoiceMessageProcessed("success", "premium")
	monitoring.RecordVoiceMessageProcessed("success", "basic")
	monitoring.RecordVoiceProcessingDuration("whisper", 2*time.Second)
	monitoring.RecordVoiceProcessingDuration("total", 3*time.Second)

	monitoring.RecordTelegramMessageReceived("voice", "premium")
	monitoring.RecordTelegramMessageReceived("text", "basic")
	monitoring.RecordTelegramMessageSent("response")

	monitoring.RecordDeepSeekTokens("input", 100)
	monitoring.RecordDeepSeekTokens("output", 50)
	monitoring.RecordDeepSeekTokens("total", 150)

	monitoring.RecordPayment("success", "yookassa", 990.0, 2*time.Second)
	monitoring.RecordPayment("pending", "yookassa", 1990.0, 1*time.Second)
	monitoring.RecordPayment("failed", "prodamus", 500.0, 3*time.Second)

	monitoring.RecordError("payment", "yookassa")
	monitoring.RecordError("api", "deepseek")
	monitoring.RecordError("voice", "whisper")

	monitoring.MarkUserActiveGlobal(123)
	monitoring.MarkUserActiveGlobal(456)

	monitoring.RecordExternalAPICall("whisper", "success")
	monitoring.RecordExternalAPICall("deepseek", "success")
	monitoring.RecordExternalAPICall("yookassa", "error")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "success", "message": "Test metrics generated"}`))
}

// monitoringMiddleware добавляет метрики и логирование
func monitoringMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Создаем контекст с trace_id
		ctx := context.WithValue(r.Context(), "trace_id", generateTraceID())
		r = r.WithContext(ctx)

		// Логируем запрос
		logger := monitoring.NewLogger()
		logger.WithContext(ctx).WithFields(map[string]interface{}{
			"method": r.Method,
			"path":   r.URL.Path,
			"ip":     r.RemoteAddr,
		}).Info("HTTP request")

		// Обертываем ResponseWriter для отслеживания статуса
		wrapped := &responseWriter{ResponseWriter: w, statusCode: 200}

		next.ServeHTTP(wrapped, r)

		// Записываем метрики
		duration := time.Since(start)
		monitoring.RecordHTTPRequest(r.Method, r.URL.Path,
			strconv.Itoa(wrapped.statusCode), duration)

		logger.WithContext(ctx).WithFields(map[string]interface{}{
			"status_code": wrapped.statusCode,
			"duration":    duration.String(),
		}).Info("HTTP response")
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func generateTraceID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// Start запускает HTTP-сервер
func (s *Server) Start() error {
	log.Printf("HTTP-сервер запущен на порту %s", s.port)
	return http.ListenAndServe(":"+s.port, s.router)
}

// GetPort возвращает порт сервера
func (s *Server) GetPort() string {
	return s.port
}
