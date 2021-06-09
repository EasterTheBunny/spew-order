package kv

import (
	"testing"
	"time"

	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/pkg/types"
	uuid "github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestOrders(t *testing.T) {

	s := persist.NewMockKVStore()
	id := uuid.NewV4()
	account := &persist.Account{
		ID: id.String(),
	}
	or := NewOrderRepository(s, account)

	inserts := []*persist.Order{
		{
			Status: persist.StatusCanceled,
			Base: types.Order{
				OrderRequest: types.OrderRequest{
					Base:    types.SymbolBitcoin,
					Target:  types.SymbolEthereum,
					Action:  types.ActionTypeSell,
					HoldID:  "holdid",
					Owner:   "ownerid",
					Account: id,
					Type: &types.LimitOrderType{
						Base:     types.SymbolBitcoin,
						Price:    decimal.NewFromInt(5),
						Quantity: decimal.NewFromInt(3),
					},
				},
				ID:        uuid.NewV4(),
				Timestamp: time.Now(),
			},
		},
		{
			Status: persist.StatusFilled,
			Base: types.Order{
				OrderRequest: types.OrderRequest{
					Base:    types.SymbolBitcoin,
					Target:  types.SymbolEthereum,
					Action:  types.ActionTypeBuy,
					HoldID:  "holdid",
					Owner:   "ownerid",
					Account: id,
					Type: &types.LimitOrderType{
						Base:     types.SymbolBitcoin,
						Price:    decimal.NewFromInt(5),
						Quantity: decimal.NewFromInt(3),
					},
				},
				ID:        uuid.NewV4(),
				Timestamp: time.Now(),
			},
		},
		{
			Status: persist.StatusOpen,
			Base: types.Order{
				OrderRequest: types.OrderRequest{
					Base:    types.SymbolBitcoin,
					Target:  types.SymbolEthereum,
					Action:  types.ActionTypeBuy,
					HoldID:  "holdid",
					Owner:   "ownerid",
					Account: id,
					Type: &types.LimitOrderType{
						Base:     types.SymbolBitcoin,
						Price:    decimal.NewFromInt(5),
						Quantity: decimal.NewFromInt(3),
					},
				},
				ID:        uuid.NewV4(),
				Timestamp: time.Now(),
			},
		},
		{
			Status: persist.StatusPartial,
			Base: types.Order{
				OrderRequest: types.OrderRequest{
					Base:    types.SymbolBitcoin,
					Target:  types.SymbolEthereum,
					Action:  types.ActionTypeBuy,
					HoldID:  "holdid",
					Owner:   "ownerid",
					Account: id,
					Type: &types.LimitOrderType{
						Base:     types.SymbolBitcoin,
						Price:    decimal.NewFromInt(5),
						Quantity: decimal.NewFromInt(3),
					},
				},
				ID:        uuid.NewV4(),
				Timestamp: time.Now(),
			},
		},
		{
			Status: persist.StatusPartial,
			Base: types.Order{
				OrderRequest: types.OrderRequest{
					Base:    types.SymbolBitcoin,
					Target:  types.SymbolEthereum,
					Action:  types.ActionTypeBuy,
					HoldID:  "holdid",
					Owner:   "ownerid",
					Account: id,
					Type: &types.LimitOrderType{
						Base:     types.SymbolBitcoin,
						Price:    decimal.NewFromInt(5),
						Quantity: decimal.NewFromInt(3),
					},
				},
				ID:        uuid.NewV4(),
				Timestamp: time.Now(),
			},
		},
	}

	t.Run("Set", func(t *testing.T) {
		for _, item := range inserts {
			err := or.SetOrder(item)
			assert.NoError(t, err)
		}

		assert.Equal(t, len(inserts), s.Len())
	})

	t.Run("Get", func(t *testing.T) {
		for _, item := range inserts {
			o, err := or.GetOrder(item.Base.ID)
			assert.NoError(t, err)
			assert.Equal(t, item.Status, o.Status, "order status must match")
		}
	})

	t.Run("Update", func(t *testing.T) {
		err := or.UpdateOrderStatus(inserts[len(inserts)-1].Base.ID, persist.StatusOpen, []string{})
		assert.NoError(t, err)
	})

	t.Run("GetByStatus", func(t *testing.T) {
		orders, err := or.GetOrdersByStatus(persist.StatusOpen, persist.StatusCanceled)
		assert.NoError(t, err)
		assert.Len(t, orders, 3)
	})
}
