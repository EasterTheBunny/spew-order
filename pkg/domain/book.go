package domain

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/internal/persist/firebase"
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

	// a cancel order is defined as an executable order that already exists
	// on the order book. remove the book item and update the order
	log.Printf("deleting book item as book item was canceled: %s", item.Order.ID)
	err := ob.bir.DeleteBookItem(ctx, &item)
	if err != nil {
		if errors.Is(err, firebase.ErrNotFound) {
			return nil
		}
		return err
	}

	err = ob.bm.CancelOrder(ctx, order)
	if err != nil {
		return err
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
		return fmt.Errorf("action not allowed; book item exists: %s", order.ID)
	}

	var offset *persist.BookItem
	for {
		batch, err := ob.bir.GetHeadBatch(ctx, &item, 10, offset)
		if err != nil {
			return fmt.Errorf("ExecuteOrInsertOrder::head batch::%w", err)
		}

		// in the following cases, run the batch loop again
		// 1. a large order fills more book orders than the current batch size
		// 2. there exists a large number of market orders at the top of the book and the incoming is a market order
		newBatch := false

		for _, book := range batch {
			offset = book
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

				switch {
				case o != nil && o.ID == bookOrder.ID: // exits with return
					// if the returned order id matches the book order id
					// the order should be saved back to the book and the
					// matching process halted
					// update the account hold for the book order to
					// match the new order amount
					var updateError error

					smb, amt := o.Type.HoldAmount(o.Action, o.Base, o.Target)
					err = ob.bm.UpdateHoldOnAccount(ctx, &Account{ID: o.Account}, smb, amt, ky(o.HoldID))
					if err != nil {
						// errors should not be hard errors
						updateError = fmt.Errorf("update hold::%w, ", err)
					}

					// attempt to remove hold on fee
					if o.FeeHoldID != "" {
						err = ob.bm.RemoveHoldOnAccount(ctx, &Account{ID: o.Account}, types.SymbolCipherMtn, ky(o.FeeHoldID))
						if err != nil {
							updateError = fmt.Errorf("remove hold::%w, ", err)
						} else {
							o.FeeHoldID = ""
						}
					}

					// remove hold on incoming order since that order is filled
					smb, _ = order.Type.HoldAmount(order.Action, order.Base, order.Target)
					err = ob.bm.RemoveHoldOnAccount(ctx, &Account{ID: order.Account}, smb, ky(order.HoldID))
					if err != nil {
						updateError = fmt.Errorf("remove hold::%w, ", err)
					}

					// attempt to remove fee hold
					if order.FeeHoldID != "" {
						err = ob.bm.RemoveHoldOnAccount(ctx, &Account{ID: order.Account}, types.SymbolCipherMtn, ky(order.FeeHoldID))
						if err != nil {
							updateError = fmt.Errorf("remove hold::%w, ", err)
						} else {
							order.FeeHoldID = ""
						}
					}

					bi := persist.NewBookItem(*o)
					err = ob.bir.SetBookItem(ctx, &bi)
					if err != nil {
						updateError = fmt.Errorf("update book item::%w", err)
					}

					if updateError != nil {
						return fmt.Errorf("ExecuteOrInsertOrder::partial match on book order:%w", updateError)
					}

					return nil
				case o != nil && o.ID != bookOrder.ID: // continues loop
					// if the ids don't match, the request order was only
					// partially filled and needs to continue through the
					// book
					var updateError error

					order = *o
					// update the account hold for the incoming order to
					// match the new order amount
					smb, amt := o.Type.HoldAmount(order.Action, order.Base, order.Target)
					err = ob.bm.UpdateHoldOnAccount(ctx, &Account{ID: order.Account}, smb, amt, ky(order.HoldID))
					if err != nil {
						updateError = fmt.Errorf("update hold::%w, ", err)
					}

					// attempt to remove fee hold id
					if order.FeeHoldID != "" {
						err = ob.bm.RemoveHoldOnAccount(ctx, &Account{ID: order.Account}, types.SymbolCipherMtn, ky(order.FeeHoldID))
						if err != nil {
							updateError = fmt.Errorf("remove hold::%w, ", err)
						} else {
							order.FeeHoldID = ""
						}
					}

					// remove hold on book order since that order is filled
					smb, _ = book.Order.Type.HoldAmount(book.Order.Action, book.Order.Base, book.Order.Target)
					err = ob.bm.RemoveHoldOnAccount(ctx, &Account{ID: book.Order.Account}, smb, ky(book.Order.HoldID))
					if err != nil {
						updateError = fmt.Errorf("remove hold::%w, ", err)
					}

					// attempt to remove fee hold
					if book.Order.FeeHoldID != "" {
						err = ob.bm.RemoveHoldOnAccount(ctx, &Account{ID: book.Order.Account}, types.SymbolCipherMtn, ky(book.Order.FeeHoldID))
						if err != nil {
							updateError = fmt.Errorf("remove hold::%w, ", err)
						} else {
							book.Order.FeeHoldID = ""
						}
					}

					log.Printf("deleting book item as book item was closed: %s; and matched by %s", bookOrder.ID, o.ID)
					if err := ob.bir.DeleteBookItem(ctx, book); err != nil {
						updateError = fmt.Errorf("delete book item::%w", err)
					}

					if updateError != nil {
						return fmt.Errorf("ExecuteOrInsertOrder::partial match on incoming order:%w", updateError)
					}

					newBatch = true
					continue
				case o == nil: // exits with return
					// in the case that there is no order returned from resolve
					// delete the book order because both orders were closed
					// exits loop with return
					var updateError error

					// remove hold on book order since that order is filled
					smb, _ := book.Order.Type.HoldAmount(book.Order.Action, book.Order.Base, book.Order.Target)
					err = ob.bm.RemoveHoldOnAccount(ctx, &Account{ID: book.Order.Account}, smb, ky(book.Order.HoldID))
					if err != nil {
						updateError = fmt.Errorf("remove hold::%w, ", err)
					}

					// remove hold on fee amount
					if book.Order.FeeHoldID != "" {
						err = ob.bm.RemoveHoldOnAccount(ctx, &Account{ID: book.Order.Account}, types.SymbolCipherMtn, ky(book.Order.FeeHoldID))
						if err != nil {
							updateError = fmt.Errorf("remove hold::%w, ", err)
						} else {
							book.Order.FeeHoldID = ""
						}
					}

					// remove hold on incoming order since that order is filled
					smb, _ = order.Type.HoldAmount(order.Action, order.Base, order.Target)
					err = ob.bm.RemoveHoldOnAccount(ctx, &Account{ID: order.Account}, smb, ky(order.HoldID))
					if err != nil {
						updateError = fmt.Errorf("remove hold::%w, ", err)
					}

					// remove fee hold
					if order.FeeHoldID != "" {
						err = ob.bm.RemoveHoldOnAccount(ctx, &Account{ID: order.Account}, types.SymbolCipherMtn, ky(order.FeeHoldID))
						if err != nil {
							updateError = fmt.Errorf("remove hold::%w, ", err)
						} else {
							order.FeeHoldID = ""
						}
					}

					log.Printf("deleting book item as both orders were closed: %s; and matched by %s", bookOrder.ID, order.ID)
					err = ob.bir.DeleteBookItem(ctx, book)
					if err != nil {
						updateError = fmt.Errorf("delete book item::%w", err)
					}

					if updateError != nil {
						return fmt.Errorf("ExecuteOrInsertOrder::total match on both orders:%w", updateError)
					}

					return nil
				default:
					return nil
				}
			} else {
				switch order.OrderRequest.Type.(type) {
				case *types.MarketOrderType:
					newBatch = true
				case *types.LimitOrderType:
					newBatch = false
				}
			}
		}

		if !newBatch {
			// if the order book is empty, insert the order
			err = ob.bir.SetBookItem(ctx, &item)
			if err != nil {
				return fmt.Errorf("ExecuteOrInsertOrder::%w", err)
			}
			return nil
		}
	}
}

func (ob *OrderBook) pairOrders(ctx context.Context, tr *types.Transaction) error {
	log.Printf("maker order/account %s/%s :: taker order/account %s/%s", tr.A.Order.ID, tr.A.AccountID, tr.B.Order.ID, tr.B.AccountID)
	return ob.bm.PostTransactionToBalance(ctx, tr)
}

type ky string

func (f ky) String() string {
	return string(f)
}
