FROM golang:1.24

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем модули и устанавливаем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходники
COPY . .

#Создаём директорию для логов
RUN mkdir -p logs

# Сборка бинарника
RUN go build -o order-service ./cmd/main.go

# Установка dockerize (ждать готовности сервисов)
RUN wget https://github.com/jwilder/dockerize/releases/download/v0.6.1/dockerize-linux-amd64-v0.6.1.tar.gz && \
    tar -C /usr/local/bin -xzvf dockerize-linux-amd64-v0.6.1.tar.gz && \
    rm dockerize-linux-amd64-v0.6.1.tar.gz

# Порт HTTP API
EXPOSE 8081

# Команда запуска: ждем postgres и kafka, затем запускаем
CMD ["dockerize", "-wait", "tcp://postgres:5432", "-wait", "tcp://kafka:9092", "-timeout", "30s", "./order-service"]
