package domain

import (
	"context"
	"fmt"

	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/pkg/types"
)

type OrderBook struct {
	bir persist.BookRepository
	bm  *BalanceManager
}

func NewOrderBook(br persist.BookRepository, bm *BalanceManager) *OrderBook {
	return &OrderBook{bir: br, bm: bm}
}

func (ob *OrderBook) CancelOrder(ctx context.Context, order types.Order) error {
	item := persist.NewBookItem(order)

	ok, err := ob.bir.BookItemExists(ctx, &item)
	if err != nil {
		return err
	}

	// a cancel order is defined as an executable order that already exists
	// on the order book. remove the book item and update the order
	if ok {
		err := ob.bir.DeleteBookItem(ctx, &item)
		if err != nil {
			return err
		}

		err = ob.bm.CancelOrder(ctx, order)
		if err != nil {
			return err
		}

		return nil
	}

	return nil
}

// ExecuteOrInsertOrder takes an order and matches it from top down in the order
// book. This process will create account balance updates and update/delete
// account holds. It assumes holds exist and will return an error if they don't.
func (ob *OrderBook) ExecuteOrInsertOrder(ctx context.Context, order types.Order) error {
	item := persist.NewBookItem(order)

	ok, err := ob.bir.BookItemExists(ctx, &item)
	if err != nil {
		return fmt.Errorf("ExecuteOrInsertOrder::exist check::%w", err)
	}

	// maintain this function as idempotent and don't run the same action twice
	// for the same record
	if ok {
		return nil
	}

	for {
		batch, err := ob.bir.GetHeadBatch(ctx, &item, 10)
		if err != nil {
			return fmt.Errorf("ExecuteOrInsertOrder::head batch::%w", err)
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
				// this should update balances for each applicable account and
				// before any book items or holds are removed
				err = ob.pairOrders(ctx, tr)
				if err != nil {
					return fmt.Errorf("ExecuteOrInsertOrder::pair orders::%w", err)
				}

				// if an order was returned by the resolve process
				// determine whether it was the book order or the
				// request order
				if o != nil {

					// if the returned order id matches the book order id
					// the order should be saved back to the book and the
					// matching process halted
					if o.ID == bookOrder.ID {
						// update the account hold for the book order to
						// match the new order amount
						smb, amt := o.Type.HoldAmount(o.Action, o.Base, o.Target)
						err = ob.bm.UpdateHoldOnAccount(ctx, &Account{ID: o.Account}, smb, amt, ky(o.HoldID))
						if err != nil {
							return fmt.Errorf("ExecuteOrInsertOrder::update hold::%w", err)
						}

						// remove hold on incoming order since that order is filled
						smb, _ = order.Type.HoldAmount(order.Action, order.Base, order.Target)
						err = ob.bm.RemoveHoldOnAccount(ctx, &Account{ID: order.Account}, smb, ky(order.HoldID))
						if err != nil {
							return fmt.Errorf("ExecuteOrInsertOrder::remove hold::%w", err)
						}

						bi := persist.NewBookItem(*o)
						err = ob.bir.SetBookItem(ctx, &bi)
						if err != nil {
							return fmt.Errorf("ExecuteOrInsertOrder::%w", err)
						}
						return nil
					}

					// if the ids don't match, the request order was only
					// partially filled and needs to continue through the
					// book
					if o.ID != bookOrder.ID {
						// TODO: remove hold on book order
						// TODO: update hold on incoming order
						order = *o
						// update the account hold for the incoming order to
						// match the new order amount
						smb, amt := o.Type.HoldAmount(order.Action, order.Base, order.Target)
						err = ob.bm.UpdateHoldOnAccount(ctx, &Account{ID: order.Account}, smb, amt, ky(order.HoldID))
						if err != nil {
							return fmt.Errorf("ExecuteOrInsertOrder::update hold::%w", err)
						}

						// remove hold on book order since that order is filled
						smb, _ = book.Order.Type.HoldAmount(book.Order.Action, book.Order.Base, book.Order.Target)
						err = ob.bm.RemoveHoldOnAccount(ctx, &Account{ID: book.Order.Account}, smb, ky(book.Order.HoldID))
						if err != nil {
							return fmt.Errorf("ExecuteOrInsertOrder::remove hold::%w", err)
						}

						if err := ob.bir.DeleteBookItem(ctx, book); err != nil {
							return fmt.Errorf("ExecuteOrInsertOrder::%w", err)
						}
						continue
					}
				}

				// in the case that there is no order returned from resolve
				// delete the book order because both orders were closed
				if o == nil {
					// remove hold on book order since that order is filled
					smb, _ := book.Order.Type.HoldAmount(book.Order.Action, book.Order.Base, book.Order.Target)
					err = ob.bm.RemoveHoldOnAccount(ctx, &Account{ID: book.Order.Account}, smb, ky(book.Order.HoldID))
					if err != nil {
						return fmt.Errorf("ExecuteOrInsertOrder::remove hold::%w", err)
					}

					// remove hold on incoming order since that order is filled
					smb, _ = order.Type.HoldAmount(order.Action, order.Base, order.Target)
					err = ob.bm.RemoveHoldOnAccount(ctx, &Account{ID: order.Account}, smb, ky(order.HoldID))
					if err != nil {
						return fmt.Errorf("ExecuteOrInsertOrder::remove hold::%w", err)
					}

					err = ob.bir.DeleteBookItem(ctx, book)
					if err != nil {
						return fmt.Errorf("ExecuteOrInsertOrder::%w", err)
					}
					return nil
				}

				return nil
			} else {
				// TODO: not sure why this is here; should it really re-save the book item???
				return ob.bir.SetBookItem(ctx, book)
			}
		}

		// if the order book is empty, insert the order
		err = ob.bir.SetBookItem(ctx, &item)
		if err != nil {
			return fmt.Errorf("ExecuteOrInsertOrder::%w", err)
		}
		return nil
	}
}

func (ob *OrderBook) pairOrders(ctx context.Context, tr *types.Transaction) error {
	return ob.bm.PostTransactionToBalance(ctx, tr)
}

type ky string

func (f ky) String() string {
	return string(f)
}
