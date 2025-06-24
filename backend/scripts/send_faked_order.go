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
	Bank         string `json:"bank" faker:"word"`
	DeliveryCost int    `json:"delivery_cost"`
	GoodsTotal   int    `json:"goods_total"`
	CustomFee    int    `json:"custom_fee"`
}

type Item struct {
	ChrtID      int    `json:"chrt_id"`
	TrackNumber string `json:"track_number"`
	Price       int    `json:"price"`
	RID         string `json:"rid" faker:"uuid_hyphenated"`
	Name        string `json:"name" faker:"word"`
	Sale        int    `json:"sale"`
	Size        string `json:"size"`
	TotalPrice  int    `json:"total_price"`
	NMID        int    `json:"nm_id"`
	Brand       string `json:"brand" faker:"word"`
	Status      int    `json:"status"`
}

type Order struct {
	OrderUID          string   `json:"order_uid" faker:"uuid_hyphenated"`
	TrackNumber       string   `json:"track_number" faker:"word"`
	Entry             string   `json:"entry" faker:"word"`
	Delivery          Delivery `json:"delivery"`
	Payment           Payment  `json:"payment"`
	Items             []Item   `json:"items"`
	Locale            string   `json:"locale" faker:"word"`
	InternalSignature string   `json:"internal_signature" faker:"uuid_hyphenated"`
	CustomerID        string   `json:"customer_id" faker:"username"`
	DeliveryService   string   `json:"delivery_service" faker:"word"`
	ShardKey          string   `json:"shardkey" faker:"oneof: 1,2,3,4,5,6,7,8,9"`
	SMID              int      `json:"sm_id"`
	DateCreated       string   `json:"date_created"`
	OOFShard          string   `json:"oof_shard" faker:"oneof: 1,2,3"`
}

var cities = []string{"Москва", "Санкт-Петербург", "Новосибирск", "Екатеринбург", "Казань"}
var regions = []string{"Московская обл.", "Ленинградская обл.", "Новосибирская обл."}

// --- Генерация одного заказа --- //

func generateFakeOrder() Order {
	var order Order
	if err := faker.FakeData(&order); err != nil {
		log.Fatalf("Ошибка генерации заказа: %v", err)
	}

	// Генерируем базовые параметры товара
	price := rand.Intn(900) + 100
	sale := rand.Intn(50)
	total := price - (price * sale / 100)
	deliveryCost := 300
	customFee := rand.Intn(100)

	// Устанавливаем фиксированные значения после генерации faker'ом
	order.SMID = rand.Intn(100)
	order.DateCreated = time.Now().Format(time.RFC3339)
	/*p := faker.Person{}*/

	// Русское имя и фамилия
	/*firstNameMale, _ := p.RussianFirstNameMale(nil)
	lastNameMale, _ := p.RussianLastNameMale(nil)
	order.Delivery.Name = firstNameMale + " " + lastNameMale*/
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
		ChrtID:      rand.Intn(9999999),
		TrackNumber: order.TrackNumber,
		Price:       price,
		Sale:        sale,
		Size:        "0",
		TotalPrice:  total,
		NMID:        rand.Intn(9999999),
		Status:      202,
	}

	// Генерируем только те поля, которые нам нужны от faker'а
	if err := faker.FakeData(&item.RID); err != nil {
		log.Printf("Предупреждение: не удалось сгенерировать RID: %v", err)
	}
	if err := faker.FakeData(&item.Name); err != nil {
		log.Printf("Предупреждение: не удалось сгенерировать Name: %v", err)
	}
	if err := faker.FakeData(&item.Brand); err != nil {
		log.Printf("Предупреждение: не удалось сгенерировать Brand: %v", err)
	}

	order.Items = []Item{item}

	return order
}

func main() {
	// Инициализируем генератор случайных чисел
	rand.Seed(time.Now().UnixNano())

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

	writer := kafka.Writer{
		Addr:     kafka.TCP(addr),
		Topic:    "orders",
		Balancer: &kafka.LeastBytes{},
	}
	defer writer.Close()

	log.Printf("Начинаем отправку %d заказов в Kafka (%s)", n, addr)

	successCount := 0
	for i := 0; i < n; i++ {
		order := generateFakeOrder()
		data, err := json.Marshal(order)
		if err != nil {
			log.Printf("❌ Ошибка сериализации #%d: %v", i+1, err)
			continue
		}

		err = writer.WriteMessages(context.Background(), kafka.Message{
			Key:   []byte(order.OrderUID),
			Value: data,
			Time:  time.Now(),
		})
		if err != nil {
			log.Printf("❌ Ошибка Kafka #%d: %v", i+1, err)
			continue
		}

		successCount++
		log.Printf("✅ [%d/%d] Заказ отправлен: %s (сумма: %d %s)",
			i+1, n, order.OrderUID, order.Payment.Amount, order.Payment.Currency)
		faker.ResetUnique()
	}

	log.Printf("Завершено. Успешно отправлено: %d/%d заказов", successCount, n)
}
