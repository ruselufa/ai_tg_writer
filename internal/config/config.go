package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Mode                    string
	SubscriptionInterval    time.Duration
	WorkerCheckInterval     time.Duration
}

// NewConfig создает новую конфигурацию на основе переменных окружения
func NewConfig() *Config {
	mode := getenv("MODE", "production")
	
	var subscriptionInterval time.Duration
	var workerCheckInterval time.Duration
	
	switch mode {
	case "dev", "development":
		subscriptionInterval = 1 * time.Minute  // Для разработки - 1 минута
		workerCheckInterval = 30 * time.Second  // Проверяем каждые 30 секунд
	default: // production
		subscriptionInterval = 30 * 24 * time.Hour // 30 дней
		workerCheckInterval = 5 * time.Minute       // Проверяем каждые 5 минут
	}
	
	return &Config{
		Mode:                    mode,
		SubscriptionInterval:    subscriptionInterval,
		WorkerCheckInterval:     workerCheckInterval,
	}
}

// IsDevMode проверяет, работает ли приложение в режиме разработки
func (c *Config) IsDevMode() bool {
	return c.Mode == "dev" || c.Mode == "development"
}

// GetSubscriptionIntervalDays возвращает интервал подписки в днях
func (c *Config) GetSubscriptionIntervalDays() int {
	return int(c.SubscriptionInterval.Hours() / 24)
}

// getenv возвращает значение переменной окружения или значение по умолчанию
func getenv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getenvInt возвращает целочисленное значение переменной окружения
func getenvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
