package middleware

import (
	"net/http"

	"github.com/easterthebunny/spew-order/internal/contexts"
	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/pkg/domain"
)

type authKey string

func (a authKey) String() string {
	return string(a)
}

// AuthenticationProvider ...
type AuthenticationProvider interface {
	Verifier() func(http.Handler) http.Handler
	UpdateAuthz(*persist.Authorization)
	Subject() string
}

// AuthorizationCtx ...
func AuthorizationCtx(as persist.AuthorizationRepository, p AuthenticationProvider) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		hfn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			a, err := as.GetAuthorization(ctx, authKey(p.Subject()))

			if err == persist.ErrObjectNotExist {
				acc := domain.NewAccount()
				a = persist.NewAuthorization(persist.Account{ID: acc.ID.String()})

				p.UpdateAuthz(a)
				err = as.SetAuthorization(ctx, a)
			}

			if err != nil {
				// TODO: handle errors without panic; maybe write an immediate response
				panic(err)
			}

			ctx = contexts.AttachAuthorization(ctx, *a)
			next.ServeHTTP(w, r.WithContext(ctx))
		}

		return http.HandlerFunc(hfn)
	}
}
