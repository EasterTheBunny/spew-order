package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/easterthebunny/render"
	"github.com/easterthebunny/spew-order/internal/contexts"
	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/pkg/api"
	"github.com/easterthebunny/spew-order/pkg/domain"
	"github.com/easterthebunny/spew-order/pkg/types"
	"github.com/go-chi/chi"
	uuid "github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
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

		items := []api.BalanceItem{}
		for _, s := range acct.ActiveSymbols() {
			i := api.BalanceItem{}
			if hash, ok := acct.Addresses[s]; ok {
				i.Funding = hash
			}
			if bal, ok := acct.Balances[s]; ok {
				i.Quantity = api.CurrencyValue(bal.StringFixedBank(s.RoundingPlace()))
			}
			i.Symbol = api.SymbolType(s.String())
			items = append(items, i)
		}
		bl := api.BalanceList(items)
		res.Balances = &bl

		render.Render(w, r, HTTPNewOKResponse(&res))
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

// PostTransaction ...
func (h *AccountHandler) PostTransaction(b *domain.BalanceManager) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		acct := contexts.GetAccount(r.Context())

		if acct == nil {
			render.Render(w, r, HTTPInternalServerError(errors.New("incorrect route structure")))
			return
		}

		var in api.TransactionRequest
		err := json.NewDecoder(r.Body).Decode(&in)
		if err != nil {
			render.Render(w, r, HTTPBadRequest(err))
			return
		}

		amt, err := decimal.NewFromString(string(in.Quantity))
		if err != nil {
			render.Render(w, r, HTTPBadRequest(err))
			return
		}

		if amt.LessThanOrEqual(decimal.NewFromInt(0)) {
			render.Render(w, r, HTTPBadRequest(errors.New("quantity must be greater than 0")))
			return
		}

		var smb types.Symbol
		err = json.Unmarshal([]byte(fmt.Sprintf(`"%s"`, string(in.Symbol))), &smb)
		if err != nil {
			render.Render(w, r, HTTPBadRequest(err))
			return
		}

		if len(in.Address) == 0 || !smb.ValidateAddress(in.Address) {
			render.Render(w, r, HTTPBadRequest(errors.New("invalid address")))
			return
		}

		tr, err := b.WithdrawFunds(acct, smb, amt, in.Address)
		if err != nil {
			if errors.Is(domain.ErrInsufficientBalanceForHold, err) {
				render.Render(w, r, HTTPConflict(err))
				return
			}
			render.Render(w, r, HTTPInternalServerError(err))
			return
		}

		if tr == nil {
			render.Render(w, r, HTTPInternalServerError(errors.New("unexpected state")))
			return
		}

		o := api.Transaction{
			Type:            api.TransactionTypeTRANSFER,
			Symbol:          api.SymbolType(tr.Symbol),
			Quantity:        api.CurrencyValue(tr.Quantity),
			Fee:             "",
			Orderid:         "",
			Timestamp:       time.Time(tr.Timestamp).Format(time.RFC3339),
			TransactionHash: tr.TransactionHash,
		}
		render.Render(w, r, HTTPNewOKResponse(&o))
	}
}

func (h *AccountHandler) GetAccountTransactions() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		acct := contexts.GetAccount(r.Context())
		tr := h.repo.Transactions(&persist.Account{ID: acct.ID.String()})

		list, err := tr.GetTransactions()
		if err != nil {
			render.Render(w, r, HTTPInternalServerError(err))
			return
		}

		var out []render.Renderer
		for _, trans := range list {
			t := api.Transaction{
				Type:            api.StringTransactionType(trans.Type),
				Symbol:          api.SymbolType(trans.Symbol),
				Quantity:        api.CurrencyValue(trans.Quantity),
				Fee:             api.CurrencyValue(trans.Fee),
				Orderid:         trans.OrderID,
				Timestamp:       time.Time(trans.Timestamp).Format(time.RFC3339),
				TransactionHash: trans.TransactionHash,
			}
			out = append(out, &t)
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

func (h *AccountHandler) AccountCtx(bm *domain.BalanceManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var ctxAccount *domain.Account
			var err error

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

				// look for the account in storage and create the account if it doesn't exist
				ctxAccount, err = bm.GetAccount(accountID)
				if err != nil {
					render.Render(w, r, HTTPInternalServerError(err))
					return
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
