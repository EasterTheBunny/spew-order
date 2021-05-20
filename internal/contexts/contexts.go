package contexts

import (
	"context"
	"errors"

	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/pkg/domain"
)

var (
	ErrAccountNotFoundInContext   = errors.New("account not found in context")
	ErrAccountIDNotFoundInContext = errors.New("account id not found in context")
)

type contextKey int

const (
	ctxAuthzKey contextKey = iota
	ctxErrorKey
	ctxAccountKey
	ctxAccountIDKey
	ctxOrderKey
)

// AttachAuthorization ...
func AttachAuthorization(ctx context.Context, a persist.Authorization) context.Context {
	return context.WithValue(ctx, ctxAuthzKey, a)
}

// GetAuthorization ...
func GetAuthorization(ctx context.Context) *persist.Authorization {
	val := ctx.Value(ctxAuthzKey)
	if val == nil {
		return nil
	}

	auth, ok := val.(persist.Authorization)
	if !ok {
		return nil
	}

	return &auth
}

// AttachAccount ...
func AttachAccount(ctx context.Context, a domain.Account) context.Context {
	return context.WithValue(ctx, ctxAccountKey, a)
}

// GetAccount ...
func GetAccount(ctx context.Context) *domain.Account {
	val := ctx.Value(ctxAccountKey)
	if val == nil {
		return nil
	}

	a, ok := val.(domain.Account)
	if !ok {
		return nil
	}

	return &a
}

// AttachAccountID ...
func AttachAccountID(ctx context.Context, a string) context.Context {
	return context.WithValue(ctx, ctxAccountIDKey, a)
}

// GetAccountID ...
func GetAccountID(ctx context.Context) (id string, err error) {
	val := ctx.Value(ctxAccountIDKey)
	if val == nil {
		err = ErrAccountIDNotFoundInContext
		return
	}

	id, ok := val.(string)
	if !ok {
		err = ErrAccountIDNotFoundInContext
		return
	}

	return
}

// AttachOrder ...
func AttachOrder(ctx context.Context, a persist.Order) context.Context {
	return context.WithValue(ctx, ctxOrderKey, a)
}

// GetOrder ...
func GetOrder(ctx context.Context) *persist.Order {
	val := ctx.Value(ctxOrderKey)
	if val == nil {
		return nil
	}

	a, ok := val.(persist.Order)
	if !ok {
		return nil
	}

	return &a
}

func AttachError(ctx context.Context, err error) context.Context {
	var errs []error

	val := ctx.Value(ctxErrorKey)
	if val == nil {
		return context.WithValue(ctx, ctxErrorKey, []error{err})
	}

	errs, ok := val.([]error)
	if !ok {
		errs = []error{}
	}

	return context.WithValue(ctx, ctxErrorKey, append(errs, err))
}
