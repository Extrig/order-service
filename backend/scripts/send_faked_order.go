package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/segmentio/kafka-go"
)

// --- Структуры с тегами faker --- //

type Delivery struct {
	Name    string `json:"name" faker:"russian_first_name_male"`
	Phone   string `json:"phone" faker:"phone_number"`
	Zip     string
	City    string `json:"city"`
	Address string `json:"address"`
	Region  string `json:"region"`
	Email   string `json:"email" faker:"email"`
}

type Payment struct {
	Transaction  string `json:"transaction" faker:"uuid_hyphenated"`
	RequestID    string `json:"request_id" faker:"uuid_hyphenated"`
	Currency     string `json:"currency"`
	Provider     string `json:"provider" faker:"cc_type"`
	Amount       int    `json:"amount"`
	PaymentDT    int64  `json:"payment_dt"`
	Bank         string `json:"bank" faker:"oneof: Сбербанк, Т-банк, Альфа-банк"`
	DeliveryCost int    `json:"delivery_cost"`
	GoodsTotal   int    `json:"goods_total"`
	CustomFee    int    `json:"custom_fee"`
}

type Item struct {
	ChrtID      int    `json:"chrt_id"`
	TrackNumber string `json:"track_number"`
	Price       int    `json:"price"`
	RID         string `json:"rid" faker:"uuid_hyphenated"`
	Name        string `json:"name"`
	Sale        int    `json:"sale"`
	Size        string `json:"size"`
	TotalPrice  int    `json:"total_price"`
	NMID        int    `json:"nm_id"`
	Brand       string `json:"brand"`
	Status      int    `json:"status"`
}

type Order struct {
	OrderUID          string   `json:"order_uid" faker:"uuid_hyphenated"`
	TrackNumber       string   `json:"track_number" faker:"word"`
	Entry             string   `json:"entry" faker:"word"`
	Delivery          Delivery `json:"delivery"`
	Payment           Payment  `json:"payment"`
	Items             []Item   `json:"items"`
	Locale            string   `json:"locale" faker:"oneof: ru"`
	InternalSignature string   `json:"internal_signature" faker:"uuid_hyphenated"`
	CustomerID        string   `json:"customer_id" faker:"username"`
	DeliveryService   string   `json:"delivery_service" faker:"oneof: CDEK, WB, Avito"`
	ShardKey          string   `json:"shardkey" faker:"oneof: 1,2,3,4,5,6,7,8,9"`
	SMID              int      `json:"sm_id"`
	DateCreated       string   `json:"date_created"`
	OOFShard          string   `json:"oof_shard" faker:"oneof: 1,2,3"`
}

type FakeItem struct {
	RID   string `json:"rid" faker:"uuid_hyphenated"`
	Name  string `faker:"oneof: Ноутбук, Телефон, Телевизор, Часы"`
	Brand string `faker:"oneof: LG, SAMSUNG, Apple, Xiaomi"`
}

// --- Генерация одного заказа --- //

func generateFakeOrder(random *rand.Rand) Order {
	var order Order
	if err := faker.FakeData(&order); err != nil {
		log.Fatalf("❌Ошибка генерации заказа: %v", err)
	}

	// Генерируем базовые параметры товара
	price := random.Intn(900) + 100
	sale := random.Intn(50)
	total := price - (price * sale / 100)
	deliveryCost := 300
	customFee := random.Intn(100)

	// Устанавливаем фиксированные значения после генерации faker'ом
	order.SMID = random.Intn(100)
	order.DateCreated = time.Now().Format(time.RFC3339)

	fakeAddress := faker.GetRealAddress()
	order.Delivery.Zip = fakeAddress.PostalCode
	order.Delivery.City = fakeAddress.City
	order.Delivery.Region = fakeAddress.State

	// Настраиваем платеж
	order.Payment.Currency = "RUB" // Фиксированная валюта для консистентности
	order.Payment.Amount = total + deliveryCost + customFee
	order.Payment.GoodsTotal = total
	order.Payment.DeliveryCost = deliveryCost
	order.Payment.CustomFee = customFee
	order.Payment.PaymentDT = time.Now().Unix()

	// Создаем товар с правильными значениями
	item := Item{
		ChrtID:      random.Intn(9999999),
		TrackNumber: order.TrackNumber,
		Price:       price,
		Sale:        sale,
		Size:        "0",
		TotalPrice:  total,
		NMID:        random.Intn(9999999),
		Status:      202,
	}

	// Генерируем только те поля, которые нам нужны от faker'а
	var fakeItem FakeItem
	if err := faker.FakeData(&fakeItem); err != nil {
		log.Printf("‼️Предупреждение: не удалось сгенерировать fakeItem: %v", err)
	} else {
		item.RID = fakeItem.RID
		item.Name = fakeItem.Name
		item.Brand = fakeItem.Brand
	}

	order.Items = []Item{item}

	return order
}

func main() {
	// Создаём локальный генератор случайных чисел
	seed := time.Now().UnixNano()
	random := rand.New(rand.NewSource(seed))

	n := 1
	if len(os.Args) > 1 {
		if val, err := strconv.Atoi(os.Args[1]); err == nil {
			n = val
		}
	}

	addr := os.Getenv("KAFKA_ADDR")
	if addr == "" {
		addr = "kafka:9092"
	}

	topic := os.Getenv("KAFKA_TOPIC")
	if topic == "" {
		topic = "orders"
	}

	writer := kafka.Writer{
		Addr:     kafka.TCP(addr),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
	defer writer.Close()

	log.Printf("Начинаем отправку %d заказов в Kafka (%s)", n, addr)

	var messages []kafka.Message
	for i := 0; i < n; i++ {
		order := generateFakeOrder(random)
		data, err := json.Marshal(order)
		if err != nil {
			log.Printf("❌Ошибка сериализации #%d: %v", i+1, err)
			continue
		}

		messages = append(messages, kafka.Message{
			Key:   []byte(order.OrderUID),
			Value: data,
			Time:  time.Now(),
		})

		log.Printf("✅ [%d/%d] Подготовлен заказ: %s (сумма: %d %s)",
			i+1, n, order.OrderUID, order.Payment.Amount, order.Payment.Currency)

		faker.ResetUnique()
	}

	// Отправляем все заказы одной пачкой
	if err := writer.WriteMessages(context.Background(), messages...); err != nil {
		log.Fatalf("❌ Ошибка при отправке в Kafka: %v", err)
	}

	log.Printf("🔥Завершено. Успешно отправлено: %d/%d заказов", len(messages), n)
}
