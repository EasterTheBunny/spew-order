package book

import (
	"github.com/easterthebunny/spew-order/internal/account"
	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/pkg/types"
)

type OrderBook interface {
	ExecuteOrInsertOrder(order types.Order) error
}

func NewGoogleOrderBook(bucket string) OrderBook {

	kvStore, _ := persist.NewGoogleKVStore(&bucket)

	return account.NewKVBookRepository(kvStore)
}

func NewMockOrderBook() OrderBook {
	return account.NewKVBookRepository(persist.NewMockKVStore())
}
