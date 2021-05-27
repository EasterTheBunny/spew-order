package funding

import (
	"context"
	"errors"
	"net/http"

	"github.com/easterthebunny/spew-order/pkg/types"
	"github.com/shopspring/decimal"
)

var (
	ErrInvalidTransaction    = errors.New("transaction detail is invalid")
	ErrRequestBodyParseError = errors.New("request body parse error")
)

type contextKey int

const (
	ctxErrorKey contextKey = iota
	ctxDataKey
)

type CallbackError struct {
	Status int
	Err    error
}

type Transaction struct {
	Symbol          types.Symbol
	TransactionHash string
	Address         string
	Amount          decimal.Decimal
}

type Address struct {
	ID   string
	Hash string
}

// Source ...
type Source interface {
	// Name returns the name of the implemented
	Name() string
	Supports(types.Symbol) bool
	Callback() func(http.Handler) http.Handler
	CreateAddress(types.Symbol) (*Address, error)
	Withdraw(*Transaction) error
	OKResponse() int
}

func attachToContext(ctx context.Context, data interface{}, err *CallbackError) context.Context {
	if err != nil {
		val := ctx.Value(ctxErrorKey)
		if val == nil {
			ctx = context.WithValue(ctx, ctxErrorKey, &err)
		}
	}

	if data != nil {
		val := ctx.Value(ctxDataKey)
		if val == nil {
			ctx = context.WithValue(ctx, ctxDataKey, err)
		}
	}

	return ctx
}

func TransactionFromContext(ctx context.Context) (tr *Transaction, err *CallbackError) {
	val := ctx.Value(ctxErrorKey)
	if val != nil {
		var ok bool
		var e CallbackError
		e, ok = val.(CallbackError)
		if ok {
			err = &e
		}
	}

	val = ctx.Value(ctxDataKey)
	if val != nil {
		if x, ok := val.(Transaction); ok {
			tr = &x
		}
	}

	return
}
