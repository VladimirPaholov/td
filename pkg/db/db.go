package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

var db *sql.DB

const schema = `
CREATE TABLE IF NOT EXISTS scheduler (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date CHAR(8) NOT NULL,
    title VARCHAR(255) NOT NULL,
    comment TEXT,
    repeat VARCHAR(128) NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_scheduler_date ON scheduler(date);
`

// Init открывает базу данных и создает таблицу/индекс при первом запуске
func Init(dbFile string) error {
	install := false

	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		install = true
	}

	d, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return fmt.Errorf("ошибка при открытии базы: %w", err)
	}

	if install {
		if _, err := d.Exec(schema); err != nil {
			d.Close()
			return fmt.Errorf("ошибка при создании схемы: %w", err)
		}
	}

	db = d
	return nil
}

// GetDB возвращает глобальное подключение к базе
func GetDB() *sql.DB {
	return db
}
