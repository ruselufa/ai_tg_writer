package monitoring

import (
	"sync"
	"time"
)

// ActiveUsersManager управляет активными пользователями
type ActiveUsersManager struct {
	activeUsers map[int64]time.Time
	mutex       sync.RWMutex
	timeout     time.Duration
}

var activeUsersManager *ActiveUsersManager

// InitActiveUsersManager инициализирует менеджер активных пользователей
func InitActiveUsersManager(timeout time.Duration) {
	activeUsersManager = &ActiveUsersManager{
		activeUsers: make(map[int64]time.Time),
		timeout:     timeout,
	}

	// Запускаем горутину для очистки неактивных пользователей
	go activeUsersManager.cleanupInactiveUsers()
}

// MarkUserActive отмечает пользователя как активного
func (aum *ActiveUsersManager) MarkUserActive(userID int64) {
	aum.mutex.Lock()
	defer aum.mutex.Unlock()

	aum.activeUsers[userID] = time.Now()
}

// GetActiveUsersCount возвращает количество активных пользователей
func (aum *ActiveUsersManager) GetActiveUsersCount() int {
	aum.mutex.RLock()
	defer aum.mutex.RUnlock()

	now := time.Now()
	count := 0

	for _, lastActivity := range aum.activeUsers {
		if now.Sub(lastActivity) <= aum.timeout {
			count++
		}
	}

	return count
}

// cleanupInactiveUsers периодически очищает неактивных пользователей
func (aum *ActiveUsersManager) cleanupInactiveUsers() {
	ticker := time.NewTicker(5 * time.Second) // Проверяем каждые 5 секунд
	defer ticker.Stop()

	for range ticker.C {
		aum.mutex.Lock()
		now := time.Now()

		for userID, lastActivity := range aum.activeUsers {
			if now.Sub(lastActivity) > aum.timeout {
				delete(aum.activeUsers, userID)
			}
		}

		// Обновляем метрику
		activeCount := 0
		for _, lastActivity := range aum.activeUsers {
			if now.Sub(lastActivity) <= aum.timeout {
				activeCount++
			}
		}

		SetActiveTelegramUsers(activeCount)
		aum.mutex.Unlock()
	}
}

// MarkUserActiveGlobal - глобальная функция для отметки пользователя как активного
func MarkUserActiveGlobal(userID int64) {
	if activeUsersManager != nil {
		activeUsersManager.MarkUserActive(userID)
	}
}

// GetActiveUsersCountGlobal - глобальная функция для получения количества активных пользователей
func GetActiveUsersCountGlobal() int {
	if activeUsersManager != nil {
		return activeUsersManager.GetActiveUsersCount()
	}
	return 0
}
