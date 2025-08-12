package service

import (
	"ai_tg_writer/internal/config"
	"ai_tg_writer/internal/domain"
	"ai_tg_writer/internal/infrastructure/yookassa"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

type SubscriptionService struct {
	repo   domain.SubscriptionRepository
	yk     *yookassa.Client
	config *config.Config
}

func NewSubscriptionService(repo domain.SubscriptionRepository, ykClient *yookassa.Client, cfg *config.Config) *SubscriptionService {
	return &SubscriptionService{
		repo:   repo,
		yk:     ykClient,
		config: cfg,
	}
}

// CreateSubscription создает новую подписку
func (s *SubscriptionService) CreateSubscription(userID int64, tariff string, amount float64) (*domain.Subscription, error) {
	// Проверяем, нет ли уже активной подписки
	existing, err := s.repo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("error checking existing subscription: %w", err)
	}

	if existing != nil && existing.Active {
		return nil, fmt.Errorf("user already has active subscription")
	}

	// Создаем подписку в базе данных
	subscription := &domain.Subscription{
		UserID:         userID,
		SubscriptionID: nil, // Не используется для YooKassa (только для Prodamus)
		Tariff:         tariff,
		Status:         string(domain.SubscriptionStatusPending),
		Amount:         amount,
		NextPayment:    time.Now().Add(s.config.SubscriptionInterval), // Используем конфигурацию
		LastPayment:    time.Now(),
		Active:         false, // Станет true после успешной оплаты
	}

	if err := s.repo.Create(subscription); err != nil {
		return nil, fmt.Errorf("error creating subscription: %w", err)
	}

	return subscription, nil
}

// GetUserSubscription получает подписку пользователя
func (s *SubscriptionService) GetUserSubscription(userID int64) (*domain.Subscription, error) {
	return s.repo.GetByUserID(userID)
}

// CancelSubscription отменяет подписку
func (s *SubscriptionService) CancelSubscription(userID int64) error {
	log.Printf("🔄 Starting subscription cancellation for user %d", userID)

	subscription, err := s.repo.GetByUserID(userID)
	if err != nil {
		log.Printf("❌ Error getting subscription for user %d: %v", userID, err)
		return fmt.Errorf("error getting subscription: %w", err)
	}

	if subscription == nil {
		log.Printf("❌ Subscription not found for user %d", userID)
		return fmt.Errorf("subscription not found")
	}

	log.Printf("✅ Found subscription for user %d: ID=%d, Status=%s, Active=%v",
		userID, subscription.ID, subscription.Status, subscription.Active)

	// Временно отключаем интеграцию с Prodamus
	// TODO: Добавить интеграцию с новым платежным модулем
	// if err := s.prodamusHandler.SetSubscriptionActivity(subscription.SubscriptionID, userID, false); err != nil {
	// 	return fmt.Errorf("error cancelling subscription in Prodamus: %w", err)
	// }

	// Отменяем подписку в базе данных
	log.Printf("🔄 Cancelling subscription in database for user %d", userID)
	if err := s.repo.Cancel(userID); err != nil {
		log.Printf("❌ Error cancelling subscription in database for user %d: %v", userID, err)
		return fmt.Errorf("error cancelling subscription in database: %w", err)
	}

	log.Printf("✅ Subscription cancelled successfully for user %d", userID)
	return nil
}

// ProcessPayment обрабатывает успешный платеж
func (s *SubscriptionService) ProcessPayment(userID int64, amount float64) error {
	// Ищем любую подписку пользователя (включая неактивную pending)
	subscription, err := s.repo.GetAnyByUserID(userID)
	if err != nil {
		return fmt.Errorf("error getting subscription: %w", err)
	}

	if subscription == nil {
		return fmt.Errorf("subscription not found")
	}

	log.Printf("📝 Activating subscription for user %d: ID=%d, Status=%s, Active=%v",
		userID, subscription.ID, subscription.Status, subscription.Active)

	// Обновляем статус и даты
	subscription.Status = string(domain.SubscriptionStatusActive)
	subscription.LastPayment = time.Now()
	subscription.NextPayment = time.Now().Add(s.config.SubscriptionInterval) // Используем конфигурацию
	subscription.Active = true                                               // Активируем подписку

	if err := s.repo.Update(subscription); err != nil {
		return fmt.Errorf("error updating subscription: %w", err)
	}

	log.Printf("✅ Subscription activated successfully for user %d", userID)
	return nil
}

// IsUserSubscribed проверяет, есть ли у пользователя активная подписка
func (s *SubscriptionService) IsUserSubscribed(userID int64) (bool, error) {
	subscription, err := s.repo.GetByUserID(userID)
	if err != nil {
		return false, fmt.Errorf("error checking subscription: %w", err)
	}

	return subscription != nil && subscription.Active && subscription.Status == string(domain.SubscriptionStatusActive), nil
}

// GetUserTariff получает тариф пользователя
func (s *SubscriptionService) GetUserTariff(userID int64) (string, error) {
	subscription, err := s.repo.GetByUserID(userID)
	if err != nil {
		return "", fmt.Errorf("error getting subscription: %w", err)
	}

	if subscription == nil || !subscription.Active {
		return "free", nil
	}

	return subscription.Tariff, nil
}

// CreateSubscriptionLink создает ссылку для оплаты подписки
func (s *SubscriptionService) CreateSubscriptionLink(userID int64, tariff string, amount float64) (string, error) {
	log.Printf("=== CreateSubscriptionLink START ===")
	log.Printf("UserID: %d, Tariff: %s, Amount: %.2f", userID, tariff, amount)

	// Убедимся, что есть запись подписки в БД (pending)
	sub, err := s.repo.GetByUserID(userID)
	if err != nil {
		log.Printf("❌ Error getting subscription: %v", err)
		return "", fmt.Errorf("get subscription: %w", err)
	}
	log.Printf("✅ Subscription check passed, sub=%v", sub != nil)

	if sub == nil {
		log.Printf("📝 Creating new subscription...")
		if _, err := s.CreateSubscription(userID, tariff, amount); err != nil {
			log.Printf("❌ Error creating subscription: %v", err)
			return "", err
		}
		log.Printf("✅ Subscription created successfully")
	}

	if s.yk == nil {
		log.Printf("❌ YooKassa client is nil")
		return "", fmt.Errorf("yookassa client is not configured")
	}
	log.Printf("✅ YooKassa client is configured")

	// Формируем платеж с сохранением метода
	value := fmt.Sprintf("%.2f", amount)
	idem := fmt.Sprintf("%d-%d", userID, time.Now().UnixNano())
	returnURL := getenv("YK_RETURN_URL_ADDRESS", "")

	log.Printf("💳 Calling YooKassa CreateInitialPayment...")
	log.Printf("   Value: %s, IdempotenceKey: %s", value, idem)
	log.Printf("   ReturnURL: %s", returnURL)
	log.Printf("   CustomerID: %s", strconv.FormatInt(userID, 10))

	payment, err := s.yk.CreateInitialPayment(
		idem,
		yookassa.Amount{Value: value, Currency: "RUB"},
		"Подписка AI TG Writer",
		strconv.FormatInt(userID, 10),
		returnURL,
		map[string]string{"tg_user_id": strconv.FormatInt(userID, 10)},
	)
	if err != nil {
		log.Printf("❌ YooKassa CreateInitialPayment error: %v", err)
		return "", fmt.Errorf("create initial payment: %w", err)
	}
	log.Printf("✅ YooKassa CreateInitialPayment success")

	// Логируем весь ответ от YooKassa для отладки
	log.Printf("=== YooKassa CreateInitialPayment Response ===")
	log.Printf("UserID: %d, Amount: %s, IdempotenceKey: %s", userID, value, idem)
	log.Printf("Full response: %+v", payment)

	// Выводим ключевые поля
	if id, ok := payment["id"].(string); ok {
		log.Printf("Payment ID: %s", id)
	}
	if status, ok := payment["status"].(string); ok {
		log.Printf("Payment Status: %s", status)
	}
	if conf, ok := payment["confirmation"].(map[string]any); ok {
		log.Printf("Confirmation: %+v", conf)
		if confURL, ok := conf["confirmation_url"].(string); ok {
			log.Printf("Confirmation URL: %s", confURL)
		}
	}
	log.Printf("=== End YooKassa Response ===")

	// Достаем confirmation_url из ответа
	conf, ok := payment["confirmation"].(map[string]any)
	if !ok {
		return "", fmt.Errorf("confirmation not found in response")
	}
	url, _ := conf["confirmation_url"].(string)
	if url == "" {
		return "", fmt.Errorf("confirmation_url not found")
	}
	return url, nil
}

// SaveYooKassaBindingAndActivate сохраняет customer/payment_method и активирует подписку
func (s *SubscriptionService) SaveYooKassaBindingAndActivate(userID int64, customerID, paymentMethodID, paymentID string, amount float64) error {
	if err := s.repo.UpdateYooKassaBindings(userID, customerID, paymentMethodID, paymentID); err != nil {
		return fmt.Errorf("update bindings: %w", err)
	}
	return s.ProcessPayment(userID, amount)
}

// small helper for env with default
func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// GetSubscriptionsDueForRenewal возвращает подписки, которые нужно продлить
func (s *SubscriptionService) GetSubscriptionsDueForRenewal() ([]*domain.Subscription, error) {
	return s.repo.GetSubscriptionsDueForRenewal()
}

// ProcessRecurringPayment обрабатывает рекуррентный платеж для подписки
func (s *SubscriptionService) ProcessRecurringPayment(subscription *domain.Subscription) error {
	if subscription.YKCustomerID == nil || subscription.YKPaymentMethodID == nil {
		return fmt.Errorf("missing YooKassa binding data")
	}

	log.Printf("🔄 Processing recurring payment for user %d, subscription ID %d",
		subscription.UserID, subscription.ID)

	// Создаем идемпотентный ключ
	idempotenceKey := fmt.Sprintf("%d-recurring-%d", subscription.UserID, time.Now().Unix())

	// Создаем рекуррентный платеж
	payment, err := s.yk.CreateRecurringPayment(
		idempotenceKey,
		yookassa.Amount{
			Value:    fmt.Sprintf("%.2f", subscription.Amount),
			Currency: "RUB",
		},
		"Продление подписки AI TG Writer",
		*subscription.YKCustomerID,
		*subscription.YKPaymentMethodID,
		map[string]string{
			"tg_user_id":      fmt.Sprintf("%d", subscription.UserID),
			"subscription_id": fmt.Sprintf("%d", subscription.ID),
			"type":            "recurring",
		},
	)

	if err != nil {
		log.Printf("❌ Recurring payment failed for user %d: %v", subscription.UserID, err)
		return err
	}

	log.Printf("✅ Recurring payment created for user %d: %s", subscription.UserID, payment["id"])

	// Обновляем дату следующего платежа на точно такой же период
	subscription.NextPayment = time.Now().Add(s.config.SubscriptionInterval)
	subscription.LastPayment = time.Now()
	if err := s.repo.Update(subscription); err != nil {
		log.Printf("❌ Failed to update next payment date: %v", err)
		return err
	}

	if s.config.IsDevMode() {
		log.Printf("📅 [DEV] Next payment scheduled for user %d at %s",
			subscription.UserID, subscription.NextPayment.Format("15:04:05"))
	} else {
		log.Printf("📅 [PROD] Next payment scheduled for user %d at %s",
			subscription.UserID, subscription.NextPayment.Format("2006-01-02 15:04:05"))
	}

	return nil
}

// GetAvailableTariffs возвращает доступные тарифы
func (s *SubscriptionService) GetAvailableTariffs() []domain.Tariff {
	return []domain.Tariff{
		{
			ID:          "premium",
			Name:        "Premium",
			Price:       990.0,
			Period:      "month",
			Description: "Премиум подписка с неограниченными возможностями",
			Features: []string{
				"Неограниченное количество запросов",
				"Приоритетная поддержка",
				"Расширенные функции",
				"Доступ к новым возможностям",
			},
		},
	}
}
