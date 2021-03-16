package persist

import (
	"sort"
	"testing"
	"time"

	"github.com/easterthebunny/spew-order/pkg/types"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestExecuteOrInsertOrder(t *testing.T) {
	st := NewGoogleStorageMock()
	s := NewGoogleStorage(st)

	// setup the data set for the later match
	base := newOrderList(types.ActionTypeBuy, types.OrderTypeLimit)
	base = append(base, newOrderList(types.ActionTypeSell, types.OrderTypeLimit)...)
	for _, b := range base {
		err := s.saveOrder(NewStoredOrder(b))
		if err != nil {
			t.Fatalf("error: %s", err)
		}
	}

	t.Run("MatchMarketOrderSellToBuy", func(t *testing.T) {
		// the expectation of this new order is to match two items from the order book
		// and to remove one
		expected := len(base) - 1

		order := makeOrder(0, 1.2, 12700, types.ActionTypeSell, types.OrderTypeMarket)
		err := s.ExecuteOrInsertOrder(order)

		assert.NoError(t, err)
		assert.Equal(t, expected, st.Len())
	})

	t.Run("MatchMarketOrderBuyToSell", func(t *testing.T) {
		// the expectation of this new order is to match three items from the order book
		// and to remove two
		expected := len(base) - 3

		order := makeOrder(0, 1.2, 12700, types.ActionTypeBuy, types.OrderTypeMarket)
		s.ExecuteOrInsertOrder(order)

		assert.Equal(t, expected, st.Len())

	})

	t.Run("InsertLimitSell", func(t *testing.T) {
		expected := len(base) - 2

		order := makeOrder(0.47, 1.2, 12700, types.ActionTypeSell, types.OrderTypeLimit)
		s.ExecuteOrInsertOrder(order)

		assert.Equal(t, expected, st.Len())

	})

	t.Run("InsertLimitBuy", func(t *testing.T) {
		expected := len(base) - 1

		order := makeOrder(0.33, 1.2, 12700, types.ActionTypeBuy, types.OrderTypeLimit)
		s.ExecuteOrInsertOrder(order)

		assert.Equal(t, expected, st.Len())

	})
}

func TestMarketOrder(t *testing.T) {
	st := NewGoogleStorageMock()
	s := NewGoogleStorage(st)

	// setup the data set for the later match
	base := newOrderList(types.ActionTypeBuy, types.OrderTypeLimit)
	base = append(base, newOrderList(types.ActionTypeSell, types.OrderTypeLimit)...)
	for _, b := range base {
		err := s.saveOrder(NewStoredOrder(b))
		if err != nil {
			t.Fatalf("error: %s", err)
		}
	}

	order := makeOrder(0.33, 1.2, 12700, types.ActionTypeSell, types.OrderTypeMarket)
	o, err := s.marketOrder(NewStoredOrder(order))
	if err != nil {
		t.Fatalf("error: %s", err)
	}

	// expect that 2 pairings were created
	assert.Equal(t, 10, st.Len())

	actual, _ := o.Order.Quantity.Float64()
	assert.Equal(t, 3.22, actual)
}

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

func makeOrder(price, quantity float64, m int64, a types.ActionType, t types.OrderType) types.Order {
	return types.Order{
		OrderRequest: types.OrderRequest{
			Base:     types.SymbolBitcoin,
			Target:   types.SymbolEthereum,
			Price:    decimal.NewFromFloat(price),
			Quantity: decimal.NewFromFloat(quantity),
			Action:   a,
			Type:     t},
		Timestamp: time.Unix(m, 0),
	}
}

func newOrderList(a types.ActionType, t types.OrderType) []types.Order {
	var l []types.Order
	var prices [][]float64

	if a == types.ActionTypeBuy {
		prices = buyPrices
	} else {
		prices = sellPrices
	}

	for i, p := range prices {
		l = append(l, makeOrder(p[0], p[1], times[i], a, t))
	}

	return l
}
