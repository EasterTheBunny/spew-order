package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/pkg/types"
	"github.com/shopspring/decimal"
)

func OrderRequestFromBytes(b []byte) (or types.OrderRequest, err error) {

	var o OrderRequest
	if err = json.Unmarshal(b, &o); err != nil {
		return
	}

	err = json.Unmarshal([]byte(fmt.Sprintf(`"%s"`, string(o.Base))), &or.Base)
	if err != nil {
		return
	}

	err = json.Unmarshal([]byte(fmt.Sprintf(`"%s"`, string(o.Target))), &or.Target)
	if err != nil {
		return
	}

	err = json.Unmarshal([]byte(fmt.Sprintf(`"%s"`, string(o.Action))), &or.Action)
	if err != nil {
		return
	}

	t, ok := o.Type.(map[string]interface{})
	if !ok {
		err = errors.New("parse error")
		return
	}

	ot, err := OrderTypeFromMap(t)
	if err != nil {
		return
	}

	or.Type = ot

	return
}

func OrderTypeFromMap(m map[string]interface{}) (types.OrderType, error) {

	fieldName := reflect.TypeOf(OrderType{}).Field(0).Tag.Get("json")
	typeName := m[fieldName].(string)

	valueBytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	switch typeName {
	case string(OrderTypeNameLIMIT):
		ot := types.LimitOrderType{}
		o := LimitOrderRequest{}
		if err = json.Unmarshal(valueBytes, &o); err != nil {
			return nil, err
		}

		var err error
		p, err := decimal.NewFromString(string(o.Price))
		if err != nil {
			return nil, err
		}
		ot.Price = p

		q, err := decimal.NewFromString(string(o.Quantity))
		if err != nil {
			return nil, err
		}
		ot.Quantity = q

		err = json.Unmarshal([]byte(fmt.Sprintf(`"%s"`, string(o.Base))), &ot.Base)
		if err != nil {
			return nil, err
		}
		return &ot, nil
	case string(OrderTypeNameMARKET):
		ot := types.MarketOrderType{}
		o := MarketOrderRequest{}
		if err = json.Unmarshal(valueBytes, &o); err != nil {
			return nil, err
		}

		q, err := decimal.NewFromString(string(o.Quantity))
		if err != nil {
			return nil, err
		}
		ot.Quantity = q

		err = json.Unmarshal([]byte(fmt.Sprintf(`"%s"`, string(o.Base))), &ot.Base)
		if err != nil {
			return nil, err
		}
		return &ot, nil
	default:
		return nil, errors.New("unemplemented")
	}
}

func BuildOrderRequest(or types.OrderRequest) OrderRequest {
	out := OrderRequest{
		Action: ActionType(or.Action.String()),
		Base:   SymbolType(or.Base.String()),
		Target: SymbolType(or.Target.String()),
	}

	switch tp := or.Type.(type) {
	case *types.LimitOrderType:
		out.Type = LimitOrderRequest{
			OrderType: OrderType{Name: OrderTypeNameLIMIT},
			Base:      SymbolType(tp.Base.String()),
			Price:     CurrencyValue(tp.Price.StringFixedBank(tp.Base.RoundingPlace())),
			Quantity:  CurrencyValue(tp.Quantity.StringFixedBank(tp.Base.RoundingPlace())),
		}
	case *types.MarketOrderType:
		out.Type = MarketOrderRequest{
			OrderType: OrderType{Name: OrderTypeNameMARKET},
			Base:      SymbolType(tp.Base.String()),
			Quantity:  CurrencyValue(tp.Quantity.StringFixedBank(tp.Base.RoundingPlace())),
		}
	}

	return out
}

func StringOrderStatus(f persist.FillStatus) OrderStatus {
	switch f {
	case persist.StatusOpen:
		return OrderStatusOPEN
	case persist.StatusPartial:
		return OrderStatusPARTIAL
	case persist.StatusFilled:
		return OrderStatusFILLED
	case persist.StatusCanceled:
		return OrderStatusCANCELLED
	default:
		return ""
	}
}
