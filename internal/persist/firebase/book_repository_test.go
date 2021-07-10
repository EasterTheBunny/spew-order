package firebase

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/pkg/types"
	uuid "github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestItemKey_Buys_Order(t *testing.T) {

	/*
		BUY
		1:market 0:9
		0:market 0:10
		3:limit 0.9:8
		2:limit 0.9:10
		4:limit 0.7:9
	*/

	buys := []*persist.BookItem{
		{
			Order: types.Order{
				Timestamp: time.Unix(10, 0),
				OrderRequest: types.OrderRequest{
					Action: types.ActionTypeBuy,
					Type:   &types.MarketOrderType{},
				},
			},
		},
		{
			Order: types.Order{
				Timestamp: time.Unix(9, 0),
				OrderRequest: types.OrderRequest{
					Action: types.ActionTypeBuy,
					Type:   &types.MarketOrderType{},
				},
			},
		},
		{
			Order: types.Order{
				Timestamp: time.Unix(10, 0),
				OrderRequest: types.OrderRequest{
					Action: types.ActionTypeBuy,
					Type: &types.LimitOrderType{
						Price: decimal.NewFromFloat(0.9),
					},
				},
			},
		},
		{
			Order: types.Order{
				Timestamp: time.Unix(8, 0),
				OrderRequest: types.OrderRequest{
					Action: types.ActionTypeBuy,
					Type: &types.LimitOrderType{
						Price: decimal.NewFromFloat(0.9),
					},
				},
			},
		},
		{
			Order: types.Order{
				Timestamp: time.Unix(9, 0),
				OrderRequest: types.OrderRequest{
					Action: types.ActionTypeBuy,
					Type: &types.LimitOrderType{
						Price: decimal.NewFromFloat(0.7),
					},
				},
			},
		},
	}

	mp := make(map[string]int)
	keys := make([]string, len(buys))
	for i, b := range buys {
		key := itemKey(b)
		keys[i] = key
		mp[key] = i
	}

	expectedOrder := "10324"
	var sb strings.Builder
	sort.Strings(keys)
	for _, k := range keys {
		sb.WriteString(fmt.Sprintf("%d", mp[k]))
	}

	assert.Equal(t, expectedOrder, sb.String())
}

func TestItemKey_Sells_Order(t *testing.T) {

	/*
		SELL
		1:market 0:9
		0:market 0:10
		2:limit 10:10
		4:limit 11:7
		3:limit 11:8
		5:limit 15:10
	*/

	sells := []*persist.BookItem{
		{
			Order: types.Order{
				Timestamp: time.Unix(0, 1625857979098625710),
				OrderRequest: types.OrderRequest{
					Action: types.ActionTypeSell,
					Type:   &types.MarketOrderType{},
				},
			},
		},
		{
			Order: types.Order{
				Timestamp: time.Unix(0, 1625857979098625709),
				OrderRequest: types.OrderRequest{
					Action: types.ActionTypeSell,
					Type:   &types.MarketOrderType{},
				},
			},
		},
		{
			Order: types.Order{
				Timestamp: time.Unix(0, 1625857979098625710),
				OrderRequest: types.OrderRequest{
					Action: types.ActionTypeSell,
					Type: &types.LimitOrderType{
						Price: decimal.NewFromFloat(10),
					},
				},
			},
		},
		{
			Order: types.Order{
				Timestamp: time.Unix(0, 1625857979098625708),
				OrderRequest: types.OrderRequest{
					Action: types.ActionTypeSell,
					Type: &types.LimitOrderType{
						Price: decimal.NewFromFloat(11),
					},
				},
			},
		},
		{
			Order: types.Order{
				Timestamp: time.Unix(0, 1625857979098625707),
				OrderRequest: types.OrderRequest{
					Action: types.ActionTypeSell,
					Type: &types.LimitOrderType{
						Price: decimal.NewFromFloat(11),
					},
				},
			},
		},
		{
			Order: types.Order{
				Timestamp: time.Unix(0, 1625857979098625710),
				OrderRequest: types.OrderRequest{
					Action: types.ActionTypeSell,
					Type: &types.LimitOrderType{
						Price: decimal.NewFromFloat(15),
					},
				},
			},
		},
	}

	mp := make(map[string]int)
	keys := make([]string, len(sells))
	for i, b := range sells {
		key := itemKey(b)
		keys[i] = key
		mp[key] = i
	}

	expectedOrder := "102435"
	var sb strings.Builder
	sort.Strings(keys)
	for _, k := range keys {
		sb.WriteString(fmt.Sprintf("%d", mp[k]))
	}

	assert.Equal(t, expectedOrder, sb.String())
}

func TestItemKey(t *testing.T) {

	item := &persist.BookItem{
		Order: types.Order{
			Timestamp: time.Unix(0, 1625857979098625710),
			OrderRequest: types.OrderRequest{
				Action: types.ActionTypeSell,
				Type:   &types.MarketOrderType{},
			},
		},
	}

	key := itemKey(item)

	assert.Equal(t, "0.00000000.1625857979098625710", key)
}

func TestItemKeyFromMarshaledSource(t *testing.T) {
	order := types.Order{
		OrderRequest: types.OrderRequest{
			Base:    types.SymbolBitcoin,
			Target:  types.SymbolEthereum,
			Action:  types.ActionTypeSell,
			HoldID:  "",
			Owner:   "",
			Account: uuid.NewV4(),
			Type: &types.MarketOrderType{
				Base:     types.SymbolBitcoin,
				Quantity: decimal.NewFromInt(1),
			},
		},
		ID:        uuid.NewV4(),
		Timestamp: time.Unix(0, 1625857979098625710),
	}

	bytes, err := json.Marshal(order)
	assert.NoError(t, err)

	var test types.Order
	err = json.Unmarshal(bytes, &test)
	assert.NoError(t, err)

	bi := persist.NewBookItem(test)
	key := itemKey(&bi)

	assert.Equal(t, "0.00000000.1625857979098625710", key)
}
