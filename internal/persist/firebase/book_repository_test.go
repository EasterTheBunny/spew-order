package firebase

import (
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/pkg/types"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestItemKey_Buys(t *testing.T) {

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

func TestItemKey_Sells(t *testing.T) {

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
				Timestamp: time.Unix(10, 0),
				OrderRequest: types.OrderRequest{
					Action: types.ActionTypeSell,
					Type:   &types.MarketOrderType{},
				},
			},
		},
		{
			Order: types.Order{
				Timestamp: time.Unix(9, 0),
				OrderRequest: types.OrderRequest{
					Action: types.ActionTypeSell,
					Type:   &types.MarketOrderType{},
				},
			},
		},
		{
			Order: types.Order{
				Timestamp: time.Unix(10, 0),
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
				Timestamp: time.Unix(8, 0),
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
				Timestamp: time.Unix(7, 0),
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
				Timestamp: time.Unix(10, 0),
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
