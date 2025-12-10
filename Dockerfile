# Stage 1: Сборка
FROM golang:1.23-alpine AS builder

WORKDIR /build

# Копируем go.mod и загружаем зависимости
COPY go.mod ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем статический бинарник
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o passwordgen ./cmd/passwordgen

# Stage 2: Финальный образ
FROM alpine:latest

WORKDIR /

# Копируем бинарник из builder
COPY --from=builder /build/passwordgen /passwordgen

# Устанавливаем точку входа
ENTRYPOINT ["/passwordgen"]

# Пустой CMD позволяет передавать аргументы при запуске
CMD []

