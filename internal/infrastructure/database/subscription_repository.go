package database

import (
	"ai_tg_writer/internal/domain"
	"database/sql"
	"time"
)

type SubscriptionRepository struct {
	db *DB
}

func NewSubscriptionRepository(db *DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

func (r *SubscriptionRepository) Create(subscription *domain.Subscription) error {
	query := `
		INSERT INTO subscriptions (user_id, subscription_id, tariff, status, amount, next_payment, last_payment, active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at`

	return r.db.QueryRow(
		query,
		subscription.UserID,
		subscription.SubscriptionID,
		subscription.Tariff,
		subscription.Status,
		subscription.Amount,
		subscription.NextPayment,
		subscription.LastPayment,
		subscription.Active,
	).Scan(&subscription.ID, &subscription.CreatedAt)
}

func (r *SubscriptionRepository) GetByUserID(userID int64) (*domain.Subscription, error) {
	query := `
		SELECT id, user_id, subscription_id, tariff, status, amount, next_payment, last_payment, created_at, cancelled_at, active
		FROM subscriptions
		WHERE user_id = $1 AND active = true
		ORDER BY created_at DESC
		LIMIT 1`

	subscription := &domain.Subscription{}
	err := r.db.QueryRow(query, userID).Scan(
		&subscription.ID,
		&subscription.UserID,
		&subscription.SubscriptionID,
		&subscription.Tariff,
		&subscription.Status,
		&subscription.Amount,
		&subscription.NextPayment,
		&subscription.LastPayment,
		&subscription.CreatedAt,
		&subscription.CancelledAt,
		&subscription.Active,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	return subscription, err
}

func (r *SubscriptionRepository) GetAnyByUserID(userID int64) (*domain.Subscription, error) {
	query := `
		SELECT id, user_id, subscription_id, tariff, status, amount, next_payment, last_payment, created_at, cancelled_at, active,
		       yk_customer_id, yk_payment_method_id, yk_last_payment_id
		FROM subscriptions
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 1`

	subscription := &domain.Subscription{}
	err := r.db.QueryRow(query, userID).Scan(
		&subscription.ID,
		&subscription.UserID,
		&subscription.SubscriptionID,
		&subscription.Tariff,
		&subscription.Status,
		&subscription.Amount,
		&subscription.NextPayment,
		&subscription.LastPayment,
		&subscription.CreatedAt,
		&subscription.CancelledAt,
		&subscription.Active,
		&subscription.YKCustomerID,
		&subscription.YKPaymentMethodID,
		&subscription.YKLastPaymentID,
	)

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, nil
		}
		return nil, err
	}

	return subscription, nil
}

func (r *SubscriptionRepository) GetSubscriptionsDueForRenewal() ([]*domain.Subscription, error) {
	query := `
		SELECT id, user_id, subscription_id, tariff, status, amount, next_payment, last_payment, created_at, cancelled_at, active,
		       yk_customer_id, yk_payment_method_id, yk_last_payment_id
		FROM subscriptions
		WHERE active = true 
		  AND status = 'active'
		  AND next_payment <= NOW()
		  AND yk_customer_id IS NOT NULL 
		  AND yk_payment_method_id IS NOT NULL`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subscriptions []*domain.Subscription
	for rows.Next() {
		subscription := &domain.Subscription{}
		err := rows.Scan(
			&subscription.ID,
			&subscription.UserID,
			&subscription.SubscriptionID,
			&subscription.Tariff,
			&subscription.Status,
			&subscription.Amount,
			&subscription.NextPayment,
			&subscription.LastPayment,
			&subscription.CreatedAt,
			&subscription.CancelledAt,
			&subscription.Active,
			&subscription.YKCustomerID,
			&subscription.YKPaymentMethodID,
			&subscription.YKLastPaymentID,
		)
		if err != nil {
			return nil, err
		}
		subscriptions = append(subscriptions, subscription)
	}

	return subscriptions, nil
}

// GetDB возвращает подключение к базе данных (временный метод для прямых запросов)
func (r *SubscriptionRepository) GetDB() *DB {
	return r.db
}

func (r *SubscriptionRepository) GetBySubscriptionID(subscriptionID int) (*domain.Subscription, error) {
	query := `
		SELECT id, user_id, subscription_id, tariff, status, amount, next_payment, last_payment, created_at, cancelled_at, active
		FROM subscriptions
		WHERE subscription_id = $1`

	subscription := &domain.Subscription{}
	err := r.db.QueryRow(query, subscriptionID).Scan(
		&subscription.ID,
		&subscription.UserID,
		&subscription.SubscriptionID,
		&subscription.Tariff,
		&subscription.Status,
		&subscription.Amount,
		&subscription.NextPayment,
		&subscription.LastPayment,
		&subscription.CreatedAt,
		&subscription.CancelledAt,
		&subscription.Active,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	return subscription, err
}

func (r *SubscriptionRepository) Update(subscription *domain.Subscription) error {
	query := `
		UPDATE subscriptions
		SET tariff = $1, status = $2, amount = $3, next_payment = $4, last_payment = $5, active = $6, cancelled_at = $7
		WHERE id = $8`

	_, err := r.db.Exec(
		query,
		subscription.Tariff,
		subscription.Status,
		subscription.Amount,
		subscription.NextPayment,
		subscription.LastPayment,
		subscription.Active,
		subscription.CancelledAt,
		subscription.ID,
	)
	return err
}

func (r *SubscriptionRepository) UpdateStatus(userID int64, status domain.SubscriptionStatus) error {
	query := `UPDATE subscriptions SET status = $1 WHERE user_id = $2 AND active = true`
	_, err := r.db.Exec(query, string(status), userID)
	return err
}

func (r *SubscriptionRepository) UpdateNextPayment(userID int64, nextPayment time.Time) error {
	query := `UPDATE subscriptions SET next_payment = $1 WHERE user_id = $2 AND active = true`
	_, err := r.db.Exec(query, nextPayment, userID)
	return err
}

func (r *SubscriptionRepository) Cancel(userID int64) error {
	now := time.Now()
	query := `UPDATE subscriptions SET active = false, cancelled_at = $1 WHERE user_id = $2 AND active = true`
	_, err := r.db.Exec(query, now, userID)
	return err
}

func (r *SubscriptionRepository) GetActiveSubscriptions() ([]*domain.Subscription, error) {
	query := `
		SELECT id, user_id, subscription_id, tariff, status, amount, next_payment, last_payment, created_at, cancelled_at, active
		FROM subscriptions
		WHERE active = true AND status = 'active'`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subscriptions []*domain.Subscription
	for rows.Next() {
		subscription := &domain.Subscription{}
		err := rows.Scan(
			&subscription.ID,
			&subscription.UserID,
			&subscription.SubscriptionID,
			&subscription.Tariff,
			&subscription.Status,
			&subscription.Amount,
			&subscription.NextPayment,
			&subscription.LastPayment,
			&subscription.CreatedAt,
			&subscription.CancelledAt,
			&subscription.Active,
		)
		if err != nil {
			return nil, err
		}
		subscriptions = append(subscriptions, subscription)
	}

	return subscriptions, nil
}

func (r *SubscriptionRepository) UpdateYooKassaBindings(userID int64, customerID, paymentMethodID, lastPaymentID string) error {
	// Обновляем самую новую подписку пользователя (включая pending)
	_, err := r.db.Exec(`
		UPDATE subscriptions
		SET yk_customer_id = $1, yk_payment_method_id = $2, yk_last_payment_id = $3
		WHERE id = (
			SELECT id FROM subscriptions 
			WHERE user_id = $4 
			ORDER BY created_at DESC 
			LIMIT 1
		)`,
		customerID, paymentMethodID, lastPaymentID, userID,
	)
	return err
}
