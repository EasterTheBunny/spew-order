package domain

import (
	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/pkg/types"
	"github.com/shopspring/decimal"
)

type OrderBook struct {
	bir persist.BookRepository
	bm  *BalanceManager
}

func NewOrderBook(br persist.BookRepository) *OrderBook {
	return &OrderBook{bir: br}
}

// ExecuteOrInsertOrder takes an order and matches it from top down in the order
// book. This process will create account balance updates and update/delete
// account holds. It assumes holds exist and will return an error if they don't.
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
				// this should update balances for each applicable account and
				// before any book items or holds are removed
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
						// update the account hold for the book order to
						// match the new order amount
						smb, amt := o.Type.HoldAmount(o.Action, o.Base, o.Target)
						err = ob.bm.UpdateHoldOnAccount(&Account{ID: o.Account}, smb, amt, ky(o.HoldID))
						if err != nil {
							return err
						}

						// remove hold on incoming order since that order is filled
						err = ob.bm.RemoveHoldOnAccount(&Account{ID: order.Account}, order.Base, ky(order.HoldID))
						if err != nil {
							return err
						}

						bi := persist.NewBookItem(*o)
						return ob.bir.SetBookItem(&bi)
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
						err = ob.bm.UpdateHoldOnAccount(&Account{ID: order.Account}, smb, amt, ky(order.HoldID))
						if err != nil {
							return err
						}

						// remove hold on book order since that order is filled
						err = ob.bm.RemoveHoldOnAccount(&Account{ID: book.Order.Account}, book.Order.Base, ky(book.Order.HoldID))
						if err != nil {
							return err
						}

						if err := ob.bir.DeleteBookItem(book); err != nil {
							return err
						}
						continue
					}
				}

				// in the case that there is no order returned from resolve
				// delete the book order because both orders were closed
				if o == nil {
					// remove hold on book order since that order is filled
					err = ob.bm.RemoveHoldOnAccount(&Account{ID: book.Order.Account}, book.Order.Base, ky(book.Order.HoldID))
					if err != nil {
						return err
					}

					// remove hold on incoming order since that order is filled
					err = ob.bm.RemoveHoldOnAccount(&Account{ID: order.Account}, order.Base, ky(order.HoldID))
					if err != nil {
						return err
					}

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
	var err error

	err = ob.bm.PostToBalance(&Account{ID: tr.A.AccountID}, tr.A.AddSymbol, tr.A.AddQuantity)
	if err != nil {
		return err
	}

	err = ob.bm.PostToBalance(&Account{ID: tr.A.AccountID}, tr.A.SubSymbol, tr.A.SubQuantity.Mul(decimal.NewFromInt(-1)))
	if err != nil {
		return err
	}

	err = ob.bm.PostToBalance(&Account{ID: tr.B.AccountID}, tr.B.AddSymbol, tr.B.AddQuantity)
	if err != nil {
		return err
	}

	err = ob.bm.PostToBalance(&Account{ID: tr.B.AccountID}, tr.B.SubSymbol, tr.B.SubQuantity.Mul(decimal.NewFromInt(-1)))
	if err != nil {
		return err
	}

	return nil
}

type ky string

func (f ky) String() string {
	return string(f)
}
