package persist

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/easterthebunny/spew-order/internal/key"
	"github.com/easterthebunny/spew-order/pkg/types"
)

// SortSwitch is a magic number for swapping the key sort of buy orders;
// does not work for math.MaxInt64 and could pose a problem for orders with
// a price larger than this current value
const SortSwitch = math.MaxInt32
const (
	priceMetaKey = "prc"
)

// MatchOrInsertOrder ...
func (gs *GoogleStorage) MatchOrInsertOrder(order types.Order) error {
	s := NewStoredOrder(order)

	// attempt a match first
	m := NewStoredOrder(order)
	if m.Action == types.ActionTypeBuy {
		m.Action = types.ActionTypeSell
	} else {
		m.Action = types.ActionTypeBuy
	}

	// get the head of the list for the opposite action type
	query := &storage.Query{
		StartOffset: m.HeadKey().String(),
		Projection:  storage.ProjectionNoACL}

	query.SetAttrSelection([]string{"Name", "MetaData", "Created"})
	qattrs, err := gs.store.RangeGet(query, 1)
	if err != nil {
		return err
	}

	if len(qattrs) > 0 && strings.Compare(m.ShortKey().String(), qattrs[0].Name) > 0 {
		return gs.ExecuteMatch(order, query)
	}

	// no match was found. proceed to insert
	data, err := json.Marshal(order)
	if err != nil {
		return err
	}

	attrs := &storage.ObjectAttrsToUpdate{Metadata: s.MetaData}

	return gs.store.Set(s.Key().String(), data, attrs)
}

// ExecuteMatch ...
func (gs *GoogleStorage) ExecuteMatch(order types.Order, query *storage.Query) error {

	/*
		fmt.Printf("match %f\n", order.Price)
		attrs, err := gs.store.RangeGet(query, 2)
		if err != nil {
			return err
		}

		for _, a := range attrs {
			fmt.Printf("%s\n", a.Metadata[priceMetaKey])
		}
	*/

	return nil
}

// NewStoredOrder returns a new StoredOrder where the meta data for range queries
// includes the order Quantity and Timestamp
func NewStoredOrder(order types.Order) StoredOrder {
	meta := map[string]string{
		priceMetaKey: fmt.Sprintf("%f", order.Price),
		"qty":        fmt.Sprintf("%f", order.Quantity),
		"time":       fmt.Sprintf("%d", order.Timestamp.Unix())}

	return StoredOrder{
		Order:    order,
		MetaData: meta}
}

// StoredOrder is a struct for holding an order in storage
type StoredOrder struct {
	types.Order
	MetaData map[string]string
}

// Key generates a key that will sort ASC lexigraphically, but remain in type
// sorted order: buys are sorted largest/oldest to smallest/newest and sells
// are sorted smallest/oldest to largest/newest
func (o StoredOrder) Key() key.Key {
	pr := o.Price
	if o.Action == types.ActionTypeBuy {
		pr = float64(SortSwitch) - o.Price
	}
	return gsRoot.Sub(uint(o.Base)).Sub(uint(o.Target)).Sub(uint(o.Action)).Pack(key.Tuple{pr, o.Timestamp.Unix()})
}

// ShortKey generates a key that will sort ASC lexigraphically, but remain in type
// sorted order: buys are sorted largest/oldest to smallest/newest and sells
// are sorted smallest/oldest to largest/newest
func (o StoredOrder) ShortKey() key.Key {
	pr := o.Price
	if o.Action == types.ActionTypeBuy {
		pr = float64(SortSwitch) - o.Price
	}
	return gsRoot.Sub(uint(o.Base)).Sub(uint(o.Target)).Sub(uint(o.Action)).Pack(key.Tuple{pr})
}

// HeadKey returns a key that can be used to range query a lexigraphically sorted set
func (o StoredOrder) HeadKey() key.Key {
	return gsRoot.Sub(uint(o.Base)).Sub(uint(o.Target)).Pack(key.Tuple{uint(o.Action)})
}
