package contexts

import (
	"context"

	"github.com/easterthebunny/spew-order/pkg/types"
)

type contextKey int

const (
	ctxAuthzKey contextKey = iota
	ctxErrorKey
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
