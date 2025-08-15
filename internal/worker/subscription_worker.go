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

	// Обрабатываем истекшие отмененные подписки
	w.processExpiredCancelledSubscriptions()
}

// processRenewals обрабатывает подписки для продления
func (w *SubscriptionWorker) processRenewals() {
	now := time.Now()
	utcNow := time.Now().UTC()
	if w.config.IsDevMode() {
		log.Printf("⏰ [DEV] Checking for subscriptions due for renewal... [NOW: %s, UTC: %s]",
			now.Format("2006-01-02 15:04:05"), utcNow.Format("2006-01-02 15:04:05"))
	} else {
		log.Printf("⏰ [PROD] Checking for subscriptions due for renewal... [NOW: %s, UTC: %s]",
			now.Format("2006-01-02 15:04:05"), utcNow.Format("2006-01-02 15:04:05"))
	}

	// Диагностика: показываем все активные подписки
	allActive, err := w.subscriptionService.GetAllActiveSubscriptions()
	if err != nil {
		log.Printf("⚠️ [DEBUG] Error getting all active subscriptions: %v", err)
	} else {
		log.Printf("🔍 [DEBUG] All active subscriptions (%d):", len(allActive))
		for _, sub := range allActive {
			nextPaymentStr := "NULL"
			if sub.NextPayment != (time.Time{}) {
				nextPaymentStr = sub.NextPayment.Format("2006-01-02 15:04:05")
				// Проверяем, прошло ли время next_payment
				isPast := sub.NextPayment.Before(now)
				isPastUTC := sub.NextPayment.Before(utcNow)
				timeDiff := now.Sub(sub.NextPayment)
				timeDiffUTC := utcNow.Sub(sub.NextPayment)

				nextRetryStr := "NULL"
				if sub.NextRetry != nil {
					nextRetryStr = sub.NextRetry.Format("2006-01-02 15:04:05")
				}

				log.Printf("   ID=%d, UserID=%d, Status=%s, NextPayment=%s, IsPast(local)=%v, IsPast(UTC)=%v, TimeDiff(local)=%v, TimeDiff(UTC)=%v, Active=%v, YKCustomerID=%v, YKPaymentMethodID=%v, FailedAttempts=%d, NextRetry=%s",
					sub.ID, sub.UserID, sub.Status, nextPaymentStr, isPast, isPastUTC, timeDiff, timeDiffUTC, sub.Active,
					sub.YKCustomerID != nil, sub.YKPaymentMethodID != nil, sub.FailedAttempts, nextRetryStr)
			} else {
				nextRetryStr := "NULL"
				if sub.NextRetry != nil {
					nextRetryStr = sub.NextRetry.Format("2006-01-02 15:04:05")
				}

				log.Printf("   ID=%d, UserID=%d, Status=%s, NextPayment=%s, Active=%v, YKCustomerID=%v, YKPaymentMethodID=%v, FailedAttempts=%d, NextRetry=%s",
					sub.ID, sub.UserID, sub.Status, nextPaymentStr, sub.Active,
					sub.YKCustomerID != nil, sub.YKPaymentMethodID != nil, sub.FailedAttempts, nextRetryStr)
			}
		}
	}

	subscriptions, err := w.subscriptionService.GetSubscriptionsDueForRenewal()
	if err != nil {
		log.Printf("❌ Error getting subscriptions for renewal: %v", err)
		return
	}

	if len(subscriptions) == 0 {
		if w.config.IsDevMode() {
			log.Printf("✅ [DEV] No subscriptions due for renewal [NOW: %s]", now.Format("2006-01-02 15:04:05"))
		} else {
			log.Printf("✅ [PROD] No subscriptions due for renewal [NOW: %s]", now.Format("2006-01-02 15:04:05"))
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
	now := time.Now()
	if w.config.IsDevMode() {
		log.Printf("🔄 [DEV] Checking for subscriptions due for retry... [NOW: %s]", now.Format("2006-01-02 15:04:05"))
	} else {
		log.Printf("🔄 [PROD] Checking for subscriptions due for retry... [NOW: %s]", now.Format("2006-01-02 15:04:05"))
	}

	subscriptions, err := w.subscriptionService.GetSubscriptionsDueForRetry()
	if err != nil {
		log.Printf("❌ Error getting subscriptions for retry: %v", err)
		return
	}

	if len(subscriptions) == 0 {
		if w.config.IsDevMode() {
			log.Printf("✅ [DEV] No subscriptions due for retry [NOW: %s]", now.Format("2006-01-02 15:04:05"))
		} else {
			log.Printf("✅ [PROD] No subscriptions due for retry [NOW: %s]", now.Format("2006-01-02 15:04:05"))
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

// processExpiredCancelledSubscriptions обрабатывает истекшие отмененные подписки
func (w *SubscriptionWorker) processExpiredCancelledSubscriptions() {
	now := time.Now()
	if w.config.IsDevMode() {
		log.Printf("⏰ [DEV] Checking for expired cancelled subscriptions... [NOW: %s]", now.Format("2006-01-02 15:04:05"))
	} else {
		log.Printf("⏰ [PROD] Checking for expired cancelled subscriptions... [NOW: %s]", now.Format("2006-01-02 15:04:05"))
	}

	// Получаем все активные подписки со статусом 'cancelled'
	allActive, err := w.subscriptionService.GetAllActiveSubscriptions()
	if err != nil {
		log.Printf("❌ Error getting all active subscriptions: %v", err)
		return
	}

	expiredCount := 0
	for _, sub := range allActive {
		// Проверяем только отмененные подписки
		if sub.Status == "cancelled" && sub.NextPayment.Before(now) {
			log.Printf("🔄 Found expired cancelled subscription for user %d (expired at %s)",
				sub.UserID, sub.NextPayment.Format("2006-01-02 15:04:05"))

			// Полностью отменяем подписку
			if err := w.subscriptionService.CancelExpiredSubscription(sub.UserID); err != nil {
				log.Printf("❌ Failed to cancel expired subscription for user %d: %v", sub.UserID, err)
			} else {
				log.Printf("✅ Successfully cancelled expired subscription for user %d", sub.UserID)
				expiredCount++
			}
		}
	}

	if expiredCount > 0 {
		log.Printf("✅ Processed %d expired cancelled subscription(s)", expiredCount)
	} else {
		if w.config.IsDevMode() {
			log.Printf("✅ [DEV] No expired cancelled subscriptions found")
		} else {
			log.Printf("✅ [PROD] No expired cancelled subscriptions found")
		}
	}
}
