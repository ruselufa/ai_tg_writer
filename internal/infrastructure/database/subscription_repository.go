package database

import (
	"ai_tg_writer/internal/domain"
	"database/sql"
	"log"
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
		SELECT id, user_id, subscription_id, tariff, status, amount, next_payment, last_payment, created_at, cancelled_at, active,
		       yk_customer_id, yk_payment_method_id, yk_last_payment_id, failed_attempts, next_retry, suspended_at
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
		&subscription.YKCustomerID,
		&subscription.YKPaymentMethodID,
		&subscription.YKLastPaymentID,
		&subscription.FailedAttempts,
		&subscription.NextRetry,
		&subscription.SuspendedAt,
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
	// Теперь все время хранится в UTC, поэтому используем просто NOW()
	query := `
		SELECT id, user_id, subscription_id, tariff, status, amount, next_payment, last_payment, created_at, cancelled_at, active,
		       yk_customer_id, yk_payment_method_id, yk_last_payment_id
		FROM subscriptions
		WHERE active = true 
		  AND status = 'active'
		  AND next_payment <= NOW()
		  AND yk_customer_id IS NOT NULL 
		  AND yk_payment_method_id IS NOT NULL`

	// Добавляем отладочную информацию
	log.Printf("🔍 [SQL DEBUG] GetSubscriptionsDueForRenewal query: %s", query)
	log.Printf("🔍 [SQL DEBUG] Current time (NOW()): %s", time.Now().Format("2006-01-02 15:04:05"))
	log.Printf("🔍 [SQL DEBUG] Current time (UTC): %s", time.Now().UTC().Format("2006-01-02 15:04:05"))

	// Проверяем время в базе данных
	var dbNow time.Time
	err := r.db.QueryRow("SELECT NOW()").Scan(&dbNow)
	if err != nil {
		log.Printf("⚠️ [SQL DEBUG] Error getting DB time: %v", err)
	} else {
		log.Printf("🔍 [SQL DEBUG] Database time (NOW()): %s", dbNow.Format("2006-01-02 15:04:05"))
		log.Printf("🔍 [SQL DEBUG] Database time (UTC): %s", dbNow.UTC().Format("2006-01-02 15:04:05"))
	}

	// Тестируем условие next_payment <= NOW() для каждой подписки
	testQuery := `
		SELECT id, user_id, next_payment, 
		       next_payment <= NOW() as is_due,
		       NOW() as current_db_time
		FROM subscriptions
		WHERE active = true AND status = 'active'`

	testRows, err := r.db.Query(testQuery)
	if err != nil {
		log.Printf("⚠️ [SQL DEBUG] Test query error: %v", err)
	} else {
		defer testRows.Close()
		log.Printf("🔍 [SQL DEBUG] Testing next_payment <= NOW() condition:")
		for testRows.Next() {
			var id int64
			var userID int64
			var nextPayment sql.NullTime
			var isDue bool
			var currentDBTime time.Time

			err := testRows.Scan(&id, &userID, &nextPayment, &isDue, &currentDBTime)
			if err != nil {
				log.Printf("⚠️ [SQL DEBUG] Test row scan error: %v", err)
				continue
			}

			nextPaymentStr := "NULL"
			if nextPayment.Valid {
				nextPaymentStr = nextPayment.Time.Format("2006-01-02 15:04:05")
			}

			log.Printf("   ID=%d, UserID=%d, NextPayment=%s, IsDue=%v, CurrentDBTime=%s",
				id, userID, nextPaymentStr, isDue, currentDBTime.Format("2006-01-02 15:04:05"))
		}
	}

	// Сначала проверим все активные подписки без фильтров
	debugQuery := `
		SELECT id, user_id, tariff, status, active, next_payment, yk_customer_id, yk_payment_method_id
		FROM subscriptions
		WHERE active = true`

	debugRows, err := r.db.Query(debugQuery)
	if err != nil {
		log.Printf("⚠️ [SQL DEBUG] Debug query error: %v", err)
	} else {
		defer debugRows.Close()
		log.Printf("🔍 [SQL DEBUG] All active subscriptions (before filters):")
		for debugRows.Next() {
			var id int64
			var userID int64
			var tariff, status string
			var active bool
			var nextPayment sql.NullTime
			var ykCustomerID, ykPaymentMethodID sql.NullString

			err := debugRows.Scan(&id, &userID, &tariff, &status, &active, &nextPayment, &ykCustomerID, &ykPaymentMethodID)
			if err != nil {
				log.Printf("⚠️ [SQL DEBUG] Debug row scan error: %v", err)
				continue
			}

			nextPaymentStr := "NULL"
			if nextPayment.Valid {
				nextPaymentStr = nextPayment.Time.Format("2006-01-02 15:04:05")
			}

			log.Printf("   ID=%d, UserID=%d, Status=%s, Active=%v, NextPayment=%s, YKCustomerID=%v, YKPaymentMethodID=%v",
				id, userID, status, active, nextPaymentStr,
				ykCustomerID.Valid, ykPaymentMethodID.Valid)
		}
	}

	rows, err := r.db.Query(query)
	if err != nil {
		log.Printf("❌ [SQL DEBUG] Query error: %v", err)
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
			log.Printf("❌ [SQL DEBUG] Row scan error: %v", err)
			return nil, err
		}
		subscriptions = append(subscriptions, subscription)
		log.Printf("🔍 [SQL DEBUG] Found subscription: ID=%d, UserID=%d, NextPayment=%s, Active=%v, Status=%s",
			subscription.ID, subscription.UserID, subscription.NextPayment.Format("2006-01-02 15:04:05"),
			subscription.Active, subscription.Status)
	}

	log.Printf("🔍 [SQL DEBUG] Total subscriptions found: %d", len(subscriptions))
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
		SET tariff = $1, status = $2, amount = $3, next_payment = $4, last_payment = $5, active = $6, cancelled_at = $7,
		    yk_customer_id = $8, yk_payment_method_id = $9, yk_last_payment_id = $10, 
		    failed_attempts = $11, next_retry = $12, suspended_at = $13
		WHERE id = $14`

	_, err := r.db.Exec(
		query,
		subscription.Tariff,
		subscription.Status,
		subscription.Amount,
		subscription.NextPayment,
		subscription.LastPayment,
		subscription.Active,
		subscription.CancelledAt,
		subscription.YKCustomerID,
		subscription.YKPaymentMethodID,
		subscription.YKLastPaymentID,
		subscription.FailedAttempts,
		subscription.NextRetry,
		subscription.SuspendedAt,
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
	now := time.Now().UTC() // Используем UTC время
	query := `UPDATE subscriptions SET 
		active = false, 
		cancelled_at = $1,
		next_payment = NULL,
		yk_payment_method_id = NULL,
		yk_last_payment_id = NULL,
		failed_attempts = 0,
		next_retry = NULL,
		suspended_at = NULL
		WHERE user_id = $2 AND active = true`
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

// GetSubscriptionsDueForRetry получает подписки для повторной попытки оплаты
func (r *SubscriptionRepository) GetSubscriptionsDueForRetry() ([]*domain.Subscription, error) {
	// Теперь все время хранится в UTC, поэтому используем просто NOW()
	query := `
		SELECT id, user_id, subscription_id, tariff, status, amount, next_payment, last_payment, created_at, cancelled_at, active,
		       yk_customer_id, yk_payment_method_id, yk_last_payment_id, failed_attempts, next_retry, suspended_at
		FROM subscriptions
		WHERE active = true 
		  AND status = 'active'
		  AND failed_attempts > 0
		  AND failed_attempts < 3
		  AND next_retry <= NOW()
		  AND yk_customer_id IS NOT NULL 
		  AND yk_payment_method_id IS NOT NULL`

	// Добавляем отладочную информацию
	log.Printf("🔍 [SQL DEBUG] GetSubscriptionsDueForRetry query: %s", query)
	log.Printf("🔍 [SQL DEBUG] Current time (NOW()): %s", time.Now().Format("2006-01-02 15:04:05"))
	log.Printf("🔍 [SQL DEBUG] Current time (UTC): %s", time.Now().UTC().Format("2006-01-02 15:04:05"))

	// Проверяем время в базе данных
	var dbNow time.Time
	err := r.db.QueryRow("SELECT NOW()").Scan(&dbNow)
	if err != nil {
		log.Printf("⚠️ [SQL DEBUG] Error getting DB time: %v", err)
	} else {
		log.Printf("🔍 [SQL DEBUG] Database time (NOW()): %s", dbNow.Format("2006-01-02 15:04:05"))
		log.Printf("🔍 [SQL DEBUG] Database time (UTC): %s", dbNow.UTC().Format("2006-01-02 15:04:05"))
	}

	rows, err := r.db.Query(query)
	if err != nil {
		log.Printf("❌ [SQL DEBUG] Query error: %v", err)
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
			&subscription.FailedAttempts,
			&subscription.NextRetry,
			&subscription.SuspendedAt,
		)
		if err != nil {
			log.Printf("❌ [SQL DEBUG] Row scan error: %v", err)
			return nil, err
		}
		subscriptions = append(subscriptions, subscription)
		log.Printf("🔍 [SQL DEBUG] Found retry subscription: ID=%d, UserID=%d, FailedAttempts=%d, NextRetry=%s",
			subscription.ID, subscription.UserID, subscription.FailedAttempts,
			subscription.NextRetry.Format("2006-01-02 15:04:05"))
	}

	log.Printf("🔍 [SQL DEBUG] Total retry subscriptions found: %d", len(subscriptions))
	return subscriptions, nil
}

// IncrementFailedAttempts увеличивает счетчик неудачных попыток
func (r *SubscriptionRepository) IncrementFailedAttempts(userID int64) error {
	query := `UPDATE subscriptions SET failed_attempts = failed_attempts + 1 WHERE user_id = $1 AND active = true`
	_, err := r.db.Exec(query, userID)
	return err
}

// SuspendSubscription приостанавливает подписку после 3 неудачных попыток
func (r *SubscriptionRepository) SuspendSubscription(userID int64) error {
	now := time.Now().UTC() // Используем UTC время
	query := `UPDATE subscriptions SET 
		status = 'suspended',
		suspended_at = $1,
		next_retry = NULL
		WHERE user_id = $2 AND active = true`
	_, err := r.db.Exec(query, now, userID)
	return err
}

// GetAllActiveSubscriptions получает все активные подписки для диагностики
func (r *SubscriptionRepository) GetAllActiveSubscriptions() ([]*domain.Subscription, error) {
	query := `
		SELECT id, user_id, subscription_id, tariff, status, amount, next_payment, last_payment, created_at, cancelled_at, active,
		       yk_customer_id, yk_payment_method_id, yk_last_payment_id, failed_attempts, next_retry, suspended_at
		FROM subscriptions
		WHERE active = true 
		ORDER BY next_payment ASC`

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
			&subscription.FailedAttempts,
			&subscription.NextRetry,
			&subscription.SuspendedAt,
		)
		if err != nil {
			return nil, err
		}
		subscriptions = append(subscriptions, subscription)
	}

	return subscriptions, nil
}
