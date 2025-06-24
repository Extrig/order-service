package db

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // регистрирует "pgx" как драйвер
)

func InitPostgres() (*sql.DB, error) {
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DB_URL не задан в .env")
	}

	var db *sql.DB
	var err error

	for i := 1; i <= 10; i++ {
		db, err = sql.Open("pgx", dbURL) // теперь используем "pgx", не "postgres"
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
