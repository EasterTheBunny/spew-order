package contexts

import (
	"context"
	"errors"

	"github.com/easterthebunny/spew-order/pkg/types"
)

var (
	ErrAccountNotFoundInContext = errors.New("account not found in context")
)

type contextKey int

const (
	ctxAuthzKey contextKey = iota
	ctxErrorKey
	ctxAccountKey
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
