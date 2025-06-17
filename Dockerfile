FROM golang:1.23-alpine AS builder

# Установка зависимостей для сборки
RUN apk add --no-cache git ca-certificates build-base

# Создание рабочей директории
WORKDIR /src

# Копируем модули
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Сборка приложения
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /src/bin/sso-service ./cmd/sso-service

# Миграции
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /src/bin/migrate ./cmd/migrate

FROM alpine:3.21

# Копируем бинарник, миграции и env
COPY --from=builder /src/bin/sso-service /usr/local/bin/
COPY --from=builder /src/bin/migrate /usr/local/bin/
COPY --from=builder /src/migrations ./migrations
COPY --from=builder /src/security ./security
COPY --from=builder /src/.env.example .

# Настройки окружения
ENV APP_ENV=example \
    PORT=50051

# Открываем порт
EXPOSE $PORT

ENTRYPOINT [ "/bin/sh", "-c", \
    "echo 'Running migrations...' && \
    migrate && \
    echo 'Starting SSO service...' && \
    sso-service" \
]