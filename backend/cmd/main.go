package main

import (
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
	database, err := db.InitPostgres()
	if err != nil {
		log.Fatal().Err(err).Msg("❌ Не удалось подключиться к БД")
	}
	defer database.Close()

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
