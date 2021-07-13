package kv

import (
	"context"
	"testing"
	"time"

	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/pkg/types"
	uuid "github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestSetBookItem(t *testing.T) {

	s := persist.NewMockKVStore()
	r := &BookRepository{kvstore: s}

	i := persist.NewBookItem(types.NewOrderFromRequest(types.OrderRequest{
		Base:    types.SymbolBitcoin,
		Target:  types.SymbolEthereum,
		Action:  types.ActionTypeBuy,
		HoldID:  "holdid",
		Owner:   "owner",
		Account: uuid.NewV4(),
		Type: &types.MarketOrderType{
			Base:     types.SymbolBitcoin,
			Quantity: decimal.NewFromFloat(5.542),
		},
	}))

	err := r.SetBookItem(context.Background(), &i)
	assert.NoError(t, err)

	assert.Equal(t, 1, s.Len(), "kv store should have 1 item")
}

func TestDeleteBookItem(t *testing.T) {

	s := persist.NewMockKVStore()
	r := &BookRepository{kvstore: s}

	i := persist.NewBookItem(types.NewOrderFromRequest(types.OrderRequest{
		Base:    types.SymbolBitcoin,
		Target:  types.SymbolEthereum,
		Action:  types.ActionTypeBuy,
		HoldID:  "holdid",
		Owner:   "owner",
		Account: uuid.NewV4(),
		Type: &types.MarketOrderType{
			Base:     types.SymbolBitcoin,
			Quantity: decimal.NewFromFloat(5.542),
		},
	}))
	ctx := context.Background()

	err := r.SetBookItem(ctx, &i)
	assert.NoError(t, err)

	err = r.DeleteBookItem(ctx, &i)
	assert.NoError(t, err)

	assert.Equal(t, 0, s.Len(), "kv store should have 0 items")
}

func TestGetHeadBatch(t *testing.T) {

	s := persist.NewMockKVStore()
	r := &BookRepository{kvstore: s}
	ctx := context.Background()

	req := types.NewOrderFromRequest(types.OrderRequest{
		Base:    types.SymbolBitcoin,
		Target:  types.SymbolEthereum,
		Action:  types.ActionTypeBuy,
		HoldID:  "holdid",
		Owner:   "owner",
		Account: uuid.NewV4(),
		Type: &types.MarketOrderType{
			Base:     types.SymbolBitcoin,
			Quantity: decimal.NewFromFloat(5.542),
		},
	})

	var expected []persist.BookItem

	count := 10
	for x := 0; x < count; x++ {
		j := req
		j.ID = uuid.NewV4()
		j.Timestamp = time.Now()

		i := persist.NewBookItem(j)

		err := r.SetBookItem(ctx, &i)
		assert.NoError(t, err)

		expected = append(expected, i)
	}

	assert.Equal(t, count, s.Len())
	assert.Len(t, expected, count)

	batch, err := r.GetHeadBatch(ctx, &expected[len(expected)-1], count/2, nil)
	assert.NoError(t, err)

	// reverse the expected array
	for i, j := 0, len(expected)-1; i < j; i, j = i+1, j-1 {
		expected[i], expected[j] = expected[j], expected[i]
	}
	expected = expected[0:5]

	// first item in batch response should be the last item in
	for i, item := range batch {
		exp := expected[i]
		assert.Equal(t, exp.Order.ID.String(), item.Order.ID.String())
	}
}
