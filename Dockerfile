# Используем официальный Go образ
FROM golang:1.23-alpine AS builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем go.mod и go.sum
COPY go.mod go.sum ./

# Скачиваем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/ai_tg_writer/main.go

# Финальный образ
FROM alpine:latest

# Устанавливаем ca-certificates для HTTPS запросов
RUN apk --no-cache add ca-certificates

# Создаем пользователя для безопасности
RUN adduser -D -s /bin/sh appuser

# Устанавливаем рабочую директорию
WORKDIR /root/

# Копируем собранное приложение
COPY --from=builder /app/main .

# Копируем необходимые файлы
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/monitoring ./monitoring

# Создаем папку для аудио
RUN mkdir -p audio && chown appuser:appuser audio

# Устанавливаем права доступа для migrations
RUN chown -R appuser:appuser migrations

# Временно запускаем от root для отладки
# USER appuser

# Открываем порт
EXPOSE 8080

# Запускаем приложение
CMD ["./main"]