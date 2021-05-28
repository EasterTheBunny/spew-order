package handlers

import (
	"fmt"
	"net/http"

	"github.com/easterthebunny/spew-order/internal/middleware"
	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/pkg/api"
	"github.com/easterthebunny/spew-order/pkg/domain"
	"github.com/go-chi/chi"
)

type Router struct {
	AuthStore persist.AuthorizationRepository
	Balance   *domain.BalanceManager
	AuthProv  middleware.AuthenticationProvider
	Orders    *OrderHandler
	Accounts  *AccountHandler
}

// Router ...
func (d *Router) Routes() http.Handler {
	r := chi.NewRouter()

	// set CORS headers early and short circuit the response loop
	r.Use(middleware.SetCORSHeaders)

	// set up routes that require authorization
	r.Route("/", d.AuthorizedRoutes())

	return r
}

// AuthorizedRoutes ...
func (d *Router) AuthorizedRoutes() func(r chi.Router) {
	return func(r chi.Router) {

		// authentication middleware for accessing and verifying the bearer token
		r.Use(d.AuthProv.Verifier())

		// put the authorization in the context
		r.Use(middleware.AuthorizationCtx(d.AuthStore, d.AuthProv))

		r.Route("/account", d.AccountRoutes())
	}
}

func (d *Router) AccountRoutes() func(r chi.Router) {
	return func(r chi.Router) {
		r.Get("/", d.Accounts.GetAccounts())
		r.Route(fmt.Sprintf("/{%s}", api.AccountPathParamName), d.AccountSubRoutes())
	}
}

func (d *Router) AccountSubRoutes() func(r chi.Router) {
	return func(r chi.Router) {
		r.Use(d.Accounts.AccountCtx(d.Balance))
		r.Get("/", d.Accounts.GetAccount())
		r.Route("/order", d.OrderRoutes())
	}
}

func (d *Router) OrderRoutes() func(r chi.Router) {
	return func(r chi.Router) {
		r.Post("/", d.Orders.PostOrder())
		r.Get("/", d.Accounts.GetAccountOrders())
		r.Route(fmt.Sprintf("/{%s}", api.OrderPathParamName), d.OrderSubRoutes())
	}
}

func (d *Router) OrderSubRoutes() func(r chi.Router) {
	return func(r chi.Router) {
		r.Use(d.Accounts.OrderCtx())
		r.Get("/", d.Accounts.GetAccountOrder())
	}
}

type WebhookRouter struct {
	Funding *FundingHandler
}

func (wr *WebhookRouter) Routes() http.Handler {

	r := chi.NewRouter()

	// set CORS headers early and short circuit the response loop
	r.Use(middleware.SetCORSHeaders)
	r.Use(wr.Funding.Source.Callback())

	r.Post("/funding", wr.Funding.PostFunding())

	return r
}
