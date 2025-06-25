package main

import (
	handlers "github.com/Extrig/order-service/internal/handler"
	"github.com/Extrig/order-service/internal/kafka"
	"net/http"
	"os"
	"path/filepath"

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

	// Восстановление кеша
	if err := db.LoadCacheFromDB(); err != nil {
		log.Err(err).Msg("❌ Ошибка при загрузке кеша")
	}

	//Запускаем kafka-consumer
	go kafka.StartConsumer()

	// Инициализируем роутер
	r := mux.NewRouter()
	r.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	}).Methods("GET")

	r.HandleFunc("/order/{uid}", handlers.GetOrderHandler).Methods("GET")

	// Путь до frontend
	frontendPath := "/frontend"

	//Для статики
	r.PathPrefix("/").Handler(http.FileServer(http.Dir(frontendPath)))

	// Главная страница — index.html
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(frontendPath, "index.html"))
	})

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
