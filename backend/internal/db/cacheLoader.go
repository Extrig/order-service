package db

import (
	"context"
	"fmt"
	"github.com/Extrig/order-service/internal/cache"
	"github.com/Extrig/order-service/internal/model"
	"time"
)

func LoadCacheFromDB() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rows, err := DB.Query(ctx, `SELECT order_uid FROM orders`)
	if err != nil {
		return fmt.Errorf("❌ не удалось получить список заказов: %w", err)
	}
	defer rows.Close()

	cached := make(map[string]model.Order)
	for rows.Next() {
		var uid string
		if err := rows.Scan(&uid); err != nil {
			continue
		}

		order, err := GetOrderById(uid)
		if err == nil {
			cached[uid] = order
		}
	}

	cache.SetAll(cached)
	fmt.Printf("📦 Кеш загружен: %d заказов\n", len(cached))
	return nil
}
