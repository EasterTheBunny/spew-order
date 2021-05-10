package types

import (
	uuid "github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
)

// Account ...
type Account struct {
	ID       uuid.UUID                  `json:"id"`
	Balances map[Symbol]decimal.Decimal `json:"-"`
}

// NewAccount ...
func NewAccount() Account {
	return Account{
		ID: uuid.NewV4()}
}

type AccountBalanceConfig struct {
	Account *Account
	Symbol  Symbol
}
