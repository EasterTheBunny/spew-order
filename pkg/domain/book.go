package domain

import (
	"log"

	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/pkg/types"
)

type OrderBook struct {
	bir persist.BookRepository
}

func (ob *OrderBook) ExecuteOrInsertOrder(order types.Order) error {
	item := persist.NewBookItem(order)

	for {
		batch, err := ob.bir.GetHeadBatch(&item, 10)
		if err != nil {
			return err
		}

		for _, book := range batch {
			bookOrder := &book.Order
			// primary check for order owner match
			// two orders by the same owner cannot resolve each other
			// prevents a person from buying their own order
			if bookOrder.Owner == order.Owner || bookOrder.Account.String() == order.Account.String() {
				continue
			}

			tr, o := bookOrder.Resolve(order)

			// a transaction indicates that order pairing occurred
			// otherwise save the request order to the book
			if tr != nil {
				// since a transaction exists, save it
				err = ob.pairOrders(tr)
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
						bi := persist.NewBookItem(*o)
						return ob.bir.SetBookItem(&bi)
					}

					// if the ids don't match, the request order was only
					// partially filled and needs to continue through the
					// book
					if o.ID != bookOrder.ID {
						order = *o
						if err := ob.bir.DeleteBookItem(book); err != nil {
							return err
						}
						continue
					}
				}

				// in the case that there is no order returned from resolve
				// delete the book order because both orders were closed
				if o == nil {
					return ob.bir.DeleteBookItem(book)
				}

				return nil
			} else {
				ob.bir.SetBookItem(book)
			}
		}

		// if the order book is empty, insert the order
		return ob.bir.SetBookItem(&item)
	}
}

func (ob *OrderBook) pairOrders(tr *types.Transaction) error {

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
