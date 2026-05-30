package rest

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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

	// added a global middleware
	r.Use(middleware.Logger)    // logs the start and end of every request
	r.Use(middleware.Recoverer) // handles panics by returning a 500

	// auth routes
	r.Post("/auth/register", authHandler.handleRegister)
	r.Post("/auth/login", authHandler.handleLogin)
	r.Post("/auth/refresh", authHandler.handleRefresh)
	r.Post("/auth/logout", authHandler.handleLogout)

	// food routes are protected, every request must carry a valid JWT
	// chi.group creates a subrouter, where we apply middleware ONLY to these routes

	r.Group(func(r chi.Router) {
		r.Use(JWTMiddleWare) // runs before every handler inside this group

		// read routers - any authenticated user can access
		r.Get("/food", marketHandler.returnAllItems)
		r.Get("/food/{name}", marketHandler.handleGetRequest)

		// write routers, only the users with admin access can use it.
		// NOTE - r.with applies the role middleware to single route instead of all of them
		r.With(RoleMiddleware("admin")).Post("/food", marketHandler.handlePostRequest)
		r.With(RoleMiddleware("admin")).Put("/food/{name}", marketHandler.handlePutRequest)
		r.With(RoleMiddleware("admin")).Delete("/food/{name}", marketHandler.handleDeleteRequest)
	})

	market.Handler = r

	return market
}
