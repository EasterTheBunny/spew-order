package api

import (
	"fmt"
	"testing"

	"github.com/easterthebunny/spew-order/pkg/types"
)

func TestOrderRequestFromBytes(t *testing.T) {
	data := `{"base":"BTC","target":"ETH","action":"BUY","type":%s}`
	limitType := `{"base":"BTC","name":"LIMIT","price":"0.0234","quantity":"0.0000042"}`

	b := []byte(fmt.Sprintf(data, limitType))

	or, err := OrderRequestFromBytes(b)
	if err != nil {
		t.Fatalf("error encountered: %s", err)
	}

	if or.Action != types.ActionTypeBuy {
		t.Errorf("unexpected action type")
	}

	if or.Base != types.SymbolBitcoin {
		t.Errorf("unexpected symbol")
	}

	if or.Type.Name() != "LIMIT" {
		t.Errorf("unexpected type")
	}
}

func TestOrderTypeFromMap(t *testing.T) {

	m := map[string]interface{}{
		"name":     "MARKET",
		"base":     "BTC",
		"quantity": "0.004"}

	ot, err := OrderTypeFromMap(m)
	if err != nil {
		t.Fatalf("error encountered: %s", err)
	}

	if ot.Name() != string(OrderTypeNameMARKET) {
		t.Errorf("wrong order type")
	}
}
