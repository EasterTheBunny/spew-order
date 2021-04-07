package types

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestUnmarshalOrderRequest(t *testing.T) {
	data := `{"base":"BTC","target":"ETH","action":"BUY","type":%s}`
	limitType := `{"base":"BTC","name":"LIMIT","price":"0.0234","quantity":"0.0000042"}`
	marketType := `{"base":"ETH","name":"MARKET","quantity":0.0000042}`

	var limitReq OrderRequest
	err := json.Unmarshal([]byte(fmt.Sprintf(data, limitType)), &limitReq)
	if err != nil {
		t.Errorf("error encountered: %s", err)
	}

	assert.Equal(t, limitReq.Base, SymbolBitcoin)
	assert.Equal(t, limitReq.Target, SymbolEthereum)
	assert.Equal(t, limitReq.Action, ActionTypeBuy)

	assert.IsType(t, &LimitOrderType{}, limitReq.Type)

	var markReq OrderRequest
	err = json.Unmarshal([]byte(fmt.Sprintf(data, marketType)), &markReq)
	if err != nil {
		t.Errorf("error encountered: %s", err)
	}

	assert.IsType(t, &MarketOrderType{}, markReq.Type)
}

func TestMarshalOrderRequest(t *testing.T) {

	req := OrderRequest{
		Base:   SymbolBitcoin,
		Target: SymbolEthereum,
		Action: ActionTypeBuy,
		Type: &LimitOrderType{
			Base:     SymbolBitcoin,
			Price:    decimal.NewFromFloat(0.0234),
			Quantity: decimal.NewFromFloat(0.0000042),
		},
	}

	b, err := json.Marshal(req)
	if err != nil {
		t.Errorf("error encountered: %s", err)
	}

	data := `{"action":"BUY","base":"BTC","target":"ETH","type":%s}`
	limitType := `{"base":"BTC","name":"LIMIT","price":"0.0234","quantity":"0.0000042"}`
	expected := fmt.Sprintf(data, limitType)

	assert.Equal(t, expected, string(b), "json value must match expected")
}
