package api

import (
	"ai_tg_writer/internal/infrastructure/bot"
	"ai_tg_writer/internal/infrastructure/database"
	"ai_tg_writer/internal/service"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type Server struct {
	router *mux.Router
	port   string
}

func NewServer(port string) *Server {
	return &Server{
		router: mux.NewRouter(),
		port:   port,
	}
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

	// Добавляем middleware для логирования
	s.router.Use(loggingMiddleware)

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

// loggingMiddleware добавляет логирование запросов
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
		next.ServeHTTP(w, r)
	})
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
