package api

import (
	"ai_tg_writer/internal/infrastructure/database"
	"ai_tg_writer/internal/infrastructure/yookassa"
	"ai_tg_writer/internal/service"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type YooKassaHandler struct {
	subs *service.SubscriptionService
	db   *database.DB
	yc   *yookassa.Client
}

func NewYooKassaHandler(subs *service.SubscriptionService, db *database.DB) *YooKassaHandler {
	return &YooKassaHandler{subs: subs, db: db, yc: yookassa.New()}
}

// 6.1 Создать первичный платеж для привязки карты
// POST /yookassa/init?user_id=123&amount=990.00
func (h *YooKassaHandler) CreateInit(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	amount := r.URL.Query().Get("amount")
	if userID == "" || amount == "" {
		http.Error(w, "user_id and amount required", http.StatusBadRequest)
		return
	}
	idem := time.Now().UTC().Format("20060102T150405.000000000Z") + "-" + userID
	retURL := os.Getenv("YK_RETURN_URL_BASE")
	payment, err := h.yc.CreateInitialPayment(
		idem,
		yookassa.Amount{Value: amount, Currency: "RUB"},
		"Подписка AI TG Writer",
		userID,
		retURL,
		map[string]string{"tg_user_id": userID},
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payment)
}

// 6.2 Вебхук
// URL: POST /yookassa/webhook
func (h *YooKassaHandler) Webhook(w http.ResponseWriter, r *http.Request) {
	var evt struct {
		Event  string         `json:"event"`
		Object map[string]any `json:"object"`
	}
	if err := json.NewDecoder(r.Body).Decode(&evt); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	log.Printf("YK event: %s", evt.Event)

	// Получим платеж, чтобы подтвердить статус
	id, _ := evt.Object["id"].(string)
	if id == "" {
		w.WriteHeader(http.StatusOK)
		return
	}
	payment, err := h.yc.GetPayment(id)
	if err != nil {
		log.Printf("YK get payment err: %v", err)
		w.WriteHeader(http.StatusOK)
		return
	}

	status, _ := payment["status"].(string)
	if status == "succeeded" {
		// Метаданные и пользователь
		meta, _ := payment["metadata"].(map[string]any)
		tgUser := ""
		if meta != nil {
			if v, ok := meta["tg_user_id"].(string); ok {
				tgUser = v
			}
		}
		// payment_method.id
		pm := ""
		if pmObj, ok := payment["payment_method"].(map[string]any); ok {
			if v, ok := pmObj["id"].(string); ok {
				pm = v
			}
		}
		// customer.id
		cust := ""
		if custObj, ok := payment["customer"].(map[string]any); ok {
			if v, ok := custObj["id"].(string); ok {
				cust = v
			}
		}
		// Сохраняем привязку (если есть все данные)
		if tgUser != "" && pm != "" && cust != "" {
			// Пробуем извлечь сумму
			amountValue := 0.0
			if amt, ok := payment["amount"].(map[string]any); ok {
				if val, ok := amt["value"].(string); ok {
					if f, err := strconv.ParseFloat(val, 64); err == nil {
						amountValue = f
					}
				}
			}

			if uid, err := strconv.ParseInt(tgUser, 10, 64); err == nil {
				if err := h.subs.SaveYooKassaBindingAndActivate(uid, cust, pm, id, amountValue); err != nil {
					log.Printf("save binding error: %v", err)
				}
			}
		}
	}
	w.Write([]byte("ok"))
}

// 6.3 Принудительное списание (тест рекуррента)
// POST /yookassa/charge?user_id=123&amount=990.00
func (h *YooKassaHandler) Charge(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	amount := r.URL.Query().Get("amount")
	if userID == "" || amount == "" {
		http.Error(w, "user_id and amount required", http.StatusBadRequest)
		return
	}
	// достать из БД yk_customer_id, yk_payment_method_id
	// ...
	idem := time.Now().UTC().Format("20060102T150405.000000000Z") + "-rec-" + userID
	payment, err := h.yc.CreateRecurringPayment(
		idem,
		yookassa.Amount{Value: amount, Currency: "RUB"},
		"Продление подписки AI TG Writer",
		/*customerID*/ userID,
		/*paymentMethodID*/ "pm_xxx",
		map[string]string{"tg_user_id": userID},
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payment)
}

func (h *YooKassaHandler) SetupRoutes(r *mux.Router) {
	r.HandleFunc("/yookassa/init", h.CreateInit).Methods("POST")
	r.HandleFunc("/yookassa/webhook", h.Webhook).Methods("POST")
	r.HandleFunc("/yookassa/charge", h.Charge).Methods("POST")
}
