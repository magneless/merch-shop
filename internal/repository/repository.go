package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/magneless/merch-shop/internal/models"
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

func (r *Repository) GetBalanceAndId(username string) (int, int, error) {
	const op = "repository.GetBalanceAndId"
	var userID, balance int

	err := r.db.QueryRow("SELECT id, balance FROM employees WHERE username = $1", username).
		Scan(&userID, &balance)
	if err != nil {
		return 0, 0, fmt.Errorf("%s: error fetching user info: %w", op, err)
	}

	return userID, balance, nil
}

func (r *Repository) GetSentTransactions(userID int, fromUsername string) ([]models.CoinTransaction, error) {
	const op = "repository.GetSentTransactions"
	rows, err := r.db.Query(`
		SELECT t.amount, e.username
		FROM transactions t
		JOIN employees e ON t.receiver_id = e.id
		WHERE t.sender_id = $1
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: error fetching sent transactions: %w", op, err)
	}
	defer rows.Close()

	var transactions []models.CoinTransaction
	for rows.Next() {
		var amount int
		var toUsername string
		if err := rows.Scan(&amount, &toUsername); err != nil {
			return nil, fmt.Errorf("%s: error scanning sent transaction: %w", op, err)
		}
		transactions = append(transactions, models.CoinTransaction{
			FromUser: fromUsername,
			ToUser:   toUsername,
			Amount:   amount,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: error iterating sent transactions: %w", op, err)
	}

	return transactions, nil
}

func (r *Repository) GetReceivedTransactions(userID int, toUsername string) ([]models.CoinTransaction, error) {
	const op = "repository.GetReceivedTransactions"
	rows, err := r.db.Query(`
		SELECT t.amount, e.username
		FROM transactions t
		JOIN employees e ON t.sender_id = e.id
		WHERE t.receiver_id = $1
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: error fetching received transactions: %w", op, err)
	}
	defer rows.Close()

	var transactions []models.CoinTransaction
	for rows.Next() {
		var amount int
		var fromUsername string
		if err := rows.Scan(&amount, &fromUsername); err != nil {
			return nil, fmt.Errorf("%s: error scanning received transaction: %w", op, err)
		}
		transactions = append(transactions, models.CoinTransaction{
			FromUser: fromUsername,
			ToUser:   toUsername,
			Amount:   amount,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: error iterating received transactions: %w", op, err)
	}

	return transactions, nil
}

func (r *Repository) GetInventory(userID int) ([]models.InventoryItem, error) {
	const op = "repository.GetInventory"
	rows, err := r.db.Query(`
		SELECT m.merch_name, p.count
		FROM purchases p
		JOIN merch m ON p.merch_id = m.id
		WHERE p.employee_id = $1
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: error fetching inventory: %w", op, err)
	}
	defer rows.Close()

	var inventory []models.InventoryItem
	for rows.Next() {
		var merchName string
		var count int
		if err := rows.Scan(&merchName, &count); err != nil {
			return nil, fmt.Errorf("%s: error scanning inventory item: %w", op, err)
		}
		inventory = append(inventory, models.InventoryItem{
			Type:     merchName,
			Quantity: count,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: error iterating inventory items: %w", op, err)
	}

	return inventory, nil
}

func (r *Repository) PurchaseMerch(username, merchName string, quantity int) error {
	const op = "repository.PurchaseMerch"
	ctx := context.Background()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s: could not begin transaction: %w", op, err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	var employeeID, balance int
	err = tx.QueryRowContext(ctx, `
		SELECT id, balance 
		FROM employees 
		WHERE username = $1
	`, username).Scan(&employeeID, &balance)
	if err != nil {
		return fmt.Errorf("%s: could not fetch employee data: %w", op, err)
	}

	var price, merchID int
	err = tx.QueryRowContext(ctx, `
		SELECT id, price
		FROM merch 
		WHERE merch_name = $1
	`, merchName).Scan(&merchID, &price)
	if err != nil {
		return fmt.Errorf("%s: could not fetch merch price and id: %w", op, err)
	}

	totalCost := price * quantity
	if balance < totalCost {
		return fmt.Errorf("%s: insufficient balance", op)
	}

	res, err := tx.ExecContext(ctx, `
		UPDATE employees 
		SET balance = balance - $1 
		WHERE id = $2
	`, totalCost, employeeID)
	if err != nil {
		return fmt.Errorf("%s: could not update employee balance: %w", op, err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: could not get affected rows: %w", op, err)
	}
	if rowsAffected != 1 {
		return fmt.Errorf("%s: no employee row updated", op)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO purchases (employee_id, merch_id, count)
		VALUES ($1, $2, $3)
		ON CONFLICT (employee_id, merch_id)
		DO UPDATE SET count = purchases.count + EXCLUDED.count
	`, employeeID, merchID, quantity)
	if err != nil {
		return fmt.Errorf("%s: could not update purchases: %w", op, err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("%s: could not commit transaction: %w", op, err)
	}

	return nil
}

func (r *Repository) SendCoins(senderUsername, receiverUsername string, amount int) error {
	const op = "repository.SendCoins"
	ctx := context.Background()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s: could not begin transaction: %w", op, err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	var senderID, senderBalance int
	err = tx.QueryRowContext(ctx, `
		SELECT id, balance 
		FROM employees 
		WHERE username = $1
	`, senderUsername).Scan(&senderID, &senderBalance)
	if err != nil {
		return fmt.Errorf("%s: could not fetch sender data: %w", op, err)
	}

	if senderBalance < amount {
		return fmt.Errorf("%s: insufficient balance", op)
	}

	var receiverID int
	err = tx.QueryRowContext(ctx, `
		SELECT id 
		FROM employees 
		WHERE username = $1
	`, receiverUsername).Scan(&receiverID)
	if err != nil {
		return fmt.Errorf("%s: could not fetch receiver data: %w", op, err)
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE employees 
		SET balance = balance - $1 
		WHERE id = $2
	`, amount, senderID)
	if err != nil {
		return fmt.Errorf("%s: could not update sender balance: %w", op, err)
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE employees 
		SET balance = balance + $1 
		WHERE id = $2
	`, amount, receiverID)
	if err != nil {
		return fmt.Errorf("%s: could not update receiver balance: %w", op, err)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO transactions (sender_id, receiver_id, amount) 
		VALUES ($1, $2, $3)
	`, senderID, receiverID, amount)
	if err != nil {
		return fmt.Errorf("%s: could not insert transaction record: %w", op, err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("%s: could not commit transaction: %w", op, err)
	}

	return nil
}
