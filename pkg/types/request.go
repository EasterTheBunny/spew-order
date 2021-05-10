package types

import (
	"encoding/json"

	"github.com/shopspring/decimal"
)

// OrderRequest represents an incoming order request.
type OrderRequest struct {
	Base   Symbol     `json:"base"`
	Target Symbol     `json:"target"`
	Action ActionType `json:"action"`
	HoldID string     `json:"holdID"`
	Type   OrderType  `json:"-"`
}

func (r OrderRequest) MarshalMap() map[string]interface{} {
	data := make(map[string]interface{})

	data["base"] = r.Base
	data["target"] = r.Target
	data["action"] = r.Action

	switch x := r.Type.(type) {
	case *MarketOrderType:
		data["type"] = *x
	case *LimitOrderType:
		data["type"] = *x
	}

	return data
}

func (r OrderRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.MarshalMap())
}

func (r *OrderRequest) UnmarshalJSON(b []byte) error {
	tp := struct {
		Base   Symbol          `json:"base"`
		Target Symbol          `json:"target"`
		Action ActionType      `json:"action"`
		Type   json.RawMessage `json:"type"`
	}{}
	if err := json.Unmarshal(b, &tp); err != nil {
		return err
	}
	r.Base = tp.Base
	r.Target = tp.Target
	r.Action = tp.Action

	name := struct {
		Name string `json:"name"`
	}{}
	if err := json.Unmarshal(tp.Type, &name); err != nil {
		return err
	}

	switch name.Name {
	case "LIMIT":
		order := struct {
			Base     Symbol          `json:"base"`
			Price    decimal.Decimal `json:"price"`
			Quantity decimal.Decimal `json:"quantity"`
		}{}
		if err := json.Unmarshal(tp.Type, &order); err != nil {
			return err
		}
		r.Type = &LimitOrderType{
			Base:     order.Base,
			Price:    order.Price,
			Quantity: order.Quantity,
		}
	case "MARKET":
		order := struct {
			Base     Symbol          `json:"base"`
			Quantity decimal.Decimal `json:"quantity"`
		}{}
		if err := json.Unmarshal(tp.Type, &order); err != nil {
			return err
		}
		r.Type = &MarketOrderType{
			Base:     order.Base,
			Quantity: order.Quantity,
		}
	}

	return nil
}
