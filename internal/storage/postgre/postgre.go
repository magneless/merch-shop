package postgre

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/magneless/merch-shop/internal/config"
)

func New(cfg config.Storage) (*sql.DB, error) {
	const op = "storage.postgre.New"

	db, err := sql.Open("postgres", fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.DBName, cfg.Password, cfg.SSLMode,
	))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}
