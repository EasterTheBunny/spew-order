package account

import (
	"errors"

	"github.com/easterthebunny/spew-order/pkg/types"
	"github.com/shopspring/decimal"
)

var (
	ErrInsufficientBalanceForHold = errors.New("account balance too low for hold")
)

type BalanceService interface {
	GetBalanceForAccount(*types.AccountBalanceConfig) (decimal.Decimal, error)
	PlaceHoldOnAccount(*types.AccountBalanceConfig, decimal.Decimal) error
	RollupHoldToBalance(*types.AccountBalanceConfig, *types.BalanceItem, decimal.Decimal) error
	PostToBalance(*types.AccountBalanceConfig, *types.BalanceItem, decimal.Decimal) error
}
