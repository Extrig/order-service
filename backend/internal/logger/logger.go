package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func InitLogger() {
	// Создаём/открываем файл
	file, err := os.OpenFile("logs/app.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal().Err(err).Msg("не удалось открыть лог-файл")
	}

	// Настройка глобального логгера
	multi := zerolog.MultiLevelWriter(os.Stdout, file)
	zerolog.TimeFieldFormat = time.RFC3339
	log.Logger = zerolog.New(multi).With().Timestamp().Logger()
}
