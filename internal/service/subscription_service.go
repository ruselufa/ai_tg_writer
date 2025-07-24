package service

import (
	"ai_tg_writer/internal/domain"
	"ai_tg_writer/internal/infrastructure/prodamus_payments"
	"fmt"
	"time"
)

type SubscriptionService struct {
	repo            domain.SubscriptionRepository
	prodamusHandler *prodamus_payments.ProdamusHandler
}

func NewSubscriptionService(repo domain.SubscriptionRepository, prodamusHandler *prodamus_payments.ProdamusHandler) *SubscriptionService {
	return &SubscriptionService{
		repo:            repo,
		prodamusHandler: prodamusHandler,
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

	// Отменяем подписку в Prodamus
	if err := s.prodamusHandler.SetSubscriptionActivity(subscription.SubscriptionID, userID, false); err != nil {
		return fmt.Errorf("error cancelling subscription in Prodamus: %w", err)
	}

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
	subscriptionID := 1 // ID подписки в Prodamus

	return s.prodamusHandler.CreateSubscriptionLink(userID, tariff, amount, subscriptionID)
}

// GetAvailableTariffs возвращает доступные тарифы
func (s *SubscriptionService) GetAvailableTariffs() []domain.Tariff {
	return []domain.Tariff{
		{
			ID:          "premium",
			Name:        "Premium",
			Price:       299.0,
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
