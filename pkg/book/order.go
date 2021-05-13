package book

import (
	"github.com/easterthebunny/spew-order/internal/account"
	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/pkg/types"
)

type OrderBook interface {
	ExecuteOrInsertOrder(order types.Order) error
}

func NewMockOrderBook() OrderBook {
	return account.NewKVBookRepository(persist.NewMockKVStore())
}
