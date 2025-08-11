package service

import (
	"ai_tg_writer/internal/domain"
	"ai_tg_writer/internal/infrastructure/yookassa"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

type SubscriptionService struct {
	repo domain.SubscriptionRepository
	yk   *yookassa.Client
}

func NewSubscriptionService(repo domain.SubscriptionRepository, ykClient *yookassa.Client) *SubscriptionService {
	return &SubscriptionService{
		repo: repo,
		yk:   ykClient,
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
		SubscriptionID: 1, // ID подписки в Prodamus
		Tariff:         tariff,
		Status:         string(domain.SubscriptionStatusPending),
		Amount:         amount,
		NextPayment:    time.Now().AddDate(0, 1, 0), // +1 месяц
		LastPayment:    time.Now(),
		Active:         true,
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
	subscription, err := s.repo.GetByUserID(userID)
	if err != nil {
		return fmt.Errorf("error getting subscription: %w", err)
	}

	if subscription == nil {
		return fmt.Errorf("subscription not found")
	}

	// Временно отключаем интеграцию с Prodamus
	// TODO: Добавить интеграцию с новым платежным модулем
	// if err := s.prodamusHandler.SetSubscriptionActivity(subscription.SubscriptionID, userID, false); err != nil {
	// 	return fmt.Errorf("error cancelling subscription in Prodamus: %w", err)
	// }

	// Отменяем подписку в базе данных
	if err := s.repo.Cancel(userID); err != nil {
		return fmt.Errorf("error cancelling subscription in database: %w", err)
	}

	return nil
}

// ProcessPayment обрабатывает успешный платеж
func (s *SubscriptionService) ProcessPayment(userID int64, amount float64) error {
	subscription, err := s.repo.GetByUserID(userID)
	if err != nil {
		return fmt.Errorf("error getting subscription: %w", err)
	}

	if subscription == nil {
		return fmt.Errorf("subscription not found")
	}

	// Обновляем статус и даты
	subscription.Status = string(domain.SubscriptionStatusActive)
	subscription.LastPayment = time.Now()
	subscription.NextPayment = time.Now().AddDate(0, 1, 0) // +1 месяц

	if err := s.repo.Update(subscription); err != nil {
		return fmt.Errorf("error updating subscription: %w", err)
	}

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
	// Убедимся, что есть запись подписки в БД (pending)
	sub, err := s.repo.GetByUserID(userID)
	if err != nil {
		return "", fmt.Errorf("get subscription: %w", err)
	}
	if sub == nil {
		if _, err := s.CreateSubscription(userID, tariff, amount); err != nil {
			return "", err
		}
	}

	if s.yk == nil {
		return "", fmt.Errorf("yookassa client is not configured")
	}

	// Формируем платеж с сохранением метода
	value := fmt.Sprintf("%.2f", amount)
	idem := fmt.Sprintf("%d-%d", userID, time.Now().UnixNano())
	returnURL := getenv("YK_RETURN_URL_ADDRESS", "")

	payment, err := s.yk.CreateInitialPayment(
		idem,
		yookassa.Amount{Value: value, Currency: "RUB"},
		"Подписка AI TG Writer",
		strconv.FormatInt(userID, 10),
		returnURL,
		map[string]string{"tg_user_id": strconv.FormatInt(userID, 10)},
	)
	if err != nil {
		return "", fmt.Errorf("create initial payment: %w", err)
	}

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
