package rest

// FoodItem represents a single food product in the market.
// json tags control how fields are named when encoding/decoding JSON.
type FoodItem struct {
	Name     string  `json:"name"`
	Price    float32 `json:"price"`
	Calories int     `json:"calories"`
	Sugar    float32 `json:"sugar"`
}

// FoodStore defines the contract for any storage backend.
// Both PostgresFoodStore and StubFoodStore implement this interface,
// which is what allows handler tests to work without a real database.
type FoodStore interface {
	ListAllFoodItems() ([]FoodItem, error)
	ListFoodItem(name string) (FoodItem, error)
	UpdateFoodItem(name string, item FoodItem) (bool, error)
	AddFoodItem(item FoodItem) error
	DeleteFoodItem(name string) (bool, error)
}
