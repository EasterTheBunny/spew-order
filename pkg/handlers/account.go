package handlers

import (
	"errors"
	"net/http"

	"github.com/easterthebunny/render"
	"github.com/easterthebunny/spew-order/internal/contexts"
	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/pkg/api"
	"github.com/easterthebunny/spew-order/pkg/domain"
	"github.com/easterthebunny/spew-order/pkg/types"
	"github.com/go-chi/chi"
	uuid "github.com/satori/go.uuid"
)

type AccountHandler struct {
	repo persist.AccountRepository
}

func NewAccountHandler(r persist.AccountRepository) *AccountHandler {
	return &AccountHandler{repo: r}
}

func (h *AccountHandler) GetAccounts() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		authz := contexts.GetAuthorization(r.Context())

		var res []render.Renderer
		for _, acct := range authz.Accounts {
			res = append(res, &api.Account{
				Id: acct,
			})
		}

		render.Render(w, r, HTTPNewOKListResponse(res))
	}
}

func (h *AccountHandler) GetAccount() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		acct := contexts.GetAccount(r.Context())

		res := api.Account{
			Id: acct.ID.String(),
		}

		render.Render(w, r, HTTPNewOKResponse(&res))
	}
}

func (h *AccountHandler) GetAccountBalances(bm *domain.BalanceManager) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		acct := contexts.GetAccount(r.Context())

		var err error
		var list []render.Renderer

		// set btc balance
		btcBal := &api.BalanceItem{
			Symbol: api.SymbolType(types.SymbolBitcoin.String())}
		amt, err := bm.GetAvailableBalance(acct, types.SymbolBitcoin)
		if err != nil {
			render.Render(w, r, HTTPInternalServerError(err))
			return
		}
		btcBal.Quantity = api.CurrencyValue(amt.StringFixedBank(types.SymbolBitcoin.RoundingPlace()))
		list = append(list, btcBal)

		// set eth balance
		ethBal := &api.BalanceItem{
			Symbol: api.SymbolType(types.SymbolEthereum.String())}
		amt, err = bm.GetAvailableBalance(acct, types.SymbolEthereum)
		if err != nil {
			render.Render(w, r, HTTPInternalServerError(err))
			return
		}
		ethBal.Quantity = api.CurrencyValue(amt.StringFixedBank(types.SymbolEthereum.RoundingPlace()))
		list = append(list, ethBal)

		render.Render(w, r, HTTPNewOKListResponse(list))
	}
}

func (h *AccountHandler) GetAccountOrder() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ord := contexts.GetOrder(r.Context())

		res := api.BookOrder{
			Guid:   ord.Base.ID.String(),
			Order:  api.BuildOrderRequest(ord.Base.OrderRequest),
			Status: api.StringOrderStatus(ord.Status),
		}

		render.Render(w, r, HTTPNewOKResponse(&res))
	}
}

func (h *AccountHandler) GetAccountOrders() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		acct := contexts.GetAccount(r.Context())
		or := h.repo.Orders(&persist.Account{ID: acct.ID.String()})

		list, err := or.GetOrdersByStatus(persist.StatusOpen, persist.StatusPartial, persist.StatusFilled, persist.StatusCanceled)
		if err != nil {
			render.Render(w, r, HTTPInternalServerError(err))
			return
		}

		var out []render.Renderer
		for _, ord := range list {
			o := api.BookOrder{
				Guid:   ord.Base.ID.String(),
				Order:  api.BuildOrderRequest(ord.Base.OrderRequest),
				Status: api.StringOrderStatus(ord.Status),
			}
			out = append(out, &o)
		}

		render.Render(w, r, HTTPNewOKListResponse(out))
	}
}

func (h *AccountHandler) OrderCtx() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var ctxOrder *persist.Order
			var err error
			var id uuid.UUID

			acct := contexts.GetAccount(r.Context())

			if orderID := chi.URLParam(r, api.OrderPathParamName); orderID != "" {

				id, err = uuid.FromString(orderID)
				if err != nil {
					render.Render(w, r, HTTPBadRequest(err))
					return
				}

				or := h.repo.Orders(&persist.Account{ID: acct.ID.String()})
				ctxOrder, err = or.GetOrder(id)
				if err != nil {
					render.Render(w, r, HTTPInternalServerError(err))
					return
				}
			} else {
				render.Render(w, r, HTTPBadRequest(errors.New("order id not found")))
				return
			}

			ctx := contexts.AttachOrder(r.Context(), *ctxOrder)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (h *AccountHandler) AccountCtx() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var ctxAccount *domain.Account
			var err error
			var id uuid.UUID

			authz := contexts.GetAuthorization(r.Context())

			var availableID string
			for _, j := range authz.Accounts {
				availableID = j
			}

			if availableID == "" {
				render.Render(w, r, HTTPUnauthorized(errors.New("no accounts authorized")))
				return
			}

			if accountID := chi.URLParam(r, api.AccountPathParamName); accountID != "" {
				if accountID != availableID {
					render.Render(w, r, HTTPUnauthorized(errors.New("invalid authorization to access this account")))
					return
				}

				id, err = uuid.FromString(accountID)
				if err != nil {
					render.Render(w, r, HTTPBadRequest(err))
					return
				}

				ctxAccount = domain.NewAccount()
				ctxAccount.ID = id

				// look for the account in storage and create the account if it doesn't exist
				_, err = h.repo.Find(id)
				if err != nil {
					if errors.Is(err, persist.ErrObjectNotExist) {

						acct := &persist.Account{ID: ctxAccount.ID.String()}
						err = h.repo.Save(acct)
						if err != nil {
							render.Render(w, r, HTTPInternalServerError(err))
							return
						}
					} else {
						render.Render(w, r, HTTPInternalServerError(err))
						return
					}
				}
			} else {
				render.Render(w, r, HTTPBadRequest(errors.New("account id not found")))
				return
			}

			ctx := contexts.AttachAccount(r.Context(), *ctxAccount)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
