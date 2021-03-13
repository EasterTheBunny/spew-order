package persist

import (
	"sort"
	"testing"
	"time"

	"github.com/easterthebunny/spew-order/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestMatchOrInsertOrder(t *testing.T) {
	st := NewGoogleStorageMock()
	s := NewGoogleStorage(st)

	// setup the data set for the later match
	base := newOrderList(types.ActionTypeBuy)
	base = append(base, newOrderList(types.ActionTypeSell)...)
	for _, b := range base {
		s.MatchOrInsertOrder(b)
	}

	t.Run("MatchNewSellToBuy", func(t *testing.T) {
		expected := len(base)

		order := makeOrder(0.33, 1.2, 12700, types.ActionTypeSell)
		s.MatchOrInsertOrder(order)

		assert.Equal(t, expected, st.Len())

	})

	t.Run("MatchNewBuyToSell", func(t *testing.T) {
		expected := len(base)

		order := makeOrder(0.43, 1.2, 12700, types.ActionTypeBuy)
		s.MatchOrInsertOrder(order)

		assert.Equal(t, expected, st.Len())

	})

	t.Run("InsertNewSell", func(t *testing.T) {
		expected := len(base) + 1

		order := makeOrder(0.47, 1.2, 12700, types.ActionTypeSell)
		s.MatchOrInsertOrder(order)

		assert.Equal(t, expected, st.Len())

	})

	t.Run("InsertNewBuy", func(t *testing.T) {
		expected := len(base) + 2

		order := makeOrder(0.33, 1.2, 12700, types.ActionTypeBuy)
		s.MatchOrInsertOrder(order)

		assert.Equal(t, expected, st.Len())

	})
}

func TestKeyOrder_Buy(t *testing.T) {

	// buy orders should be listed highest to lowest
	testOrders := newOrderList(types.ActionTypeBuy)
	sort.Slice(testOrders, func(i, j int) bool {
		if testOrders[i].Price > testOrders[j].Price {
			return true
		}

		if testOrders[i].Price == testOrders[j].Price {
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
	testOrders := newOrderList(types.ActionTypeSell)
	sort.Slice(testOrders, func(i, j int) bool {
		if testOrders[i].Price < testOrders[j].Price {
			return true
		}

		if testOrders[i].Price == testOrders[j].Price {
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

func makeOrder(price, quantity float64, m int64, t types.ActionType) types.Order {
	return types.Order{
		OrderRequest: types.OrderRequest{
			Base:     types.SymbolBitcoin,
			Target:   types.SymbolEthereum,
			Price:    price,
			Quantity: quantity,
			Action:   t},
		Timestamp: time.Unix(m, 0),
	}
}

func newOrderList(t types.ActionType) []types.Order {
	var l []types.Order
	var prices [][]float64

	if t == types.ActionTypeBuy {
		prices = buyPrices
	} else {
		prices = sellPrices
	}

	for i, p := range prices {
		l = append(l, makeOrder(p[0], p[1], times[i], t))
	}

	return l
}
