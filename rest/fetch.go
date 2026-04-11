package rest

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// FoodItem represents a single food product in the market.
type FoodItem struct {
	Name     string  `json:"name"`
	Price    float32 `json:"price"`
	Calories int     `json:"calories"`
	Sugar    float32 `json:"sugar"`
}

// FoodStore defines the contract for any storage backend (in-memory, postgres, etc.)
type FoodStore interface {
	ListAllFoodItems() ([]FoodItem, error)
	ListFoodItem(name string) (FoodItem, error)
	UpdateFoodItem(name string, item FoodItem) (bool, error)
	AddFoodItem(item FoodItem) error
	DeleteFoodItem(name string) (bool, error)
}

// Market wires the HTTP router to the store.
type Market struct {
	store FoodStore
	http.Handler
}

// NewMarket creates a Market and registers all routes.
func NewMarket(store FoodStore) *Market {

	market := &Market{store: store}

	r := chi.NewRouter()

	r.Get("/food", market.returnAllItems)
	r.Post("/food", market.handlePostRequest)
	r.Get("/food/{name}", market.handleGetRequest)
	r.Put("/food/{name}", market.handlePutRequest) // PUT replaces the full resource
	r.Delete("/food/{name}", market.handleDeleteRequest)

	market.Handler = r

	return market
}

// handleGetRequest retrieves a single food item by ID.
func (m *Market) handleGetRequest(w http.ResponseWriter, r *http.Request) {

	name := chi.URLParam(r, "name")

	foodItem, err := m.store.ListFoodItem(name)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Food item not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Failed to retrieve the data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(foodItem)
}

// handlePutRequest fully replaces an existing food item by ID.
func (m *Market) handlePutRequest(w http.ResponseWriter, r *http.Request) {

	idStr := chi.URLParam(r, "name")

	var item FoodItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	updated, err := m.store.UpdateFoodItem(idStr, item)
	if err != nil {
		http.Error(w, "Failed to update food item", http.StatusInternalServerError)
		return
	}
	if !updated {
		// Item with given ID does not exist
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// handleDeleteRequest removes a food item by ID.
func (m *Market) handleDeleteRequest(w http.ResponseWriter, r *http.Request) {

	idStr := chi.URLParam(r, "id")
	deleted, err := m.store.DeleteFoodItem(idStr)
	if err != nil {
		http.Error(w, "Failed to delete the food item", http.StatusInternalServerError)
		return
	}
	if !deleted {
		// Item with given ID does not exist
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// 202 Accepted: request was valid and the item has been removed
	w.WriteHeader(http.StatusAccepted)
}

// handlePostRequest adds a new food item.
// Returns 201 with no body — the client supplied all the data,
// so echoing it back provides no extra value.
func (m *Market) handlePostRequest(w http.ResponseWriter, r *http.Request) {

	var foodItem FoodItem

	if err := json.NewDecoder(r.Body).Decode(&foodItem); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := m.store.AddFoodItem(foodItem); err != nil {
		http.Error(w, "Failed to add the food item", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// returnAllItems retrieves and returns every food item in the store.
func (m *Market) returnAllItems(w http.ResponseWriter, r *http.Request) {

	food, err := m.store.ListAllFoodItems()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return // bug fix: was missing return, would encode even on error
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(food)
}
