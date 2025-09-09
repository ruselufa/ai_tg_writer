package monitoring

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"
)

type HealthChecker struct {
	db *sql.DB
}

type HealthStatus struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Services  map[string]string `json:"services"`
}

func NewHealthChecker(db *sql.DB) *HealthChecker {
	return &HealthChecker{db: db}
}

func (h *HealthChecker) CheckHealth(ctx context.Context) *HealthStatus {
	status := &HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now(),
		Services:  make(map[string]string),
	}

	// Проверяем базу данных
	if err := h.checkDatabase(ctx); err != nil {
		status.Status = "unhealthy"
		status.Services["database"] = fmt.Sprintf("error: %v", err)
	} else {
		status.Services["database"] = "ok"
	}

	// Проверяем внешние сервисы
	if err := h.checkExternalServices(ctx); err != nil {
		status.Status = "degraded"
		status.Services["external_apis"] = fmt.Sprintf("warning: %v", err)
	} else {
		status.Services["external_apis"] = "ok"
	}

	return status
}

func (h *HealthChecker) checkDatabase(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return h.db.PingContext(ctx)
}

func (h *HealthChecker) checkExternalServices(ctx context.Context) error {
	// Проверяем доступность внешних API
	// Здесь можно добавить проверки для Whisper, DeepSeek и других API
	// Пока возвращаем nil для простоты
	return nil
}

func (h *HealthChecker) HealthHandler(w http.ResponseWriter, r *http.Request) {
	status := h.CheckHealth(r.Context())

	w.Header().Set("Content-Type", "application/json")

	if status.Status == "healthy" {
		w.WriteHeader(http.StatusOK)
	} else if status.Status == "degraded" {
		w.WriteHeader(http.StatusOK) // 200, но с предупреждением
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	// Здесь можно добавить JSON encoding
	fmt.Fprintf(w, `{"status":"%s","timestamp":"%s","services":%v}`,
		status.Status, status.Timestamp.Format(time.RFC3339), status.Services)
}
