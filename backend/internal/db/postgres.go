package db

import (
	"context"
	"fmt"
	"github.com/Extrig/order-service/internal/cache"
	"github.com/Extrig/order-service/internal/model"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func InitPostgres() error {
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		return fmt.Errorf("DB_URL не задан в .env")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	for i := 1; i <= 10; i++ {
		DB, err = pgxpool.New(ctx, dbURL)
		if err == nil {
			err = DB.Ping(ctx)
			if err == nil {
				fmt.Println("✅ Подключение к PostgreSQL установлено")
				return nil
			}
		}
		fmt.Printf("⏳ Попытка %d: ошибка подключения к БД: %v\n", i, err)
		time.Sleep(2 * time.Second)
	}

	return fmt.Errorf("❌ не удалось подключиться к БД после 10 попыток: %v", err)
}

func SaveOrder(order model.Order) error {
	//Создаёт контекст с таймаутом 5 секунд — на случай, если БД долго отвечает или зависает
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	//Начинаем транзакцию
	tx, err := DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("❌ ошибка при начале транзакции: %w", err)
	}
	//Делаем rollback в случае ошибки
	defer tx.Rollback(ctx)

	// Сохраняем order
	_, err = tx.Exec(ctx, `
		INSERT INTO orders (order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`,
		order.OrderUID,
		order.TrackNumber,
		order.Entry,
		order.Locale,
		order.InternalSignature,
		order.CustomerID,
		order.DeliveryService,
		order.ShardKey,
		order.SMID,
		order.DateCreated,
		order.OOFShard,
	)
	if err != nil {
		return fmt.Errorf("❌ ошибка при вставке order: %w", err)
	}

	// Сохраняем delivery
	_, err = tx.Exec(ctx, `
		INSERT INTO delivery (order_uid, name, phone, zip, city, address, region, email)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`,
		order.OrderUID,
		order.Delivery.Name,
		order.Delivery.Phone,
		order.Delivery.Zip,
		order.Delivery.City,
		order.Delivery.Address,
		order.Delivery.Region,
		order.Delivery.Email,
	)
	if err != nil {
		return fmt.Errorf("❌ ошибка при вставке delivery: %w", err)
	}

	// Сохраняем payment
	_, err = tx.Exec(ctx, `
		INSERT INTO payment (order_uid, transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`,
		order.OrderUID,
		order.Payment.Transaction,
		order.Payment.RequestID,
		order.Payment.Currency,
		order.Payment.Provider,
		order.Payment.Amount,
		order.Payment.PaymentDT,
		order.Payment.Bank,
		order.Payment.DeliveryCost,
		order.Payment.GoodsTotal,
		order.Payment.CustomFee,
	)
	if err != nil {
		return fmt.Errorf("❌ ошибка при вставке payment: %w", err)
	}

	// Сохраняем items
	for _, item := range order.Items {
		_, err = tx.Exec(ctx, `
			INSERT INTO items (order_uid, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		`,
			order.OrderUID,
			item.ChrtID,
			item.TrackNumber,
			item.Price,
			item.RID,
			item.Name,
			item.Sale,
			item.Size,
			item.TotalPrice,
			item.NMID,
			item.Brand,
			item.Status,
		)
		if err != nil {
			return fmt.Errorf("❌ ошибка при вставке item: %w", err)
		}
	}

	// Коммитим
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("❌ ошибка при коммите транзакции: %w", err)
	}
	cache.Set(order)
	return nil
}

func GetOrderById(orderUID string) (model.Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var order model.Order

	// 1. Получаем данные заказа
	err := DB.QueryRow(ctx, `
		SELECT order_uid, track_number, entry, locale, internal_signature,
		       customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
		FROM orders
		WHERE order_uid = $1
	`, orderUID).Scan(
		&order.OrderUID,
		&order.TrackNumber,
		&order.Entry,
		&order.Locale,
		&order.InternalSignature,
		&order.CustomerID,
		&order.DeliveryService,
		&order.ShardKey,
		&order.SMID,
		&order.DateCreated,
		&order.OOFShard,
	)
	if err != nil {
		return order, fmt.Errorf("❌ ошибка при получении заказа: %w", err)
	}

	// 2. Получаем доставку
	err = DB.QueryRow(ctx, `
		SELECT name, phone, zip, city, address, region, email
		FROM delivery
		WHERE order_uid = $1
	`, orderUID).Scan(
		&order.Delivery.Name,
		&order.Delivery.Phone,
		&order.Delivery.Zip,
		&order.Delivery.City,
		&order.Delivery.Address,
		&order.Delivery.Region,
		&order.Delivery.Email,
	)
	if err != nil {
		return order, fmt.Errorf("❌ ошибка при получении доставки: %w", err)
	}

	// 3. Получаем оплату
	err = DB.QueryRow(ctx, `
		SELECT transaction, request_id, currency, provider, amount, payment_dt,
		       bank, delivery_cost, goods_total, custom_fee
		FROM payment
		WHERE order_uid = $1
	`, orderUID).Scan(
		&order.Payment.Transaction,
		&order.Payment.RequestID,
		&order.Payment.Currency,
		&order.Payment.Provider,
		&order.Payment.Amount,
		&order.Payment.PaymentDT,
		&order.Payment.Bank,
		&order.Payment.DeliveryCost,
		&order.Payment.GoodsTotal,
		&order.Payment.CustomFee,
	)
	if err != nil {
		return order, fmt.Errorf("❌ ошибка при получении оплаты: %w", err)
	}

	// 4. Получаем все товары
	rows, err := DB.Query(ctx, `
		SELECT chrt_id, track_number, price, rid, name, sale,
		       size, total_price, nm_id, brand, status
		FROM items
		WHERE order_uid = $1
	`, orderUID)
	if err != nil {
		return order, fmt.Errorf("❌ ошибка при получении товаров: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item model.Item
		err := rows.Scan(
			&item.ChrtID,
			&item.TrackNumber,
			&item.Price,
			&item.RID,
			&item.Name,
			&item.Sale,
			&item.Size,
			&item.TotalPrice,
			&item.NMID,
			&item.Brand,
			&item.Status,
		)
		if err != nil {
			return order, fmt.Errorf("❌ ошибка при сканировании товара: %w", err)
		}
		order.Items = append(order.Items, item)
	}

	return order, nil
}
