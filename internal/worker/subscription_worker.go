package worker

import (
	"ai_tg_writer/internal/config"
	"ai_tg_writer/internal/service"
	"context"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"ai_tg_writer/internal/domain"
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
	// Запускаем мониторинг производительности YooKassa
	go w.monitorYooKassaPerformance()

	go w.run(ctx)
}

// monitorYooKassaPerformance мониторит производительность YooKassa
func (w *SubscriptionWorker) monitorYooKassaPerformance() {
	ticker := time.NewTicker(5 * time.Minute) // Каждые 5 минут
	defer ticker.Stop()

	for range ticker.C {
		log.Printf("💳 [YooKassa Monitor] Воркер подписок работает, следующий запуск через 5 минут")
	}
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

		// Создаем семафор для ограничения одновременных запросов к YooKassa
		// Ограничиваем до 3, чтобы не перегружать YooKassa API
		const maxConcurrentPayments = 3
		semaphore := make(chan struct{}, maxConcurrentPayments)
		var wg sync.WaitGroup

		// Счетчик успешных и неуспешных платежей
		var successCount, errorCount int32

		log.Printf("💳 [YooKassa] Начинаем параллельную обработку %d подписок (лимит: %d)",
			len(subscriptions), maxConcurrentPayments)

		for _, sub := range subscriptions {
			wg.Add(1)

			go func(subscription *domain.Subscription) {
				defer wg.Done()

				// Получаем слот для обработки платежа
				semaphore <- struct{}{}
				defer func() { <-semaphore }()

				startTime := time.Now()
				log.Printf("💳 [YooKassa] Начинаем обработку платежа для пользователя %d (ID: %d)",
					subscription.UserID, subscription.ID)

				// Обрабатываем платеж
				if err := w.subscriptionService.ProcessRecurringPayment(subscription); err != nil {
					atomic.AddInt32(&errorCount, 1)
					log.Printf("❌ [YooKassa] Failed to process recurring payment for user %d: %v",
						subscription.UserID, err)
				} else {
					atomic.AddInt32(&successCount, 1)
					duration := time.Since(startTime)
					log.Printf("✅ [YooKassa] Successfully processed recurring payment for user %d in %v",
						subscription.UserID, duration)
				}
			}(sub)
		}

		// Ждем завершения всех платежей
		wg.Wait()

		// Логируем итоговую статистику
		finalSuccess := atomic.LoadInt32(&successCount)
		finalError := atomic.LoadInt32(&errorCount)
		log.Printf("🎉 [YooKassa] Все платежи обработаны. Успешно: %d, Ошибок: %d",
			finalSuccess, finalError)
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

		// Создаем семафор для ограничения одновременных повторных попыток
		const maxConcurrentRetries = 2 // Меньше для повторных попыток
		semaphore := make(chan struct{}, maxConcurrentRetries)
		var wg sync.WaitGroup

		// Счетчик успешных и неуспешных повторных попыток
		var successCount, errorCount int32

		log.Printf("🔄 [Retry] Начинаем параллельную обработку %d повторных попыток (лимит: %d)",
			len(subscriptions), maxConcurrentRetries)

		for _, sub := range subscriptions {
			wg.Add(1)

			go func(subscription *domain.Subscription) {
				defer wg.Done()

				// Получаем слот для обработки
				semaphore <- struct{}{}
				defer func() { <-semaphore }()

				startTime := time.Now()
				log.Printf("🔄 [Retry] Начинаем повторную попытку для пользователя %d (ID: %d, Попытка: %d)",
					subscription.UserID, subscription.ID, subscription.FailedAttempts+1)

				// Обрабатываем повторную попытку
				if err := w.subscriptionService.ProcessRecurringPayment(subscription); err != nil {
					atomic.AddInt32(&errorCount, 1)
					log.Printf("❌ [Retry] Failed to retry payment for user %d: %v", subscription.UserID, err)
				} else {
					atomic.AddInt32(&successCount, 1)
					duration := time.Since(startTime)
					log.Printf("✅ [Retry] Successfully retried payment for user %d in %v",
						subscription.UserID, duration)
				}
			}(sub)
		}

		// Ждем завершения всех повторных попыток
		wg.Wait()

		// Логируем итоговую статистику
		finalSuccess := atomic.LoadInt32(&successCount)
		finalError := atomic.LoadInt32(&errorCount)
		log.Printf("🎉 [Retry] Все повторные попытки завершены. Успешно: %d, Ошибок: %d",
			finalSuccess, finalError)
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
