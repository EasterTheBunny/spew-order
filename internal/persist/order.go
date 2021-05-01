package persist

import (
	"encoding/json"
	"log"
	"math"

	"cloud.google.com/go/storage"
	"github.com/easterthebunny/spew-order/internal/key"
	"github.com/easterthebunny/spew-order/pkg/types"
)

// SortSwitch is a magic number for swapping the key sort of buy orders;
// does not work for math.MaxInt64 and could pose a problem for orders with
// a price larger than this current value
const SortSwitch = math.MaxInt32

// ExecuteOrInsertOrder ...
func (gs *GoogleStorage) ExecuteOrInsertOrder(order types.Order) error {
	for {
		qattrs, err := gs.newHeadBatch(order)
		if err != nil {
			return err
		}

		for _, x := range qattrs {

			book, err := gs.readOrder(x.Name)
			if err != nil {
				return err
			}

			bookOrder := &book.Order
			tr, o := bookOrder.Resolve(order)

			// a transaction indicates that order pairing occurred
			// otherwise save the request order to the book
			if tr != nil {
				// since a transaction exists, save it
				err = gs.pairOrders(tr)
				if err != nil {
					return err
				}

				// if an order was returned by the resolve process
				// determine whether it was the book order or the
				// request order
				if o != nil {
					// if the returned order id matches the book order id
					// the order should be saved back to the book and the
					// matching process halted
					if o.ID == bookOrder.ID {
						return gs.saveOrder(types.NewBookOrder(*o))
					}

					// if the ids don't match, the request order was only
					// partially filled and needs to continue through the
					// book
					if o.ID != bookOrder.ID {
						order = *o
						if err := gs.store.Delete(x.Name); err != nil {
							return err
						}
						continue
					}
				}

				// in the case that there is no order returned from resolve
				// delete the book order because both orders were closed
				if o == nil {
					return gs.store.Delete(x.Name)
				}

				return nil
			} else {
				return gs.saveOrder(types.NewBookOrder(*o))
			}
		}

		// if the order book is empty, insert the order
		return gs.saveOrder(types.NewBookOrder(order))
	}
}

func (gs *GoogleStorage) newHeadBatch(order types.Order) ([]*storage.ObjectAttrs, error) {
	s := types.NewBookOrder(order)

	query := getStorageQuery(actionKey(Subspace(s), s.ActionOrder.Action).String())
	return gs.store.RangeGet(query, 10)
}

func (gs *GoogleStorage) saveOrder(order types.BookOrder) error {

	// no match was found. proceed to insert
	data, err := json.Marshal(order)
	if err != nil {
		return err
	}

	attrs := &storage.ObjectAttrsToUpdate{Metadata: order.MetaData}

	key := Key(order).String()
	return gs.store.Set(key, data, attrs)
}

func (gs *GoogleStorage) readOrder(key string) (types.BookOrder, error) {
	var so types.BookOrder
	data, err := gs.store.Get(key)
	if err != nil {
		return so, err
	}

	err = json.Unmarshal(data, &so)
	return so, err
}

func (gs *GoogleStorage) pairOrders(tr *types.Transaction) error {

	// TODO: save the transaction
	//fmt.Printf("order 1: %s %s\n", existing.Price, existing.Quantity)
	//fmt.Printf("order 2: %s %s\n", incoming.Price, incoming.Quantity)

	/*
		s := NewStoredOrder(existing)
		err := gs.store.Delete(s.Key().String())
		if err != nil {
			return err
		}
	*/
	log.Printf("%v", tr)

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

// Key generates a key that will sort ASC lexigraphically, but remain in type
// sorted order: buys are sorted largest/oldest to smallest/newest and sells
// are sorted smallest/oldest to largest/newest
func Key(o types.BookOrder) key.Key {
	t := o.Order.Type.KeyTuple(o.Order.Action)
	t = append(t, key.Tuple{o.Order.Timestamp.UnixNano()}...)
	return actionSubspace(Subspace(o), o.Order.Action).Pack(t)
}

// ActionKey generates a key that will find a sorted match in the opposite order book
func ActionKey(o types.BookOrder) key.Key {
	return actionSubspace(Subspace(o), o.ActionOrder.Action).Pack(o.Order.Type.KeyTuple(o.Order.Action))
}

// HeadKey returns a key that can be used to range query a lexigraphically sorted set
func HeadKey(o types.BookOrder) key.Key {
	return Subspace(o).Pack(key.Tuple{uint(o.Order.Action)})
}

// Subspace ...
func Subspace(o types.BookOrder) key.Subspace {
	return gsBook.Sub(uint(o.Order.Base)).Sub(uint(o.Order.Target))
}

func actionKey(sub key.Subspace, action types.ActionType) key.Key {
	return sub.Pack(key.Tuple{uint(action)})
}

func actionSubspace(sub key.Subspace, action types.ActionType) key.Subspace {
	return sub.Sub(uint(action))
}
