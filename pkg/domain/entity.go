package domain

import (
	"github.com/easterthebunny/spew-order/pkg/types"
	uuid "github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
)

type OrderMessage struct {
	Data []byte `json:"data"`
}

// Account ...
type Account struct {
	ID       uuid.UUID
	Balances map[types.Symbol]decimal.Decimal
}

// NewAccount ...
func NewAccount() *Account {
	return &Account{
		ID:       uuid.NewV4(),
		Balances: make(map[types.Symbol]decimal.Decimal)}
}
