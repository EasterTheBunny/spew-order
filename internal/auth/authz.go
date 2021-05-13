package auth

import (
	"net/http"

	"github.com/easterthebunny/spew-order/pkg/types"
	uuid "github.com/satori/go.uuid"
)

// NewAuthorization returns a new auth with values set to defaults and a new
// id generated.
func NewAuthorization(accts ...types.Account) Authorization {
	var ids []string
	for _, a := range accts {
		ids = append(ids, a.ID.String())
	}

	return Authorization{
		ID:       uuid.NewV4().String(),
		Accounts: ids}
}

// Authorization ...
type Authorization struct {
	ID       string   `json:"id"`
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Name     string   `json:"name"`
	Avatar   string   `json:"avatar"`
	Accounts []string `json:"accounts"`
}

// AuthorizationStore ...
type AuthorizationStore interface {
	GetAuthorization(string) (*Authorization, error)
	SetAuthorization(*Authorization) error
	DeleteAuthorization(*Authorization) error
}

// AuthenticationProvider ...
type AuthenticationProvider interface {
	Verifier() func(http.Handler) http.Handler
	UpdateAuthz(*Authorization)
	Subject() string
}
