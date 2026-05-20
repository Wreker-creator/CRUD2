package rest

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type marketHandler struct {
	store FoodStore
}

func (m *marketHandler) handleGetRequest(w http.ResponseWriter, r *http.Request) {

	name := chi.URLParam(r, "name")

	foodItem, err := m.store.ListFoodItem(name)
	// sql.ErrNoRows is the sentinel error database/sql returns when a query
	// finds no matching row — we map that specifically to a 404.
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

// handlePutRequest fully replaces an existing food item.
// PUT semantics: the client must send all fields, not just the ones changing.
func (m *marketHandler) handlePutRequest(w http.ResponseWriter, r *http.Request) {

	name := chi.URLParam(r, "name")

	var item FoodItem
	// Decode reads the JSON body into item. Returns an error for malformed JSON.
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	updated, err := m.store.UpdateFoodItem(name, item)
	if err != nil {
		http.Error(w, "Failed to update food item", http.StatusInternalServerError)
		return
	}
	if !updated {
		// store returns false when no rows were affected, meaning name didn't exist
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// handleDeleteRequest removes a food item by name.
func (m *marketHandler) handleDeleteRequest(w http.ResponseWriter, r *http.Request) {

	name := chi.URLParam(r, "name")

	deleted, err := m.store.DeleteFoodItem(name)
	if err != nil {
		http.Error(w, "Failed to delete the food item", http.StatusInternalServerError)
		return
	}
	if !deleted {
		// store returns false when no rows were affected, meaning name didn't exist
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// 202 Accepted: the request was valid and the item has been removed
	w.WriteHeader(http.StatusAccepted)
}

// handlePostRequest adds a new food item.
// Returns 201 with no body — the client supplied all the data,
// so echoing it back provides no extra value.
func (m *marketHandler) handlePostRequest(w http.ResponseWriter, r *http.Request) {

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
func (m *marketHandler) returnAllItems(w http.ResponseWriter, r *http.Request) {

	food, err := m.store.ListAllFoodItems()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return // must return here — without this, Encode would still run on a broken response
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(food)
}
