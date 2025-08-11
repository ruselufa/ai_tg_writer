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
	log.Printf("=== YooKassa Webhook Received ===")
	log.Printf("Event: %s", evt.Event)
	log.Printf("Object: %+v", evt.Object)

	// Получим платеж, чтобы подтвердить статус
	id, _ := evt.Object["id"].(string)
	if id == "" {
		log.Printf("❌ Payment ID not found in webhook object")
		w.WriteHeader(http.StatusOK)
		return
	}
	log.Printf("Payment ID from webhook: %s", id)

	payment, err := h.yc.GetPayment(id)
	if err != nil {
		log.Printf("❌ YK get payment err: %v", err)
		w.WriteHeader(http.StatusOK)
		return
	}
	log.Printf("=== Payment Details from YooKassa API ===")
	log.Printf("Payment ID: %s", id)
	log.Printf("Full payment response: %+v", payment)

	status, _ := payment["status"].(string)
	log.Printf("Payment Status: %s", status)

	if status == "succeeded" {
		log.Printf("✅ Payment succeeded, processing...")

		// Метаданные и пользователь
		meta, _ := payment["metadata"].(map[string]any)
		tgUser := ""
		if meta != nil {
			log.Printf("Metadata: %+v", meta)
			if v, ok := meta["tg_user_id"].(string); ok {
				tgUser = v
				log.Printf("TG User ID from metadata: %s", tgUser)
			}
		}

		// payment_method.id
		pm := ""
		if pmObj, ok := payment["payment_method"].(map[string]any); ok {
			log.Printf("Payment Method: %+v", pmObj)
			if v, ok := pmObj["id"].(string); ok {
				pm = v
				log.Printf("Payment Method ID: %s", pm)
			}
		}

		// customer.id
		cust := ""
		if custObj, ok := payment["customer"].(map[string]any); ok {
			log.Printf("Customer: %+v", custObj)
			if v, ok := custObj["id"].(string); ok {
				cust = v
				log.Printf("Customer ID: %s", cust)
			}
		}

		// Сохраняем привязку (если есть все данные)
		if tgUser != "" && pm != "" && cust != "" {
			log.Printf("✅ All required data found, saving binding...")

			// Пробуем извлечь сумму
			amountValue := 0.0
			if amt, ok := payment["amount"].(map[string]any); ok {
				log.Printf("Amount: %+v", amt)
				if val, ok := amt["value"].(string); ok {
					if f, err := strconv.ParseFloat(val, 64); err == nil {
						amountValue = f
						log.Printf("Parsed amount: %.2f", amountValue)
					}
				}
			}

			if uid, err := strconv.ParseInt(tgUser, 10, 64); err == nil {
				log.Printf("Parsed User ID: %d", uid)
				if err := h.subs.SaveYooKassaBindingAndActivate(uid, cust, pm, id, amountValue); err != nil {
					log.Printf("❌ Save binding error: %v", err)
				} else {
					log.Printf("✅ Binding saved and subscription activated successfully")
				}
			} else {
				log.Printf("❌ Failed to parse TG User ID: %v", err)
			}
		} else {
			log.Printf("❌ Missing required data: tgUser=%s, pm=%s, cust=%s", tgUser, pm, cust)
		}
	} else {
		log.Printf("⚠️ Payment status is not 'succeeded': %s", status)
	}

	log.Printf("=== End Webhook Processing ===")
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
