package kafka

import (
	"context"
	"encoding/json"
	"github.com/Extrig/order-service/internal/db"
	models "github.com/Extrig/order-service/internal/model"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
	"os"
	"time"
)

func StartConsumer() {
	broker := os.Getenv("KAFKA_ADDR")
	if broker == "" {
		broker = "kafka:9092"
	}
	topic := "orders"
	groupID := "order-consumer-group"

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{broker},
		GroupID:     groupID,
		Topic:       topic,
		StartOffset: kafka.FirstOffset,
		MinBytes:    1e3,  // 1KB
		MaxBytes:    10e6, // 10MB
		MaxWait:     1 * time.Second,
	})
	log.Info().Msg("📥 Консьюмер Kafka запущен, ожидаем сообщения...")

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			log.Error().Err(err).Msg("Ошибка чтения сообщения из Kafka")
			continue
		}
		log.Info().Str("key", string(m.Key)).Msg("🔔 Получено новое сообщение")

		var order models.Order
		if err := json.Unmarshal(m.Value, &order); err != nil {
			log.Error().Err(err).Msg("Ошибка при разборе JSON заказа")
			continue
		}

		err = db.SaveOrder(order)
		if err != nil {
			log.Error().Err(err).Msg("Ошибка при сохранении заказа в БД")
			continue
		}

		log.Info().Str("order_uid", order.OrderUID).Msg("✅ Заказ успешно сохранён")
	}
}
