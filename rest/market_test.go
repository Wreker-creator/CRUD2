package rest

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// StubFoodStore is an in-memory implementation of FoodStore used in tests.
type StubFoodStore struct {
	FoodItems []FoodItem
	role      string
}

type StubUserStore struct {
}

func (s *StubUserStore) CreateUser(email, passwordHash string, role string) error
func (s *StubUserStore) GetUserByEmail(email string) (User, error)

func (s *StubUserStore) SaveRefreshToken(userId int, token string, expiresAt time.Time) error
func (s *StubUserStore) GetRefreshToken(token string) (RefreshToken, error)
func (s *StubUserStore) DeleteRefreshToken(token string) error
func (s *StubUserStore) GetUserById(id int) (User, error)

func (s *StubFoodStore) ListAllFoodItems() ([]FoodItem, error) {
	return s.FoodItems, nil
}

func (s *StubFoodStore) ListFoodItem(name string) (FoodItem, error) {
	for _, item := range s.FoodItems {
		if item.Name == name {
			return item, nil
		}
	}
	// Return sql.ErrNoRows so the handler's errors.Is check works correctly.
	return FoodItem{}, sql.ErrNoRows
}

func (s *StubFoodStore) UpdateFoodItem(name string, item FoodItem) (bool, error) {
	for key := range s.FoodItems {
		if s.FoodItems[key].Name == name {
			s.FoodItems[key] = item
			return true, nil // bug fix: was always returning false, error
		}
	}
	// Not found is not a server error — return false, nil so handler returns 404.
	return false, nil
}

func (s *StubFoodStore) AddFoodItem(item FoodItem) error {
	// Stub just appends — validation belongs in the real store, not here.
	s.FoodItems = append(s.FoodItems, item)
	return nil
}

func (s *StubFoodStore) DeleteFoodItem(name string) (bool, error) {
	for key := range s.FoodItems {
		if s.FoodItems[key].Name == name {
			s.FoodItems = append(s.FoodItems[:key], s.FoodItems[key+1:]...)
			return true, nil
		}
	}
	// Not found is not a server error — return false, nil so handler returns 404.
	return false, nil
}

// --- helpers ---

func newRequest(method, url, body string) *http.Request {
	req := httptest.NewRequest(method, url, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func assertStatus(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("status: got %d, want %d", got, want)
	}
}

// --- GET tests ---

func TestGetRequest(t *testing.T) {

	foodSore := &StubFoodStore{
		FoodItems: []FoodItem{
			{Name: "Apple", Price: 1.5, Calories: 95, Sugar: 19},
			{Name: "Banana", Price: 0.5, Calories: 105, Sugar: 14},
		},
	}

	userStore := &StubUserStore{}

	market := NewMarket(foodSore, userStore)

	t.Run("GET /food returns all items with 200", func(t *testing.T) {
		req := newRequest(http.MethodGet, "/food", "")
		res := httptest.NewRecorder()

		market.ServeHTTP(res, req)

		assertStatus(t, res.Code, http.StatusOK)

		var got []FoodItem
		json.NewDecoder(res.Body).Decode(&got)

		if len(got) != len(foodSore.FoodItems) {
			t.Errorf("item count: got %d, want %d", len(got), len(foodSore.FoodItems))
		}
	})

	t.Run("GET /food/{name} returns correct item with 200", func(t *testing.T) {
		req := newRequest(http.MethodGet, "/food/Apple", "")
		res := httptest.NewRecorder()

		market.ServeHTTP(res, req)

		assertStatus(t, res.Code, http.StatusOK)

		var got FoodItem
		json.NewDecoder(res.Body).Decode(&got)

		if got.Name != "Apple" {
			t.Errorf("name: got %q, want %q", got.Name, "Apple")
		}
	})

	t.Run("GET /food/{name} returns 404 for missing item", func(t *testing.T) {
		req := newRequest(http.MethodGet, "/food/Mango", "")
		res := httptest.NewRecorder()

		market.ServeHTTP(res, req)

		assertStatus(t, res.Code, http.StatusNotFound)
	})
}

// --- PUT tests ---

func TestUpdateRequest(t *testing.T) {

	foodstore := &StubFoodStore{
		FoodItems: []FoodItem{
			{Name: "Apple", Price: 1.5, Calories: 95, Sugar: 19},
		},
	}

	userStore := &StubUserStore{}

	market := NewMarket(foodstore, userStore)

	t.Run("PUT /food/{name} updates existing item and returns 200", func(t *testing.T) {
		body := `{"name":"Apple","price":2.0,"calories":100,"sugar":20}`
		req := newRequest(http.MethodPut, "/food/Apple", body)
		res := httptest.NewRecorder()

		market.ServeHTTP(res, req)

		assertStatus(t, res.Code, http.StatusOK)
	})

	t.Run("PUT /food/{name} returns 404 for missing item", func(t *testing.T) {
		body := `{"name":"Mango","price":2.0,"calories":100,"sugar":20}`
		req := newRequest(http.MethodPut, "/food/Mango", body)
		res := httptest.NewRecorder()

		market.ServeHTTP(res, req)

		assertStatus(t, res.Code, http.StatusNotFound)
	})

	t.Run("PUT /food/{name} returns 400 for malformed body", func(t *testing.T) {
		req := newRequest(http.MethodPut, "/food/Apple", `{invalid}`)
		res := httptest.NewRecorder()

		market.ServeHTTP(res, req)

		assertStatus(t, res.Code, http.StatusBadRequest)
	})
}

// --- DELETE tests ---

func TestDeleteRequest(t *testing.T) {

	foodstore := &StubFoodStore{
		FoodItems: []FoodItem{
			{Name: "Apple", Price: 1.5, Calories: 95, Sugar: 19},
		},
	}

	userStore := &StubUserStore{}

	market := NewMarket(foodstore, userStore)

	t.Run("DELETE /food/{name} removes item and returns 202", func(t *testing.T) {
		req := newRequest(http.MethodDelete, "/food/Apple", "")
		res := httptest.NewRecorder()

		market.ServeHTTP(res, req)

		assertStatus(t, res.Code, http.StatusAccepted)
	})

	t.Run("DELETE /food/{name} returns 404 for missing item", func(t *testing.T) {
		req := newRequest(http.MethodDelete, "/food/Mango", "")
		res := httptest.NewRecorder()

		market.ServeHTTP(res, req)

		assertStatus(t, res.Code, http.StatusNotFound)
	})
}

// --- POST tests ---

func TestAddRequest(t *testing.T) {

	foodStore := &StubFoodStore{}
	userStore := &StubUserStore{}
	market := NewMarket(foodStore, userStore)

	t.Run("POST /food adds item and returns 201", func(t *testing.T) {
		body := `{"name":"Apple","price":1.5,"calories":95,"sugar":19}`
		req := newRequest(http.MethodPost, "/food", body)
		res := httptest.NewRecorder()

		market.ServeHTTP(res, req)

		assertStatus(t, res.Code, http.StatusCreated)

		if len(foodStore.FoodItems) != 1 {
			t.Errorf("expected 1 item in store, got %d", len(foodStore.FoodItems))
		}
	})

	t.Run("POST /food returns 400 for malformed body", func(t *testing.T) {
		req := newRequest(http.MethodPost, "/food", `{invalid}`)
		res := httptest.NewRecorder()

		market.ServeHTTP(res, req)

		assertStatus(t, res.Code, http.StatusBadRequest)
	})
}
