package domain

import "github.com/easterthebunny/spew-order/pkg/types"

type OrderBookRepository interface {
	ExecuteOrInsertOrder(order types.Order) error
}

// TODO: add convenience method to build an OrderBookRepository
