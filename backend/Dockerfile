FROM golang:1.23-alpine AS dev

WORKDIR /app

RUN apk add --no-cache git python3
RUN GOBIN=/usr/local/bin go install github.com/cosmtrek/air@v1.49.0
RUN GOBIN=/usr/local/bin go install github.com/swaggo/swag/cmd/swag@latest

COPY go.mod go.sum ./
RUN go mod download

CMD ["air", "-c", ".air.docker.toml"]

FROM golang:1.23-alpine AS builder

WORKDIR /app

# Установка зависимостей для сборки
RUN apk add --no-cache git python3

# Устанавливаем swag для генерации Swagger документации
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Копируем go mod файлы
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходники
COPY . .

# Генерируем Swagger документацию
RUN swag init

# Добавляем securityDefinitions
COPY add_security.py .
RUN python3 add_security.py

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Финальный образ
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Копируем бинарник из builder
COPY --from=builder /app/main .

EXPOSE 8000

CMD ["./main"]
