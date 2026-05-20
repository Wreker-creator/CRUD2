package rest

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Market wires the HTTP router to the store.
// Embedding http.Handler means *Market itself satisfies http.Handler,
// so it can be passed directly to http.ListenAndServe.
type Market struct {
	userStore UserStore
	foodStore FoodStore
	http.Handler
}

// NewMarket creates a Market and registers all routes.
func NewMarket(foodstore FoodStore, userstore UserStore) *Market {

	market := &Market{foodStore: foodstore, userStore: userstore}

	// auth ahndler get its own handler struct, same pattern as market
	authHandler := authHandler{store: userstore}
	marketHandler := marketHandler{store: foodstore}

	r := chi.NewRouter()

	// auth routes
	r.Post("/auth/register", authHandler.handleRegister)
	r.Post("/auth/login", authHandler.handleLogin)

	// food routes are protected, every request must carry a valid JWT
	// chi.group creates a subrouter, where we apply middleware ONLY to these routes

	r.Group(func(r chi.Router) {
		r.Use(JWTMiddleWare) // runs before every handler inside this group
		r.Get("/food", marketHandler.returnAllItems)
		r.Post("/food", marketHandler.handlePostRequest)
		r.Get("/food/{name}", marketHandler.handleGetRequest)
		r.Put("/food/{name}", marketHandler.handlePutRequest)
		r.Delete("/food/{name}", marketHandler.handleDeleteRequest)
	})

	market.Handler = r

	return market
}
