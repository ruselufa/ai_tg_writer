package worker

import (
	"ai_tg_writer/internal/config"
	"ai_tg_writer/internal/service"
	"context"
	"log"
	"time"
)

type SubscriptionWorker struct {
	subscriptionService *service.SubscriptionService
	config              *config.Config
}

// NewSubscriptionWorker создает новый воркер для обработки подписок
func NewSubscriptionWorker(subscriptionService *service.SubscriptionService, config *config.Config) *SubscriptionWorker {
	return &SubscriptionWorker{
		subscriptionService: subscriptionService,
		config:              config,
	}
}

// Start запускает воркер в горутине
func (w *SubscriptionWorker) Start(ctx context.Context) {
	go w.run(ctx)
}

// run основной цикл воркера
func (w *SubscriptionWorker) run(ctx context.Context) {
	if w.config.IsDevMode() {
		log.Printf("🔄 Starting subscription worker in DEV mode (check every %s, renew every %s)",
			w.config.WorkerCheckInterval, w.config.SubscriptionInterval)
	} else {
		log.Printf("🔄 Starting subscription worker in PRODUCTION mode (check every %s, renew every %s)",
			w.config.WorkerCheckInterval, w.config.SubscriptionInterval)
	}

	ticker := time.NewTicker(w.config.WorkerCheckInterval)
	defer ticker.Stop()

	// Первая проверка сразу при запуске
	w.processSubscriptions()

	for {
		select {
		case <-ctx.Done():
			log.Println("🛑 Subscription worker stopped")
			return
		case <-ticker.C:
			w.processSubscriptions()
		}
	}
}

// processSubscriptions обрабатывает подписки, которые нужно продлить и повторные попытки
func (w *SubscriptionWorker) processSubscriptions() {
	// Обрабатываем обычные продления
	w.processRenewals()

	// Обрабатываем повторные попытки
	w.processRetries()
}

// processRenewals обрабатывает подписки для продления
func (w *SubscriptionWorker) processRenewals() {
	if w.config.IsDevMode() {
		log.Println("⏰ [DEV] Checking for subscriptions due for renewal...")
	} else {
		log.Println("⏰ [PROD] Checking for subscriptions due for renewal...")
	}

	subscriptions, err := w.subscriptionService.GetSubscriptionsDueForRenewal()
	if err != nil {
		log.Printf("❌ Error getting subscriptions for renewal: %v", err)
		return
	}

	if len(subscriptions) == 0 {
		if w.config.IsDevMode() {
			log.Println("✅ [DEV] No subscriptions due for renewal")
		} else {
			log.Println("✅ [PROD] No subscriptions due for renewal")
		}
	} else {
		log.Printf("🔄 Found %d subscription(s) due for renewal", len(subscriptions))

		for _, sub := range subscriptions {
			if err := w.subscriptionService.ProcessRecurringPayment(sub); err != nil {
				log.Printf("❌ Failed to process recurring payment for user %d: %v", sub.UserID, err)
			} else {
				if w.config.IsDevMode() {
					log.Printf("✅ [DEV] Processed recurring payment for user %d", sub.UserID)
				} else {
					log.Printf("✅ [PROD] Processed recurring payment for user %d", sub.UserID)
				}
			}
		}
	}
}

// processRetries обрабатывает повторные попытки оплаты
func (w *SubscriptionWorker) processRetries() {
	if w.config.IsDevMode() {
		log.Println("🔄 [DEV] Checking for subscriptions due for retry...")
	} else {
		log.Println("🔄 [PROD] Checking for subscriptions due for retry...")
	}

	subscriptions, err := w.subscriptionService.GetSubscriptionsDueForRetry()
	if err != nil {
		log.Printf("❌ Error getting subscriptions for retry: %v", err)
		return
	}

	if len(subscriptions) == 0 {
		if w.config.IsDevMode() {
			log.Println("✅ [DEV] No subscriptions due for retry")
		} else {
			log.Println("✅ [PROD] No subscriptions due for retry")
		}
	} else {
		log.Printf("🔄 Found %d subscription(s) due for retry", len(subscriptions))

		for _, sub := range subscriptions {
			log.Printf("🔄 Retrying payment for user %d (attempt %d)", sub.UserID, sub.FailedAttempts+1)
			if err := w.subscriptionService.ProcessRecurringPayment(sub); err != nil {
				log.Printf("❌ Failed to retry payment for user %d: %v", sub.UserID, err)
			} else {
				if w.config.IsDevMode() {
					log.Printf("✅ [DEV] Successfully retried payment for user %d", sub.UserID)
				} else {
					log.Printf("✅ [PROD] Successfully retried payment for user %d", sub.UserID)
				}
			}
		}
	}
}
