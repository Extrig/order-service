# Имя бинарника
APP_NAME=order-service

# Путь к main
MAIN_PATH=./cmd/main.go

# Билд Go-бинарника локально
build:
	go build -o $(APP_NAME) $(MAIN_PATH)

# Запуск локально (если не в Docker)
run:
	go run $(MAIN_PATH)

# Запуск всех контейнеров с пересборкой
up:
	docker compose up --build

# Остановка и удаление всех volume'ов (чистый старт)
down:
	docker compose down -v

# Только пересобрать backend образ
rebuild:
	docker compose build backend

# Применить SQL-инициализацию вручную
migrate:
	docker exec -i order-postgres psql -U order_user -d order_service < scripts/init_db.sql

# Проверка статуса контейнеров
ps:
	docker compose ps

# Просмотр логов backend
logs:
	docker compose logs -f backend
