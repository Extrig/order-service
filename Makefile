# Имя бинарника
APP_NAME=order-service

# Путь к main
MAIN_PATH=./backend/cmd/main.go

# ========================
# ======= Go задачи ======
# ========================

build:
	cd backend && go build -o $(APP_NAME) $(MAIN_PATH)

run:
	cd backend && go run $(MAIN_PATH)

# ========================
# ==== Docker окружения ==
# ========================

dev:
	docker compose -f docker-compose.yml -f docker-compose.dev.yml up --build

prod:
	docker compose -f docker-compose.yml -f docker-compose.prod.yml up --build -d

down:
	docker compose -f docker-compose.yml -f docker-compose.dev.yml down -v

rebuild:
	docker compose -f docker-compose.yml -f docker-compose.dev.yml build backend

ps:
	docker compose -f docker-compose.yml -f docker-compose.dev.yml ps

logs:
	docker logs -f order-backend

# ========================
# ====== Утилиты =========
# ========================

migrate:
	docker exec -i order-postgres psql -U order_user -d order_service < backend/scripts/init_db.sql

# Генерация фейковых заказов (make fake-order N=5)
fake-order:
	@echo "📦 Генерация $(N) заказ(ов)..."
	@for i in $(shell seq 1 ${N}); do \
		docker exec -i order-backend go run scripts/send_faked_order.go; \
	done
