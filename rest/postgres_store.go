package rest

import (
	"database/sql"
	"errors"
	"fmt"

	// blank import: pq registers itself as the "postgres" driver with database/sql.
	// we never call anything from this package directly.
	_ "github.com/lib/pq"
)

// PostgresFoodStore holds a connection pool to Postgres.
// *sql.DB is safe for concurrent use and manages the pool internally.
type PostgresFoodStore struct {
	db *sql.DB
}

// NewPostgresFoodStore opens a connection pool and verifies the DB is reachable.
// sql.Open does not actually connect — Ping does.
func NewPostgresFoodStore(dsn string) (*PostgresFoodStore, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("Failed to start the database, '%w'", err)
	}

	// Ping establishes a real connection to verify the DSN is correct and the DB is up
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("Failed to ping the database, '%w'", err)
	}

	return &PostgresFoodStore{db: db}, nil
}

// ListAllFoodItems returns all rows from the market table ordered alphabetically.
func (p *PostgresFoodStore) ListAllFoodItems() ([]FoodItem, error) {

	rows, err := p.db.Query(
		`SELECT name, price, calories, sugar FROM market ORDER BY name ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("ListAllFoodItems query failed, '%w'", err)
	}

	// Close releases the connection back to the pool once we're done iterating.
	defer rows.Close()

	// make([]FoodItem, 0) ensures we return [] not null in JSON if the table is empty.
	foodItems := make([]FoodItem, 0)

	// rows.Next advances the cursor one row at a time. Returns false when done or on error.
	for rows.Next() {
		var item FoodItem
		// Scan copies the current row's columns into the struct fields in order.
		if err := rows.Scan(&item.Name, &item.Price, &item.Calories, &item.Sugar); err != nil {
			return nil, fmt.Errorf("ListAllFoodItems scan failed, '%w'", err)
		}

		foodItems = append(foodItems, item)
	}

	// rows.Err captures any error that caused rows.Next to stop early.
	// Without this check a network failure mid-iteration would go silently unnoticed.
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ListAllFoodItems rows error, '%w'", err)
	}

	return foodItems, nil
}

// ListFoodItem returns a single food item matching the given name.
func (p *PostgresFoodStore) ListFoodItem(name string) (FoodItem, error) {
	var item FoodItem

	// QueryRow is used when we expect exactly one row.
	// $1 is a positional placeholder — pq substitutes name safely, preventing SQL injection.
	err := p.db.QueryRow(
		`SELECT name, price, calories, sugar FROM market WHERE name = $1`, name,
	).Scan(&item.Name, &item.Price, &item.Calories, &item.Sugar)

	// sql.ErrNoRows is returned by Scan when QueryRow found no matching row.
	// We wrap it with %w so the caller can still unwrap and check it with errors.Is.
	if errors.Is(err, sql.ErrNoRows) {
		return FoodItem{}, fmt.Errorf("No such food item found '%w'", err)
	}
	if err != nil {
		return FoodItem{}, fmt.Errorf("ListFoodItem query failed, '%w'", err)
	}

	return item, nil
}

// UpdateFoodItem replaces the price, calories, and sugar of the named item.
// Returns true if a row was updated, false if the name didn't exist.
func (p *PostgresFoodStore) UpdateFoodItem(name string, item FoodItem) (bool, error) {
	// Exec is used for statements that don't return rows (INSERT, UPDATE, DELETE).
	res, err := p.db.Exec(
		`UPDATE market SET price = $1, calories = $2, sugar = $3 WHERE name = $4`,
		item.Price, item.Calories, item.Sugar, name,
	)
	if err != nil {
		return false, fmt.Errorf("Failed to update the food item's values '%w'", err)
	}

	// RowsAffected tells us how many rows the UPDATE touched.
	// 0 means the WHERE clause matched nothing — the item doesn't exist.
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("Check failed for update food item value '%w'", err)
	}

	return rowsAffected > 0, nil
}

// AddFoodItem inserts a new food item into the market table.
// id, created_at, and updated_at are handled automatically by Postgres.
func (p *PostgresFoodStore) AddFoodItem(item FoodItem) error {
	_, err := p.db.Exec(
		`INSERT INTO market (name, price, calories, sugar) VALUES ($1, $2, $3, $4)`,
		item.Name, item.Price, item.Calories, item.Sugar,
	)
	if err != nil {
		return fmt.Errorf("Failed to add the food item '%w'", err)
	}
	return nil
}

// DeleteFoodItem removes the food item matching the given name.
// Returns true if a row was deleted, false if the name didn't exist.
func (p *PostgresFoodStore) DeleteFoodItem(name string) (bool, error) {

	res, err := p.db.Exec(`DELETE FROM market WHERE name = $1`, name)
	if err != nil {
		return false, fmt.Errorf("Failed to delete the item '%w'", err)
	}

	// same pattern as UpdateFoodItem — 0 rows affected means the name wasn't found
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("Check failed for delete food item '%w'", err)
	}

	return rowsAffected > 0, nil
}
