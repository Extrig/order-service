FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o order-service ./cmd/main.go

# --- Финальный слой (только бинарник) ---
FROM debian:bullseye-slim

WORKDIR /app

COPY --from=builder /app/order-service .
COPY --from=builder /app/.env ./
COPY --from=builder /app/logs ./logs

EXPOSE 8081

CMD ["./order-service"]
