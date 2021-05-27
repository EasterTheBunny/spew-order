package funding

import (
	"errors"
	"net/http"

	"github.com/easterthebunny/spew-order/pkg/types"
)

func NewBlockchainSource() Source {
	return &blockchainSource{}
}

type blockchainSource struct {
}

func (b *blockchainSource) Name() string {
	return "BLOCKCHAIN"
}

func (b *blockchainSource) Supports(types.Symbol) bool {
	return false
}

func (b *blockchainSource) CreateAddress(types.Symbol) (*Address, error) {
	return nil, errors.New("not implemented")
}

func (b *blockchainSource) Callback() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}
}

func (b *blockchainSource) Withdraw(*Transaction) error {
	return errors.New("not implemented")
}

func (b *blockchainSource) OKResponse() int {
	return http.StatusOK
}
