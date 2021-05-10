package middleware

import (
	"net/http"

	"github.com/easterthebunny/spew-order/internal/contexts"
	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/pkg/types"
)

// AuthorizationCtx ...
func AuthorizationCtx(as types.AuthorizationStore, p types.AuthenticationProvider) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		hfn := func(w http.ResponseWriter, r *http.Request) {
			auth, err := as.GetAuthorization(p.Subject())

			if err == persist.ErrAuthzNotFound {
				acc := types.NewAccount()

				auth = &types.Authorization{
					Accounts: []string{acc.ID.String()}}

				p.UpdateAuthz(auth)
				err = as.SetAuthorization(auth)
			}

			if err != nil {
				// TODO: handle errors without panic; maybe write an immediate response
				panic(err)
			}

			ctx := contexts.AttachAuthorization(r.Context(), *auth)
			next.ServeHTTP(w, r.WithContext(ctx))
		}

		return http.HandlerFunc(hfn)
	}
}
