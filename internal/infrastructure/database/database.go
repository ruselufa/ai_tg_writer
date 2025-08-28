package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

type DB struct {
	*sql.DB
}

type User struct {
	ID           int64
	Username     string
	FirstName    string
	LastName     string
	Email        string
	Tariff       string
	UsageCount   int
	LastUsage    time.Time
	CreatedAt    time.Time
	ReferralCode string
	ReferredBy   *int64
}

type VoiceMessage struct {
	ID        int64
	UserID    int64
	FileID    string
	Duration  int
	FileSize  int
	Text      string
	Rewritten string
	CreatedAt time.Time
}

// NewConnection создает новое подключение к базе данных
func NewConnection() (*DB, error) {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "")
	dbname := getEnv("DB_NAME", "ai_tg_writer")

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	// Применяем миграции
	migrationsDir := filepath.Join("migrations")
	if err := goose.SetDialect("postgres"); err != nil {
		return nil, fmt.Errorf("ошибка установки диалекта: %v", err)
	}

	if err := goose.Up(db, migrationsDir); err != nil {
		return nil, fmt.Errorf("ошибка применения миграций: %v", err)
	}

	log.Println("Успешно подключились к базе данных и применили миграции")
	return &DB{db}, nil
}

// InitTables создает необходимые таблицы
func (db *DB) InitTables() error {
	// Таблица пользователей
	createUsersTable := `
    CREATE TABLE IF NOT EXISTS users (
        id BIGINT PRIMARY KEY,
        username VARCHAR(255),
        is_admin BOOLEAN DEFAULT FALSE,
		first_name VARCHAR(255),
		last_name VARCHAR(255),
		email VARCHAR(255),
		tariff VARCHAR(50) DEFAULT 'free',
		usage_count INTEGER DEFAULT 0,
		last_usage TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		referral_code VARCHAR(20) UNIQUE,
		referred_by BIGINT REFERENCES users(id)
	);`

	// Таблица голосовых сообщений
	createVoiceMessagesTable := `
	CREATE TABLE IF NOT EXISTS voice_messages (
		id SERIAL PRIMARY KEY,
		user_id BIGINT REFERENCES users(id),
		file_id VARCHAR(255),
		duration INTEGER,
		file_size INTEGER,
		text TEXT,
		rewritten TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	// Таблица статистики использования
	createUsageStatsTable := `
	CREATE TABLE IF NOT EXISTS usage_stats (
		id SERIAL PRIMARY KEY,
		user_id BIGINT REFERENCES users(id),
		date DATE DEFAULT CURRENT_DATE,
		usage_count INTEGER DEFAULT 0,
		UNIQUE(user_id, date)
	);`

	_, err := db.Exec(createUsersTable)
	if err != nil {
		return fmt.Errorf("ошибка создания таблицы users: %v", err)
	}

	_, err = db.Exec(createVoiceMessagesTable)
	if err != nil {
		return fmt.Errorf("ошибка создания таблицы voice_messages: %v", err)
	}

	_, err = db.Exec(createUsageStatsTable)
	if err != nil {
		return fmt.Errorf("ошибка создания таблицы usage_stats: %v", err)
	}

	log.Println("Таблицы успешно созданы")
	return nil
}

// GetOrCreateUser получает пользователя или создает нового
func (db *DB) GetOrCreateUser(userID int64, username, firstName, lastName string) (*User, error) {
	// Пытаемся найти пользователя
	user := &User{}
	err := db.QueryRow(`
		SELECT id, username, first_name, last_name, tariff, usage_count, last_usage, created_at, referral_code, referred_by
		FROM users WHERE id = $1`, userID).Scan(
		&user.ID, &user.Username, &user.FirstName, &user.LastName, &user.Tariff,
		&user.UsageCount, &user.LastUsage, &user.CreatedAt, &user.ReferralCode, &user.ReferredBy)

	if err == sql.ErrNoRows {
		// Создаем нового пользователя
		referralCode := generateReferralCode()
		_, err = db.Exec(`
			INSERT INTO users (id, username, first_name, last_name, referral_code)
			VALUES ($1, $2, $3, $4, $5)`,
			userID, username, firstName, lastName, referralCode)
		if err != nil {
			return nil, err
		}

		// Получаем созданного пользователя
		err = db.QueryRow(`
			SELECT id, username, first_name, last_name, tariff, usage_count, last_usage, created_at, referral_code, referred_by
			FROM users WHERE id = $1`, userID).Scan(
			&user.ID, &user.Username, &user.FirstName, &user.LastName, &user.Tariff,
			&user.UsageCount, &user.LastUsage, &user.CreatedAt, &user.ReferralCode, &user.ReferredBy)
	}

	return user, err
}

// GetUserUsageTotal получает общее количество использований пользователя
func (db *DB) GetUserUsageTotal(userID int64) (int, error) {
	var count int
	err := db.QueryRow(`
		SELECT COALESCE(SUM(usage_count), 0) FROM usage_stats 
		WHERE user_id = $1`, userID).Scan(&count)

	if err != nil {
		return 0, nil
	}
	return count, err
}

// GetUserUsageThisMonth получает количество использований пользователя за текущий месяц
func (db *DB) GetUserUsageThisMonth(userID int64) (int, error) {
	var count int
	err := db.QueryRow(`
		SELECT COALESCE(SUM(usage_count), 0) FROM usage_stats 
		WHERE user_id = $1 
		AND date >= date_trunc('month', CURRENT_DATE)
		AND date < date_trunc('month', CURRENT_DATE) + INTERVAL '1 month'`, userID).Scan(&count)

	if err != nil {
		return 0, nil
	}
	return count, err
}

// GetUserUsageToday получает количество использований пользователя сегодня
func (db *DB) GetUserUsageToday(userID int64) (int, error) {
	return db.GetUserUsageTotal(userID)
}

// IncrementUsage увеличивает счетчик использований
func (db *DB) IncrementUsage(userID int64) error {
	// Обновляем или создаем запись в статистике
	_, err := db.Exec(`
		INSERT INTO usage_stats (user_id, usage_count) 
		VALUES ($1, 1)
		ON CONFLICT (user_id, date) 
		DO UPDATE SET usage_count = usage_stats.usage_count + 1`, userID)

	if err != nil {
		return err
	}

	// Обновляем последнее использование
	_, err = db.Exec(`
		UPDATE users SET last_usage = CURRENT_TIMESTAMP 
		WHERE id = $1`, userID)

	return err
}

// SaveVoiceMessage сохраняет информацию о голосовом сообщении
func (db *DB) SaveVoiceMessage(userID int64, fileID string, duration, fileSize int, text, rewritten string) error {
	_, err := db.Exec(`
		INSERT INTO voice_messages (user_id, file_id, duration, file_size, text, rewritten)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		userID, fileID, duration, fileSize, text, rewritten)
	return err
}

// GetUserTariff получает тариф пользователя
func (db *DB) GetUserTariff(userID int64) (string, error) {
	var tariff string
	err := db.QueryRow(`SELECT tariff FROM users WHERE id = $1`, userID).Scan(&tariff)
	return tariff, err
}

// UpdateUserTariff обновляет тариф пользователя
func (db *DB) UpdateUserTariff(userID int64, tariff string) error {
	_, err := db.Exec(`UPDATE users SET tariff = $1 WHERE id = $2`, tariff, userID)
	return err
}

// UpdateUserEmail обновляет email пользователя
func (db *DB) UpdateUserEmail(userID int64, email string) error {
	_, err := db.Exec(`UPDATE users SET email=$1 WHERE id=$2`, email, userID)
	return err
}

// IsAdmin проверяет, является ли пользователь администратором
func (db *DB) IsAdmin(userID int64) (bool, error) {
	// Получаем список ID администраторов из переменной окружения
	adminIDs := os.Getenv("ADMIN_TELEGRAM_IDS")
	if adminIDs == "" {
		return false, nil
	}

	// Разбиваем строку на отдельные ID
	idStrs := strings.Split(adminIDs, ",")
	for _, idStr := range idStrs {
		idStr = strings.TrimSpace(idStr)
		if idStr == "" {
			continue
		}

		// Парсим ID и сравниваем
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			continue
		}
		if id == userID {
			return true, nil
		}
	}

	return false, nil
}

// Вспомогательные функции
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func generateReferralCode() string {
	// Простая генерация кода из 8 символов
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, 8)
	for i := range result {
		result[i] = chars[time.Now().UnixNano()%int64(len(chars))]
	}
	return string(result)
}
