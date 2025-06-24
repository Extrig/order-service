package main

import (
	"github.com/Extrig/order-service/internal/kafka"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	"github.com/Extrig/order-service/internal/db"
	"github.com/Extrig/order-service/internal/logger"
	"github.com/rs/zerolog/log"
)

func main() {
	// Инициализируем zerolog
	logger.InitLogger()
	log.Info().Msg("🟢 Сервис запущен")

	// Загружаем переменные окружения из .env
	if err := godotenv.Load(); err != nil {
		log.Warn().Msg(".env не найден, используются переменные окружения")
	}

	// Подключение к БД
	err := db.InitPostgres()
	if err != nil {
		log.Fatal().Err(err).Msg("❌ Не удалось инициализировать базу данных")
	}
	defer db.DB.Close()

	//Запускаем kafka-consumer
	go kafka.StartConsumer()

	// Инициализируем роутер
	r := mux.NewRouter()
	r.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	}).Methods("GET")

	// Получаем порт из .env
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	log.Info().Msgf("🚀 Сервер запущен на http://localhost:%s", port)

	// Запускаем HTTP-сервер
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal().Err(err).Msg("Сервер завершил работу с ошибкой")
	}
}
