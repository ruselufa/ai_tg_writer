FROM golang:1.21-alpine AS builder

# Устанавливаем необходимые пакеты
RUN apk add --no-cache git ca-certificates tzdata

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем файлы зависимостей
COPY go.mod go.sum ./

# Скачиваем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ai_tg_writer main.go

# Финальный образ
FROM alpine:latest

# Устанавливаем ca-certificates для HTTPS запросов
RUN apk --no-cache add ca-certificates tzdata

# Создаем пользователя для безопасности
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем бинарный файл из builder
COPY --from=builder /app/ai_tg_writer .

# Создаем директорию для аудио файлов
RUN mkdir -p audio && chown -R appuser:appgroup /app

# Переключаемся на непривилегированного пользователя
USER appuser

# Открываем порт (если понадобится в будущем)
EXPOSE 8080

# Запускаем приложение
CMD ["./ai_tg_writer"] 