package domain

import (
	"github.com/easterthebunny/spew-order/pkg/types"
	uuid "github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
)

type OrderMessageType string

const (
	CancelOrderMessageType OrderMessageType = "CANCEL"
	OpenOrderMessageType   OrderMessageType = "OPEN"
)

type PubSubMessage struct {
	Data []byte `json:"data"`
}

type OrderMessage struct {
	Action OrderMessageType `json:"action"`
	Order  types.Order      `json:"order"`
}

// Account ...
type Account struct {
	ID        uuid.UUID
	Balances  map[types.Symbol]decimal.Decimal
	Addresses map[types.Symbol]string
}

func (Account) ActiveSymbols() []types.Symbol {
	return []types.Symbol{types.SymbolBitcoin, types.SymbolEthereum, types.SymbolBitcoinCash, types.SymbolDogecoin}
}

// NewAccount ...
func NewAccount() *Account {
	// should get deposit addresses

	return &Account{
		ID:        uuid.NewV4(),
		Balances:  make(map[types.Symbol]decimal.Decimal),
		Addresses: make(map[types.Symbol]string)}
}
