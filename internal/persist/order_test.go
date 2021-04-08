package persist

import (
	"testing"
	"time"

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
	st := NewGoogleStorageMock()
	s := NewGoogleStorage(st)

	order := newMarketBookOrder(12700, 0.01, types.ActionTypeSell)
	err := s.ExecuteOrInsertOrder(order)

	assert.NoError(t, err)
	assert.Equal(t, 1, st.Len())
}

func TestExecuteOrInsertOrder(t *testing.T) {
	st := NewGoogleStorageMock()
	s := NewGoogleStorage(st)

	// setup the data set for the later match
	base := newOrderBook(times, buyPrices, types.ActionTypeBuy)
	base = append(base, newOrderBook(times, sellPrices, types.ActionTypeSell)...)
	for _, b := range base {
		err := s.saveOrder(NewBookOrder(b))
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

/*
func TestKeyOrder_Buy(t *testing.T) {

	// buy orders should be listed highest to lowest
	testOrders := newOrderList(types.ActionTypeBuy, types.OrderTypeLimit)
	sort.Slice(testOrders, func(i, j int) bool {
		if testOrders[i].Price.GreaterThan(testOrders[j].Price) {
			return true
		}

		if testOrders[i].Price.Equal(testOrders[j].Price) {
			if testOrders[i].Timestamp.Unix() < testOrders[j].Timestamp.Unix() {
				return true
			}
		}

		return false
	})

	expected := make([]string, len(testOrders))

	for i, o := range testOrders {
		s := StoredOrder{
			Order: o,
		}
		expected[i] = s.Key().String()
	}

	result := make([]string, len(expected))
	copy(result, expected)

	sort.Strings(result)

	for i, s := range result {
		if s != expected[i] {
			assert.Equal(t, s, expected[i])
		}
	}
}

func TestKeyOrder_Sell(t *testing.T) {

	// sell orders should be listed lowest to highest
	testOrders := newOrderList(types.ActionTypeSell, types.OrderTypeLimit)
	sort.Slice(testOrders, func(i, j int) bool {
		if testOrders[i].Price.LessThan(testOrders[j].Price) {
			return true
		}

		if testOrders[i].Price.Equal(testOrders[j].Price) {
			if testOrders[i].Timestamp.Unix() < testOrders[j].Timestamp.Unix() {
				return true
			}
		}

		return false
	})

	expected := make([]string, len(testOrders))

	for i, o := range testOrders {
		s := StoredOrder{
			Order: o,
		}
		expected[i] = s.Key().String()
	}

	result := make([]string, len(expected))
	copy(result, expected)

	sort.Strings(result)

	for i, s := range result {
		assert.Equal(t, s, expected[i])
	}
}
*/

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
