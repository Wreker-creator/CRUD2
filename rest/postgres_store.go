package rest

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/lib/pq"
)

type PostgresFoodStore struct {
	db *sql.DB
}

func NewPostgresFoodStore(dsn string) (*PostgresFoodStore, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("Failed to start the database, '%w'", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("Failed to ping the database, '%w'", err)
	}

	return &PostgresFoodStore{db: db}, nil
}

func (p *PostgresFoodStore) ListAllFoodItems() ([]FoodItem, error) {

	rows, err := p.db.Query(
		`SELECT name, price, calories, sugar FROM market ORDER BY name ASC`,
	)

	if err != nil {
		return nil, fmt.Errorf("ListAllFoodItems query failed, '%w'", err)
	}

	defer rows.Close()

	foodItems := make([]FoodItem, 0)

	for rows.Next() {
		var item FoodItem
		if err := rows.Scan(&item.Name, &item.Price, &item.Calories, &item.Sugar); err != nil {
			return nil, fmt.Errorf("ListAllFoodItems scan failed, '%w'", err)
		}

		foodItems = append(foodItems, item)
	}

	// checking if any error occurrerd during the iteration.
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ListAllFoodItems rows error, '%w'", err)
	}

	return foodItems, nil

}

func (p *PostgresFoodStore) ListFoodItem(name string) (FoodItem, error) {
	var item FoodItem

	err := p.db.QueryRow(`SELECT name, price, calories, sugar FROM market WHERE name = $1`, name).Scan(&item.Name, &item.Price, &item.Calories, &item.Sugar)

	if errors.Is(err, sql.ErrNoRows) {
		return FoodItem{}, fmt.Errorf("No such food item found '%w'", err)
	}

	if err != nil {
		return FoodItem{}, fmt.Errorf("ListFoodItem query failed, '%w'", err)
	}

	return item, nil
}

func (p *PostgresFoodStore) UpdateFoodItem(name string, item FoodItem) (bool, error) {
	res, err := p.db.Exec(`UPDATE market SET price = $1, calories = $2, sugar = $3 WHERE name = $4`, item.Price, item.Calories, item.Sugar, name)
	if err != nil {
		return false, fmt.Errorf("Failed to update the food item's values '%w'", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("Check failed for update food item value '%w'", err)
	}

	return rowsAffected > 0, nil
}

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

func (p *PostgresFoodStore) DeleteFoodItem(name string) (bool, error) {

	res, err := p.db.Exec(`DELETE FROM market WHERE name = $1`, name)
	if err != nil {
		return false, fmt.Errorf("Failed to delete the item '%w'", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("Check failed for delete food item '%w'", err)
	}

	return rowsAffected > 0, nil

}
