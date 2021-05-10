package types

import (
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/easterthebunny/spew-order/internal/key"
	uuid "github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
)

// SortSwitch is a magic number for swapping the key sort of buy orders;
// does not work for math.MaxInt64 and could pose a problem for orders with
// a price larger than this current value
const SortSwitch = math.MaxInt32

// Order is the complete order representation. Built by composition of the Request.
type Order struct {
	OrderRequest
	ID        uuid.UUID `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Owner     uuid.UUID `json:"owner"`
}

func NewOrder() Order {
	return Order{
		ID:        uuid.NewV4(),
		Timestamp: time.Now(),
		Owner:     uuid.NewV4()}
}

func (o Order) MarshalJSON() ([]byte, error) {
	or := make(map[string]interface{})

	or["id"] = o.ID.String()
	or["owner"] = o.Owner.String()
	or["timestamp"] = o.Timestamp.Unix()

	for k, v := range o.OrderRequest.MarshalMap() {
		or[k] = v
	}

	return json.Marshal(or)
}

func (o *Order) UnmarshalJSON(b []byte) error {

	var req OrderRequest
	if err := json.Unmarshal(b, &req); err != nil {
		return err
	}

	tp := struct {
		ID        uuid.UUID `json:"id"`
		Owner     uuid.UUID `json:"owner"`
		Timestamp int64     `json:"timestamp"`
	}{}
	if err := json.Unmarshal(b, &tp); err != nil {
		return err
	}

	o.OrderRequest = req
	o.ID = tp.ID
	o.Owner = tp.Owner
	o.Timestamp = time.Unix(tp.Timestamp, 0)

	return nil
}

// NewOrderFromRequest ...
func NewOrderFromRequest(r OrderRequest) Order {
	order := NewOrder()
	order.OrderRequest = r
	return order
}

// Resolve returns a transaction if the orders can be resolved and a new order to save
// to the book if one is produced from the resolution process.
func (o *Order) Resolve(order Order) (*Transaction, *Order) {
	tr, ot := o.Type.FillWith(order)

	// if there is a filled order, it is assumed that the requested order
	// should be closed.
	var x *Order
	if ot != nil {
		y := order
		if len(tr.Filled) > 0 {
			y = *o
		} else {
			tr.Filled = []Order{*o}
		}
		x = &y
		x.OrderRequest.Type = ot
	} else {
		if tr != nil {
			// if the returned order type is nil, both orders were filled
			// this only happens if quantities from both orders are the same
			tr.Filled = append(tr.Filled, *o)
		} else {
			return tr, &order
		}
	}

	return tr, x
}

type BalanceEntry struct {
	// AddID       uuid.UUID
	AddSymbol   Symbol
	AddQuantity decimal.Decimal
	// SubID       uuid.UUID
	SubSymbol   Symbol
	SubQuantity decimal.Decimal
}

type Transaction struct {
	A      BalanceEntry
	B      BalanceEntry
	Filled []Order
}

// OrderType ...
type OrderType interface {
	Name() string
	FillWith(Order) (*Transaction, OrderType)
	KeyTuple(ActionType) key.Tuple
	HoldAmount(tp ActionType, base Symbol, target Symbol) (Symbol, decimal.Decimal)
	String() string
}

// MarketOrderType ...
type MarketOrderType struct {
	Base     Symbol          `json:"base"`
	Quantity decimal.Decimal `json:"quantity"`
}

func (m MarketOrderType) String() string {
	return m.Quantity.StringFixed(18)
}

func calcBalanceEntry(add bool, aB ActionType, sA, sB Symbol, qA, qB, p decimal.Decimal) (Symbol, decimal.Decimal, error) {
	x := ActionTypeBuy
	y := ActionTypeSell

	if !add {
		x = ActionTypeSell
		y = ActionTypeBuy
	}

	var quantity decimal.Decimal
	switch aB {
	case x:
		if qA.GreaterThan(qB) {
			quantity = qB
		} else {
			quantity = qA
		}
		return sB, quantity.Mul(p), nil
	case y:
		if qA.GreaterThan(qB) {
			quantity = qB
		} else {
			quantity = qA
		}
		return sA, quantity, nil
	default:
		return sA, qA, fmt.Errorf("unknown action type: %d", aB)
	}
}

func buildTransaction(aB ActionType, sA, sB Symbol, qA, qB, p decimal.Decimal) Transaction {
	tr := Transaction{
		Filled: []Order{}}

	addS, addQ, _ := calcBalanceEntry(true, aB, sA, sB, qA, qB, p)
	subS, subQ, _ := calcBalanceEntry(false, aB, sA, sB, qA, qB, p)

	// the first balance entry assumes the market order is cased in a sell order
	a := BalanceEntry{
		AddSymbol:   addS,
		AddQuantity: addQ,
		SubSymbol:   subS,
		SubQuantity: subQ,
	}

	b := BalanceEntry{
		AddSymbol:   a.SubSymbol,
		AddQuantity: a.SubQuantity,
		SubSymbol:   a.AddSymbol,
		SubQuantity: a.AddQuantity,
	}

	tr.A = a
	tr.B = b

	return tr
}

// FillWith ...
func (m *MarketOrderType) FillWith(order Order) (*Transaction, OrderType) {
	// a market order cannot be filled with a market order since market orders
	// don't include a price
	spendingLimit := m.Base == order.Base

	switch req := order.Type.(type) {
	case *LimitOrderType:
		//sA := m.Base
		qA := m.Quantity
		if spendingLimit {
			//sA = order.Target
			qA = m.Quantity.Div(req.Price)
		}

		tr := buildTransaction(order.Action,
			order.Target,
			order.Base,
			qA,
			req.Quantity,
			req.Price)

		if qA.GreaterThan(req.Quantity) {
			// the book order has a higher quantity
			// make a transaction and return a new order from
			// the book order such that the book order will be
			// updated
			tr.Filled = []Order{order}

			mt := *m
			if spendingLimit {
				mt.Quantity = mt.Quantity.Sub(req.Quantity.Mul(req.Price))
			} else {
				mt.Quantity = mt.Quantity.Sub(req.Quantity)
			}

			return &tr, &mt
		}

		// if both orders have the same quantity, there are no updates to make
		// the existing book order can be removed. indicate this by including the
		// order id in the filled array and a nil order type
		// if the order type returned is nil, the caller can infer that both orders
		// are filled to completion
		if qA.Equal(req.Quantity) {
			tr.Filled = append(tr.Filled, order)
			return &tr, nil
		}

		// if the existing book order is filled, that cannot be indicated by passing
		// back the book order ID. instead, indicate that the passed in order was not
		// filled and return the updated order type for the passed in order
		ot := *req
		ot.Quantity = ot.Quantity.Sub(qA)

		return &tr, &ot
	default:
		return nil, nil
	}
}

// Name ...
func (m MarketOrderType) Name() string {
	return "MARKET"
}

// KeyTuple ...
func (m MarketOrderType) KeyTuple(t ActionType) key.Tuple {
	pr := decimal.NewFromInt(0)
	if t == ActionTypeBuy {
		pr = decimal.NewFromInt(SortSwitch)
	}
	return key.Tuple{pr.StringFixedBank(m.Base.RoundingPlace())}
}

// HoldAmount ...
func (m MarketOrderType) HoldAmount(t ActionType, base Symbol, target Symbol) (symb Symbol, amt decimal.Decimal) {
	amt = m.Quantity
	symb = m.Base

	return
}

func (m MarketOrderType) MarshalJSON() ([]byte, error) {
	data := make(map[string]interface{})

	data["base"] = m.Base
	data["quantity"] = m.Quantity
	data["name"] = m.Name()

	return json.Marshal(data)
}

func (m *MarketOrderType) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, m)
}

// LimitOrderType ...
type LimitOrderType struct {
	Base     Symbol          `json:"base"`
	Price    decimal.Decimal `json:"price"`
	Quantity decimal.Decimal `json:"quantity"`
}

func (l LimitOrderType) String() string {
	return l.Quantity.StringFixed(18)
}

// Fill ...
func (l *LimitOrderType) FillWith(order Order) (*Transaction, OrderType) {
	switch req := order.Type.(type) {
	case *MarketOrderType:
		spendingLimit := req.Base == order.Base

		qB := req.Quantity
		if spendingLimit {
			qB = req.Quantity.Div(l.Price)
		}

		tr := buildTransaction(order.Action,
			order.Target,
			order.Base,
			l.Quantity,
			qB,
			l.Price)

		if l.Quantity.GreaterThan(qB) {
			// the book order has a higher quantity
			// make a transaction and return a new order from
			// the book order such that the book order will be
			// updated
			tr.Filled = []Order{order}

			ot := *l
			if spendingLimit {
				ot.Quantity = ot.Quantity.Sub(qB)
			} else {
				ot.Quantity = ot.Quantity.Sub(req.Quantity)
			}

			return &tr, &ot
		}

		// if both orders have the same quantity, there are no updates to make
		// the existing book order can be removed. indicate this by including the
		// order id in the filled array and a nil order type
		// if the order type returned is nil, the caller can infer that both orders
		// are filled to completion
		if l.Quantity.Equal(qB) {
			tr.Filled = append(tr.Filled, order)
			return &tr, nil
		}

		// if the existing book order is filled, that cannot be indicated by passing
		// back the book order ID. instead, indicate that the passed in order was not
		// filled and return the updated order type for the passed in order
		ot := *req
		if spendingLimit {
			ot.Quantity = ot.Quantity.Sub(l.Quantity.Mul(l.Price))
		} else {
			ot.Quantity = ot.Quantity.Sub(l.Quantity)
		}

		return &tr, &ot
	case *LimitOrderType:
		switch order.Action {
		case ActionTypeBuy:
			if req.Price.LessThan(l.Price) {
				return nil, nil
			}
		case ActionTypeSell:
			if l.Price.LessThan(req.Price) {
				return nil, nil
			}
		}

		tr := buildTransaction(order.Action,
			order.Target,
			order.Base,
			l.Quantity,
			req.Quantity,
			l.Price)

		if l.Quantity.GreaterThan(req.Quantity) {
			tr.Filled = []Order{order}

			ot := *l
			ot.Quantity = ot.Quantity.Sub(req.Quantity)

			return &tr, &ot
		}

		if l.Quantity.Equal(req.Quantity) {
			tr.Filled = append(tr.Filled, order)
			return &tr, nil
		}

		ot := *req
		ot.Quantity = ot.Quantity.Sub(l.Quantity)

		return &tr, &ot
	default:
		return nil, nil
	}
}

// Name ...
func (l LimitOrderType) Name() string {
	return "LIMIT"
}

// KeyTuple ...
func (l LimitOrderType) KeyTuple(t ActionType) key.Tuple {
	pr := l.Price
	if t == ActionTypeBuy {
		pr = decimal.NewFromInt(SortSwitch).Sub(l.Price)
	}
	return key.Tuple{pr.StringFixedBank(l.Base.RoundingPlace())}
}

// HoldAmount ...
func (l LimitOrderType) HoldAmount(t ActionType, base Symbol, target Symbol) (symb Symbol, amt decimal.Decimal) {

	switch t {
	case ActionTypeBuy:
		symb = base
		amt = l.Quantity.Mul(l.Price)
	case ActionTypeSell:
		symb = target
		amt = l.Quantity
	default:
		// in the case that an action is not matched, return a giant value for safety
		symb = base
		amt = decimal.NewFromInt(math.MaxInt64)
	}

	return
}

func (l LimitOrderType) MarshalJSON() ([]byte, error) {
	data := make(map[string]interface{})

	data["base"] = l.Base
	data["price"] = l.Price
	data["quantity"] = l.Quantity
	data["name"] = l.Name()

	return json.Marshal(data)
}

func (m *LimitOrderType) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, m)
}
