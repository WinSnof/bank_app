# Этап сборки
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Устанавливаем необходимые пакеты для сборки
RUN apk add --no-cache build-base

# Копируем файлы зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

RUN ls

# Собираем приложение
RUN CGO_ENABLED=1 GOOS=linux go build -o main ./core/app.go

# Финальный этап
FROM alpine:latest

WORKDIR /app

# Копируем бинарный файл из этапа сборки
COPY --from=builder /app/main .

# Устанавливаем необходимые пакеты
RUN apk --no-cache add ca-certificates

EXPOSE 8080

CMD ["./main"]
