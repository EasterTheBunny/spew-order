package account

import (
	"testing"
	"time"

	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/pkg/types"
	uuid "github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func newLimitBookOrder(t int64, price, quantity float64, action types.ActionType) types.Order {
	return types.Order{
		ID:        uuid.NewV4(),
		Owner:     uuid.NewV4(),
		Timestamp: time.Unix(t, 0),
		OrderRequest: types.OrderRequest{
			Base:   types.SymbolBitcoin,
			Target: types.SymbolEthereum,
			Action: action,
			Type: &types.LimitOrderType{
				Base:     types.SymbolEthereum,
				Price:    decimal.NewFromFloat(price),
				Quantity: decimal.NewFromFloat(quantity),
			},
		},
	}
}

func newMarketBookOrder(t int64, quantity float64, action types.ActionType) types.Order {
	return types.Order{
		ID:        uuid.NewV4(),
		Owner:     uuid.NewV4(),
		Timestamp: time.Unix(t, 0),
		OrderRequest: types.OrderRequest{
			Base:   types.SymbolBitcoin,
			Target: types.SymbolEthereum,
			Action: action,
			Type: &types.MarketOrderType{
				Base:     types.SymbolEthereum,
				Quantity: decimal.NewFromFloat(quantity),
			},
		},
	}
}

func newOrderBook(times []int64, amounts [][]float64, action types.ActionType) []types.Order {
	book := make([]types.Order, len(times))

	for i, t := range times {
		book[i] = newLimitBookOrder(t, amounts[i][0], amounts[i][1], action)
	}

	return book
}

func TestExecuteOrInsertOrder_EmptyBook(t *testing.T) {
	st := persist.NewMockKVStore()
	s := NewKVBookRepository(st)

	order := newMarketBookOrder(12700, 0.01, types.ActionTypeSell)
	err := s.ExecuteOrInsertOrder(order)

	assert.NoError(t, err)
	assert.Equal(t, 1, st.Len())
}

func TestExecuteOrInsertOrder(t *testing.T) {
	st := persist.NewMockKVStore()
	s := &bookRepo{store: st}

	// setup the data set for the later match
	base := newOrderBook(times, buyPrices, types.ActionTypeBuy)
	base = append(base, newOrderBook(times, sellPrices, types.ActionTypeSell)...)
	for _, b := range base {
		err := s.saveOrder(types.NewBookOrder(b))
		if err != nil {
			t.Fatalf("error: %s", err)
		}
	}

	expected := len(base)

	t.Run("SmallMarketOrder", func(t *testing.T) {
		// the expectation of this new order is to do a partial match of one item from the order book

		order := newMarketBookOrder(12700, 0.01, types.ActionTypeSell)
		err := s.ExecuteOrInsertOrder(order)

		assert.NoError(t, err)
		assert.Equal(t, expected, st.Len())
	})

	t.Run("LargeMarketOrder_3x2", func(t *testing.T) {
		// the expectation of this new order is to match three items from the order book
		// and to remove two
		expected = expected - 2

		order := newMarketBookOrder(12700, 1.2, types.ActionTypeBuy)
		s.ExecuteOrInsertOrder(order)

		assert.Equal(t, expected, st.Len())
	})

	t.Run("InsertLimitSell", func(t *testing.T) {
		expected = expected + 1

		order := newLimitBookOrder(12700, 0.47, 1.2, types.ActionTypeSell)
		s.ExecuteOrInsertOrder(order)

		assert.Equal(t, expected, st.Len())
	})

	t.Run("InsertLimitBuy", func(t *testing.T) {
		expected = expected + 1

		order := newLimitBookOrder(12700, 0.33, 1.2, types.ActionTypeBuy)
		s.ExecuteOrInsertOrder(order)

		assert.Equal(t, expected, st.Len())
	})
}

var buyPrices = [][]float64{
	{0.38, 1.02},
	{0.37, 0.2},
	{0.37, 3.4},
	{0.35, 0.04},
	{0.35, 1.2},
	{0.35, 1.1},
}

var sellPrices = [][]float64{
	{0.39, 1.02},
	{0.39, 0.02},
	{0.40, 5.02},
	{0.41, 0.2},
	{0.42, 2.089},
	{0.45, 1.12},
}

var times = []int64{
	12344,
	12345,
	12335,
	12334,
	12345,
	12346,
}
