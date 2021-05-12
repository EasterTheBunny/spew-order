package contexts

import (
	"context"
	"errors"

	"github.com/easterthebunny/spew-order/pkg/types"
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
)

// AttachAuthorization ...
func AttachAuthorization(ctx context.Context, auth types.Authorization) context.Context {
	return context.WithValue(ctx, ctxAuthzKey, auth)
}

// GetAuthorization ...
func GetAuthorization(ctx context.Context) *types.Authorization {
	val := ctx.Value(ctxAuthzKey)
	if val == nil {
		return nil
	}

	auth, ok := val.(types.Authorization)
	if !ok {
		return nil
	}

	return &auth
}

// AttachAccount ...
func AttachAccount(ctx context.Context, a types.Account) context.Context {
	return context.WithValue(ctx, ctxAccountKey, a)
}

// GetAccount ...
func GetAccount(ctx context.Context) *types.Account {
	val := ctx.Value(ctxAccountKey)
	if val == nil {
		return nil
	}

	a, ok := val.(types.Account)
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
