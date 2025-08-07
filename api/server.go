package api

import (
	"ai_tg_writer/internal/infrastructure/database"
	"ai_tg_writer/internal/service"
	"log"
	"net/http"

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
) {
	// Создаем обработчик платежей
	paymentHandler := NewPaymentHandler(subscriptionService, prodamusHandler, db)
	
	// Настраиваем маршруты для платежей
	paymentHandler.SetupRoutes(s.router)
	
	// Добавляем middleware для логирования
	s.router.Use(loggingMiddleware)
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