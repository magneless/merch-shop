package repository

import (
	"database/sql"
	"fmt"
)

type Repository struct {
	db *sql.DB
}

func New(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetUser(username, passwordHash string) error {
	const op = "repository.GetUser"

	var realPasswordHash string
	err := r.db.QueryRow(
		"SELECT password_hash FROM employees WHERE username = $1",
		username,
	).Scan(&realPasswordHash)

	switch {
	case err == sql.ErrNoRows:
		_, err = r.db.Exec(
			"INSERT INTO employees (username, password_hash, balance) VALUES ($1, $2, 1000)",
			username, passwordHash,
		)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	case err != nil:
		return fmt.Errorf("%s: %w", op, err)
	default:
		if passwordHash != realPasswordHash {
			return fmt.Errorf("%s: wrong password", op)
		}
	}

	return nil
}
