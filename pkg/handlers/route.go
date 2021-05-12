package handlers

import (
	"net/http"

	"github.com/easterthebunny/spew-order/internal/middleware"
	"github.com/easterthebunny/spew-order/pkg/types"
	"github.com/go-chi/chi"
)

// Router ...
func Router(s types.AuthorizationStore, p types.AuthenticationProvider) http.Handler {
	r := chi.NewRouter()

	// set CORS headers early and short circuit the response loop
	r.Use(middleware.SetCORSHeaders)

	// set up routes that require authorization
	r.Route("/", AuthorizedRoutes(s, p))

	return r
}

// AuthorizedRoutes ...
func AuthorizedRoutes(s types.AuthorizationStore, p types.AuthenticationProvider) func(r chi.Router) {
	return func(r chi.Router) {

		// authentication middleware for accessing the bearer token
		r.Use(p.Verifier())
		r.Use(p.Authenticator())

		// put the authorization in the context
		r.Use(middleware.AuthorizationCtx(s, p))

	}
}

func OrderSubRoutes(handlers *RESTHandler) func(r chi.Router) {
	return func(r chi.Router) {
		r.Post("/", handlers.PostOrder())
	}
}
