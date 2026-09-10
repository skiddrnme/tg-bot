# 1. Базовый образ для СБОРКИ (тяжелый, с компилятором)
FROM golang:1.22-alpine AS builder

# Устанавливаем необходимые пакеты для сборки
RUN apk add --no-cache git

# Создаем рабочую директорию в контейнере
WORKDIR /app

# Копируем файлы зависимостей (для кэширования слоев)
COPY go.mod go.sum ./

# Скачиваем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем бинарник (статически, чтобы работал без Go)
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bot ./cmd/bot/main.go

# 2. Финальный образ (легкий, только бинарник)
FROM alpine:latest

# Устанавливаем CA сертификаты (нужны для HTTPS)
RUN apk --no-cache add ca-certificates

# Создаем непривилегированного пользователя
RUN adduser -D -g '' appuser

WORKDIR /app

# Копируем бинарник из builder-образа
COPY --from=builder /app/bot .
COPY --from=builder /app/.env* ./

# Меняем владельца файлов на appuser
RUN chown -R appuser:appuser /app

# Переключаемся на непривилегированного пользователя
USER appuser

# Открываем порт
EXPOSE 8080

# Команда запуска
CMD ["./bot"]