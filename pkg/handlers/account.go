package handlers

import (
	"errors"
	"net/http"

	"github.com/easterthebunny/render"
	"github.com/easterthebunny/spew-order/internal/account"
	"github.com/easterthebunny/spew-order/internal/contexts"
	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/pkg/api"
	"github.com/easterthebunny/spew-order/pkg/types"
	"github.com/go-chi/chi"
	uuid "github.com/satori/go.uuid"
)

type AccountHandler struct {
	repo account.AccountRepository
}

func NewAccountHandler(r account.AccountRepository) *AccountHandler {
	return &AccountHandler{repo: r}
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

func (h *AccountHandler) AccountCtx() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var acct *types.Account
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

				// look for the account in storage and create the account if it doesn't exist
				acct, err = h.repo.Find(id)
				if err != nil {
					if err == persist.ErrObjectNotExist {
						acc := types.NewAccount()
						acc.ID = id
						acct = &acc
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

			ctx := contexts.AttachAccount(r.Context(), *acct)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
