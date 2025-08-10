package domain

import "time"

// Subscription представляет подписку пользователя
type Subscription struct {
	ID                int64      `json:"id"`
	UserID            int64      `json:"user_id"`
	SubscriptionID    int        `json:"subscription_id"`
	Tariff            string     `json:"tariff"`
	Status            string     `json:"status"`
	Amount            float64    `json:"amount"`
	NextPayment       time.Time  `json:"next_payment"`
	LastPayment       time.Time  `json:"last_payment"`
	CreatedAt         time.Time  `json:"created_at"`
	CancelledAt       *time.Time `json:"cancelled_at,omitempty"`
	Active            bool       `json:"active"`
	YKCustomerID      *string    `json:"yk_customer_id"`
	YKPaymentMethodID *string    `json:"yk_payment_method_id"`
	YKLastPaymentID   *string    `json:"yk_last_payment_id"`
}

// SubscriptionStatus представляет статусы подписки
type SubscriptionStatus string

const (
	SubscriptionStatusPending   SubscriptionStatus = "pending"
	SubscriptionStatusActive    SubscriptionStatus = "active"
	SubscriptionStatusCancelled SubscriptionStatus = "cancelled"
	SubscriptionStatusExpired   SubscriptionStatus = "expired"
)

// Tariff представляет тариф подписки
type Tariff struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Price       float64  `json:"price"`
	Period      string   `json:"period"` // "month", "year"
	Description string   `json:"description"`
	Features    []string `json:"features"`
}

// SubscriptionRepository интерфейс для работы с подписками
type SubscriptionRepository interface {
	Create(subscription *Subscription) error
	GetByUserID(userID int64) (*Subscription, error)
	GetBySubscriptionID(subscriptionID int) (*Subscription, error)
	Update(subscription *Subscription) error
	UpdateStatus(userID int64, status SubscriptionStatus) error
	UpdateNextPayment(userID int64, nextPayment time.Time) error
	Cancel(userID int64) error
	GetActiveSubscriptions() ([]*Subscription, error)
	UpdateYooKassaBindings(userID int64, customerID, paymentMethodID, lastPaymentID string) error
}

// SubscriptionService интерфейс для бизнес-логики подписок
type SubscriptionService interface {
	CreateSubscription(userID int64, tariff string, amount float64) (*Subscription, error)
	GetUserSubscription(userID int64) (*Subscription, error)
	CancelSubscription(userID int64) error
	ProcessPayment(userID int64, amount float64) error
	IsUserSubscribed(userID int64) (bool, error)
	GetUserTariff(userID int64) (string, error)
}
