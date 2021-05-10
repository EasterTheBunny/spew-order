package types

import (
	"encoding/json"
	"fmt"
)

const (
	priceMetaKey    = "prc"
	quantityMetaKey = "qty"
	timeMetaKey     = "tme"
)

// NewBookOrder returns a new BookOrder where the meta data for range queries
// includes the order Quantity and Timestamp
func NewBookOrder(order Order) BookOrder {
	meta := map[string]string{
		timeMetaKey: fmt.Sprintf("%d", order.Timestamp.Unix())}

	// the action order will be used to search through the opposite sorted list
	m := order
	if m.Action == ActionTypeBuy {
		m.Action = ActionTypeSell
	} else {
		m.Action = ActionTypeBuy
	}

	return BookOrder{
		Order:       order,
		ActionOrder: m,
		MetaData:    meta}
}

// BookOrder is a struct for holding an order in storage
type BookOrder struct {
	Order       Order
	ActionOrder Order
	MetaData    map[string]string
}

// MarshalJSON ...
func (o BookOrder) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.Order)
}

// UnmarshalJSON ...
func (o *BookOrder) UnmarshalJSON(b []byte) error {
	var order Order
	err := json.Unmarshal(b, &order)
	if err != nil {
		return err
	}

	so := NewBookOrder(order)
	o.Order = so.Order
	o.ActionOrder = so.ActionOrder
	o.MetaData = so.MetaData

	return nil
}
