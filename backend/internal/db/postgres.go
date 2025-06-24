package db

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq"
)

func InitPostgres() (*sql.DB, error) {
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DB_URL не задан в .env")
	}

	var db *sql.DB
	var err error

	// Повторная попытка до 10 раз
	for i := 1; i <= 10; i++ {
		db, err = sql.Open("postgres", dbURL)
		if err == nil {
			err = db.Ping()
			if err == nil {
				fmt.Println("✅ Подключение к PostgreSQL установлено")
				return db, nil
			}
		}

		fmt.Printf("⏳ Попытка %d: ошибка подключения к БД: %v\n", i, err)
		time.Sleep(2 * time.Second)
	}

	return nil, fmt.Errorf("❌ не удалось подключиться к БД после 10 попыток: %v", err)
}
