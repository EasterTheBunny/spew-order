package middleware

import (
	"net/http"

	"github.com/easterthebunny/spew-order/internal/auth"
	"github.com/easterthebunny/spew-order/internal/contexts"
	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/pkg/types"
)

type authKey string

func (a authKey) String() string {
	return string(a)
}

// AuthorizationCtx ...
func AuthorizationCtx(as persist.AuthorizationRepository, p auth.AuthenticationProvider) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		hfn := func(w http.ResponseWriter, r *http.Request) {
			a, err := as.GetAuthorization(authKey(p.Subject()))

			if err == persist.ErrAuthzNotFound {
				acc := types.NewAccount()

				a = &persist.Authorization{
					Accounts: []string{acc.ID.String()}}

				p.UpdateAuthz(a)
				err = as.SetAuthorization(a)
			}

			if err != nil {
				// TODO: handle errors without panic; maybe write an immediate response
				panic(err)
			}

			ctx := contexts.AttachAuthorization(r.Context(), *a)
			next.ServeHTTP(w, r.WithContext(ctx))
		}

		return http.HandlerFunc(hfn)
	}
}
