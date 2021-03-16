package persist

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/easterthebunny/spew-order/internal/key"
	"github.com/easterthebunny/spew-order/pkg/types"
	"github.com/shopspring/decimal"
)

// SortSwitch is a magic number for swapping the key sort of buy orders;
// does not work for math.MaxInt64 and could pose a problem for orders with
// a price larger than this current value
const SortSwitch = math.MaxInt32
const (
	priceMetaKey    = "prc"
	quantityMetaKey = "qty"
	timeMetaKey     = "tme"
)

// ExecuteOrInsertOrder ...
func (gs *GoogleStorage) ExecuteOrInsertOrder(order types.Order) error {
	var err error
	s := NewStoredOrder(order)

	// process the order based on the order type
	switch order.Type {
	case types.OrderTypeMarket:
		s, err = gs.marketOrder(s)
	case types.OrderTypeLimit:
		if ok, err := gs.executable(s); ok && err != nil {
			s, err = gs.limitOrder(s)
			if err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("unexecutable order type: %d", order.Type)
	}

	if err != nil {
		return err
	}

	return gs.saveOrder(s)
}

func (gs *GoogleStorage) saveOrder(order StoredOrder) error {

	// no match was found. proceed to insert
	data, err := json.Marshal(order)
	if err != nil {
		return err
	}

	attrs := &storage.ObjectAttrsToUpdate{Metadata: order.MetaData}

	key := order.Key().String()
	return gs.store.Set(key, data, attrs)
}

func (gs *GoogleStorage) executable(order StoredOrder) (bool, error) {

	// check the head item in the opposite action book:
	// if the order is a BUY, check the SELL action book
	query := getStorageQuery(actionKey(order.Subspace(), order.ActionOrder.Action).String())
	qattrs, err := gs.store.RangeGet(query, 1)
	if err != nil {
		return false, err
	}

	// if the incoming order is satisfied by the head order return true
	if len(qattrs) > 0 && strings.Compare(order.ActionKey().String(), qattrs[0].Name) > 0 {
		return true, nil
	}

	return false, nil
}

func (gs *GoogleStorage) limitOrder(order StoredOrder) (StoredOrder, error) {

	// a market order starts at the head of the sorted order book and
	// applies for each stored order until the market order is satisfied

	/*
		query := getStorageQuery(actionKey(order.Subspace(), order.ActionOrder.Action).String())
		qattrs, err := gs.store.RangeGet(query, 10)
		if err != nil {
			return StoredOrder{}, err
		}

		for _, x := range qattrs {
			y := x.Metadata
			priceStr, ok := y[priceMetaKey]
			qtyStr, ok := y[quantityMetaKey]
			if !ok {
				return StoredOrder{}, fmt.Errorf("missing metadata for order")
			}

			price, err := strconv.ParseFloat(priceStr, 64)
			if err != nil {
				return StoredOrder{}, err
			}

			// secondary check for price match
			// when searching through SELL orders, the incoming order price
			// MUST be larger than the available SELL
			if order.Order.Action == types.ActionTypeBuy && order.Order.Price < price {
				return StoredOrder{}, fmt.Errorf("BUY orders on the SELL book MUST have a larger price than available book orders")
			}

			if order.Order.Action == types.ActionTypeSell && order.Order.Price > price {
				return StoredOrder{}, fmt.Errorf("SELL orders on the BUY book MUST have a lower price than available book orders")
			}

			qty, err := strconv.ParseFloat(qtyStr, 64)
			if err != nil {
				return StoredOrder{}, err
			}
		}

		attrs, err := gs.store.RangeGet(query, 2)
		if err != nil {
			return err
		}
	*/

	return order, nil
}

// marketOrder ...
func (gs *GoogleStorage) marketOrder(order StoredOrder) (StoredOrder, error) {

	// a market order starts at the head of the sorted order book and
	// applies each stored order until the market order quantity is satisfied
	in := &order

	exitCondition := 0
	query := getStorageQuery(actionKey(in.Subspace(), in.ActionOrder.Action).String())
	for exitCondition <= 10 {
		exitCondition += 10
		qattrs, err := gs.store.RangeGet(query, 10)
		if err != nil {
			return *in, err
		}

		if len(qattrs) == 0 {
			return *in, fmt.Errorf("no available items for trade")
		}

		for _, x := range qattrs {
			st, err := gs.store.Get(x.Name)
			if err != nil {
				return *in, err
			}

			var o types.Order
			err = json.Unmarshal(st, &o)
			if err != nil {
				return *in, err
			}

			if o.Action == in.Order.Action {
				return *in, fmt.Errorf("cannot complete order. matching action types")
			}

			// in the case that the stored order has a larger quantity than the order coming
			// in, reduce the quantity of the stored order by the amount of the incoming order
			// and save the changes
			if o.Quantity.GreaterThan(in.Order.Quantity) {
				err = gs.pairOrders(o, in.Order)
				if err != nil {
					return *in, err
				}

				o.Quantity = o.Quantity.Sub(in.Order.Quantity)
				// return the results of the existing order so that it can be stored
				return NewStoredOrder(o), nil
			}

			err = gs.pairOrders(o, in.Order)
			if err != nil {
				return *in, err
			}

			in.Order.Quantity = in.Order.Quantity.Sub(o.Quantity)

			if in.Order.Quantity.Equal(decimal.NewFromInt(0)) {
				return *in, err
			}
		}
	}

	return *in, nil
}

func (gs *GoogleStorage) pairOrders(existing, incoming types.Order) error {

	//fmt.Printf("order 1: %s %s\n", existing.Price, existing.Quantity)
	//fmt.Printf("order 2: %s %s\n", incoming.Price, incoming.Quantity)

	s := NewStoredOrder(existing)
	err := gs.store.Delete(s.Key().String())
	if err != nil {
		return err
	}

	return nil
}

func getStorageQuery(offset string) *storage.Query {
	// get the head of the list for the opposite action type
	query := &storage.Query{
		StartOffset: offset,
		Projection:  storage.ProjectionNoACL}

	query.SetAttrSelection([]string{"Name", "MetaData", "Created"})

	return query
}

// NewStoredOrder returns a new StoredOrder where the meta data for range queries
// includes the order Quantity and Timestamp
func NewStoredOrder(order types.Order) StoredOrder {
	meta := map[string]string{
		priceMetaKey:    order.Price.String(),
		quantityMetaKey: order.Quantity.String(),
		timeMetaKey:     fmt.Sprintf("%d", order.Timestamp.Unix())}

	// the action order will be used to search through the opposite sorted list
	m := order
	if m.Action == types.ActionTypeBuy {
		m.Action = types.ActionTypeSell
	} else {
		m.Action = types.ActionTypeBuy
	}

	return StoredOrder{
		Order:       order,
		ActionOrder: m,
		MetaData:    meta}
}

// StoredOrder is a struct for holding an order in storage
type StoredOrder struct {
	Order       types.Order
	ActionOrder types.Order
	MetaData    map[string]string
}

// MarshalJSON ...
func (o StoredOrder) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.Order)
}

// UnmarshalJSON ...
func (o *StoredOrder) UnmarshalJSON(b []byte) error {
	var order types.Order
	err := json.Unmarshal(b, &order)
	if err != nil {
		return err
	}

	so := NewStoredOrder(order)
	o.Order = so.Order
	o.ActionOrder = so.ActionOrder
	o.MetaData = so.MetaData

	return nil
}

// Key generates a key that will sort ASC lexigraphically, but remain in type
// sorted order: buys are sorted largest/oldest to smallest/newest and sells
// are sorted smallest/oldest to largest/newest
func (o StoredOrder) Key() key.Key {
	pr := o.Order.Price
	if o.Order.Action == types.ActionTypeBuy {
		pr = decimal.NewFromInt(SortSwitch).Sub(o.Order.Price)
	}
	return actionSubspace(o.Subspace(), o.Order.Action).Pack(key.Tuple{pr.StringFixedBank(o.Order.Base.RoundingPlace()), o.Order.Timestamp.Unix()})
}

// ActionKey generates a key that will find a sorted match in the opposite order book
func (o StoredOrder) ActionKey() key.Key {
	pr := o.ActionOrder.Price
	if o.ActionOrder.Action == types.ActionTypeBuy {
		pr = decimal.NewFromInt(SortSwitch).Sub(o.ActionOrder.Price)
	}
	return actionSubspace(o.Subspace(), o.ActionOrder.Action).Pack(key.Tuple{pr.StringFixedBank(o.Order.Base.RoundingPlace())})
}

// HeadKey returns a key that can be used to range query a lexigraphically sorted set
func (o StoredOrder) HeadKey() key.Key {
	return o.Subspace().Pack(key.Tuple{uint(o.Order.Action)})
}

// Subspace ...
func (o StoredOrder) Subspace() key.Subspace {
	return gsRoot.Sub(uint(o.Order.Base)).Sub(uint(o.Order.Target))
}

func actionKey(sub key.Subspace, action types.ActionType) key.Key {
	return sub.Pack(key.Tuple{uint(action)})
}

func actionSubspace(sub key.Subspace, action types.ActionType) key.Subspace {
	return sub.Sub(uint(action))
}
