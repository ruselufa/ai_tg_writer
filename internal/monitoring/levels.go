package monitoring

import (
	"log"
	"os"
	"strings"
)

// Уровни логирования
const (
	LevelDebug = "DEBUG"
	LevelInfo  = "INFO"
	LevelWarn  = "WARN"
	LevelError = "ERROR"
)

var (
	currentLevel = LevelInfo
	debugEnabled = false
)

// InitLogging инициализирует систему логирования
func InitLogging() {
	// Проверяем переменную окружения для уровня логирования
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		currentLevel = strings.ToUpper(level)
	}

	// Включаем debug если уровень DEBUG
	debugEnabled = (currentLevel == LevelDebug)

	// Настраиваем стандартный логгер
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

// Debug логирует отладочную информацию
func Debug(format string, v ...interface{}) {
	if debugEnabled {
		log.Printf("[DEBUG] "+format, v...)
	}
}

// Info логирует информационные сообщения
func Info(format string, v ...interface{}) {
	if currentLevel == LevelDebug || currentLevel == LevelInfo {
		log.Printf("[INFO] "+format, v...)
	}
}

// Warn логирует предупреждения
func Warn(format string, v ...interface{}) {
	if currentLevel == LevelDebug || currentLevel == LevelInfo || currentLevel == LevelWarn {
		log.Printf("[WARN] "+format, v...)
	}
}

// Error логирует ошибки
func Error(format string, v ...interface{}) {
	log.Printf("[ERROR] "+format, v...)
}

// System логирует системные сообщения (всегда показываются)
func System(format string, v ...interface{}) {
	log.Printf("[SYSTEM] "+format, v...)
}

// IsDebugEnabled возвращает true если включен debug режим
func IsDebugEnabled() bool {
	return debugEnabled
}
