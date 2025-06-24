#!/bin/bash
set -e

# Ожидание Kafka (можно настроить ожидание дольше при необходимости)
echo "⏳ Ждём доступности Kafka на kafka:9092..."
sleep 10

# Создание топика, если не существует
echo "📦 Проверяем и создаём топик 'orders'..."
kafka-topics.sh --create \
  --if-not-exists \
  --bootstrap-server kafka:9092 \
  --replication-factor 1 \
  --partitions 1 \
  --topic orders

echo "✅ Топик 'orders' готов."
