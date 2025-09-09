package monitoring

import (
	"context"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

type Logger struct {
	*logrus.Logger
}

type Fields map[string]interface{}

func NewLogger() *Logger {
	logger := logrus.New()

	// Настройка формата в зависимости от окружения
	if os.Getenv("ENV") == "production" {
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
		logger.SetLevel(logrus.InfoLevel)
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
		logger.SetLevel(logrus.DebugLevel)
	}

	return &Logger{logger}
}

func (l *Logger) WithContext(ctx context.Context) *logrus.Entry {
	entry := l.Logger.WithContext(ctx)

	// Добавляем trace_id если есть
	if traceID := ctx.Value("trace_id"); traceID != nil {
		entry = entry.WithField("trace_id", traceID)
	}

	return entry
}

func (l *Logger) WithUser(userID int64) *logrus.Entry {
	return l.Logger.WithField("user_id", userID)
}

func (l *Logger) WithFields(fields Fields) *logrus.Entry {
	return l.Logger.WithFields(logrus.Fields(fields))
}

