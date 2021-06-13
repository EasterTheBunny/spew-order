package domain

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/easterthebunny/spew-order/internal/funding"
	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/pkg/types"
	uuid "github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
)

var (
	ErrInsufficientBalanceForHold = errors.New("account balance too low for hold")
)

func NewBalanceManager(repo persist.AccountRepository, l persist.LedgerRepository, f funding.Source) *BalanceManager {
	return &BalanceManager{acct: repo, ledger: l, funding: f}
}

type BalanceManager struct {
	acct    persist.AccountRepository
	ledger  persist.LedgerRepository
	funding funding.Source
}

// GetAccount searches the persistance layer for an account. If one doesn't
// exist, it creates one.
func (m *BalanceManager) GetAccount(id string) (a *Account, err error) {

	a = NewAccount()
	uid, err := uuid.FromString(id)
	if err != nil {
		return
	}
	a.ID = uid

	dirty := false
	p, err := m.acct.Find(context.Background(), a.ID)

	// create the account if it wasn't found
	if err != nil {
		if errors.Is(err, persist.ErrObjectNotExist) {
			p = &persist.Account{ID: a.ID.String()}
			dirty = true
		} else {
			err = fmt.Errorf("BalanceManager::GetAccount::%w", err)
			return
		}
	}

	for _, k := range p.Addresses {
		a.Addresses[k.Symbol] = k.Address
	}

	// TODO: very inefficient method of collecting account balances; refactor
	for _, s := range a.ActiveSymbols() {
		bal, err := m.GetAvailableBalance(a, s)
		if err != nil {
			err = fmt.Errorf("BalanceManager::GetAccount::%w", err)
			return nil, err
		}

		a.Balances[s] = bal

		// check for funding address; if it doesn't exist of that symbol or it
		// is blank, create a new one
		if x, ok := a.Addresses[s]; !ok || x == "" {
			addr, err := m.funding.CreateAddress(s)
			if err == nil {
				dirty = true
				a.Addresses[s] = addr.Hash
				p.Addresses = append(p.Addresses, persist.FundingAddress{Symbol: s, Address: addr.Hash})
			}

			// log errors instead of bubbling them up
			if err != nil {
				log.Printf("BalanceManager::GetAccount::%s", err)
			}
		}
	}

	// only save the value once; protect against rapid back to back updates
	if dirty {
		err := m.acct.Save(context.Background(), p)
		if err != nil {
			err = fmt.Errorf("BalanceManager::GetAccount::%w", err)
			return nil, err
		}
	}

	return
}

// GetAvailableBalance returns the total spendable balance for a single Symbol and includes all active holds
func (m *BalanceManager) GetAvailableBalance(a *Account, s types.Symbol) (balance decimal.Decimal, err error) {

	acct := &persist.Account{ID: a.ID.String()}
	r := m.acct.Balances(acct, s)
	balance, err = m.GetPostedBalance(a, s)
	if err != nil {
		err = fmt.Errorf("BalanceManager::GetAvailableBalance::%w", err)
		return
	}

	h, err := r.FindHolds()
	if err != nil {
		err = fmt.Errorf("BalanceManager::GetAvailableBalance::%w", err)
		return
	}

	for _, hold := range h {
		balance = balance.Sub(hold.Amount)
	}

	return
}

// GetPostedBalance returns total balance for a single Symbol apart from holds and
// returns both a balance and/or an error. Because multiple threads could call this
// function at the same time, one will succeed and one will fail however a failure
// will still return a balance but the balance may not be accurate.
func (m *BalanceManager) GetPostedBalance(a *Account, s types.Symbol) (balance decimal.Decimal, err error) {

	acct := &persist.Account{ID: a.ID.String()}
	r := m.acct.Balances(acct, s)
	balance, err = r.GetBalance()
	if err != nil {
		err = fmt.Errorf("BalanceManager.GetPostedBalance::%w", err)
		return
	}

	p, err := r.FindPosts()
	if err != nil {
		err = fmt.Errorf("BalanceManager.GetPostedBalance::%w", err)
		return
	}

	changeBal := decimal.NewFromInt(0)
	// for each post, remove the posting from the account,
	// update the balance variable
	for _, post := range p {
		balance = balance.Add(post.Amount)
		changeBal = changeBal.Add(post.Amount)

		err = r.DeletePost(post)
		if err != nil {
			changeBal = changeBal.Sub(post.Amount)
		}
	}

	// update the balance if posts were found
	if len(p) > 0 {
		err = r.UpdateBalance(balance)
		if err != nil {

			// at this point, the balance has not been updated
			// some of the posts might have been deleted
			// don't attempt to update the balance again
			// re-add the amount of all deleted posts

			newPost := persist.NewBalanceItem(changeBal)
			err = r.CreatePost(newPost)
			return
		}
	}

	return
}

// SetHoldOnAccount places a hold on the account and Symbol specified
// in the case of an insufficient balance, the hold will be removed and
// an error returned.
func (m *BalanceManager) SetHoldOnAccount(a *Account, s types.Symbol, amt decimal.Decimal) (holdid string, err error) {

	acct := &persist.Account{ID: a.ID.String()}
	r := m.acct.Balances(acct, s)
	newHold := persist.NewBalanceItem(amt)
	err = r.CreateHold(newHold)
	if err != nil {
		return
	}

	balance, err := m.GetPostedBalance(a, s)
	if err != nil {
		return
	}

	activeHolds, err := r.FindHolds()
	if err != nil {
		return
	}

	for _, hold := range activeHolds {
		balance = balance.Sub(hold.Amount)

		// only calculate holds up to the point of the new hold
		// more holds could have been added in another thread
		if hold.ID == newHold.ID {
			break
		}
	}

	if balance.LessThan(decimal.NewFromInt(0)) {

		err = r.DeleteHold(ky(newHold.ID))
		if err != nil {
			return
		}

		err = ErrInsufficientBalanceForHold
		return
	}

	holdid = newHold.ID
	return
}

func (m *BalanceManager) UpdateHoldOnAccount(a *Account, s types.Symbol, amt decimal.Decimal, id persist.Key) error {

	acct := &persist.Account{ID: a.ID.String()}
	r := m.acct.Balances(acct, s)

	return r.UpdateHold(id, amt)
}

func (m *BalanceManager) RemoveHoldOnAccount(a *Account, s types.Symbol, id persist.Key) error {

	acct := &persist.Account{ID: a.ID.String()}
	r := m.acct.Balances(acct, s)

	return r.DeleteHold(id)
}

// PostAmtToBalance places a balance change record on the account and Symbol provided
// does not roll posting up to the balance and is a thread safe operation.
func (m *BalanceManager) PostAmtToBalance(a *Account, s types.Symbol, amt decimal.Decimal) error {

	acct := &persist.Account{ID: a.ID.String()}
	r := m.acct.Balances(acct, s)
	newPost := persist.NewBalanceItem(amt)
	err := r.CreatePost(newPost)
	if err != nil {
		return err
	}

	return nil
}

func (m *BalanceManager) WithdrawFunds(a *Account, s types.Symbol, amt decimal.Decimal, hash string) (t *persist.Transaction, err error) {

	hid, err := m.SetHoldOnAccount(a, s, amt)
	if err != nil {
		return
	}

	tr := funding.Transaction{
		Symbol:  s,
		Address: hash,
		Amount:  amt,
	}

	trhash, err := m.funding.Withdraw(&tr)
	if err != nil {
		return
	}

	err = m.PostAmtToBalance(a, s, amt.Mul(decimal.NewFromInt(-1)))
	if err != nil {
		return
	}

	err = m.RemoveHoldOnAccount(a, s, ky(hid))
	if err != nil {
		return
	}

	var tm = persist.NanoTime(time.Now())
	trepo := m.acct.Transactions(&persist.Account{ID: a.ID.String()})
	t = &persist.Transaction{
		Type:            persist.TransferTransactionType,
		TransactionHash: trhash,
		AddressHash:     hash,
		Symbol:          s.String(),
		Quantity:        amt.Mul(decimal.NewFromInt(-1)).StringFixedBank(s.RoundingPlace()),
		Timestamp:       tm,
	}
	err = trepo.SetTransaction(t)
	if err != nil {
		return
	}

	err = m.ledger.RecordTransfer(s, amt)
	if err != nil {
		return
	}

	return
}

func (m *BalanceManager) FundAccountByID(id uuid.UUID, s types.Symbol, amt decimal.Decimal) error {
	a, err := m.acct.Find(context.Background(), id)
	if err != nil {
		return err
	}

	r := m.acct.Balances(a, s)
	newPost := persist.NewBalanceItem(amt)
	err = r.CreatePost(newPost)
	if err != nil {
		return err
	}

	var tm = persist.NanoTime(time.Now())
	trepo := m.acct.Transactions(a)
	err = trepo.SetTransaction(&persist.Transaction{
		Type:      persist.DepositTransactionType,
		Symbol:    s.String(),
		Quantity:  amt.StringFixedBank(s.RoundingPlace()),
		Timestamp: tm,
	})
	if err != nil {
		return err
	}

	err = m.ledger.RecordDeposit(s, amt)
	if err != nil {
		return err
	}

	return nil
}

func (m *BalanceManager) FundAccountByAddress(hash string, transaction string, s types.Symbol, amt decimal.Decimal) error {
	a, err := m.acct.FindByAddress(context.Background(), hash, s)
	if err != nil {
		return err
	}

	r := m.acct.Balances(a, s)
	newPost := persist.NewBalanceItem(amt)
	err = r.CreatePost(newPost)
	if err != nil {
		return err
	}

	var tm = persist.NanoTime(time.Now())
	trepo := m.acct.Transactions(a)
	err = trepo.SetTransaction(&persist.Transaction{
		Type:            persist.DepositTransactionType,
		TransactionHash: transaction,
		AddressHash:     hash,
		Symbol:          s.String(),
		Quantity:        amt.StringFixedBank(s.RoundingPlace()),
		Timestamp:       tm,
	})
	if err != nil {
		return err
	}

	err = m.ledger.RecordDeposit(s, amt)
	if err != nil {
		return err
	}

	return nil
}

// PostTransactionToBalance creates balance updates and transaction records in the appropriate
// accounts and adds fee payments to the general ledger
func (m *BalanceManager) PostTransactionToBalance(t *types.Transaction) error {

	var err error
	var tm = time.Now()
	var filled bool

	// post amounts to balance and transactions list for account a
	for _, order := range t.Filled {
		if order.ID.String() == t.A.Order.ID.String() {
			filled = true
		}
	}
	err = m.makeOrderRecords(t.A, tm, filled)
	if err != nil {
		return err
	}

	filled = false
	// post amounts to balance and transactions list for account b
	for _, order := range t.Filled {
		if order.ID.String() == t.A.Order.ID.String() {
			filled = true
		}
	}
	err = m.makeOrderRecords(t.B, tm, filled)
	if err != nil {
		return err
	}

	return nil
}

func (m *BalanceManager) makeOrderRecords(entry types.BalanceEntry, t time.Time, filled bool) error {

	var err error
	var tm = persist.NanoTime(t)

	// post amounts to balance and transactions list for account a
	a := &Account{ID: entry.AccountID}
	err = m.PostAmtToBalance(a, entry.AddSymbol, entry.AddQuantity)
	if err != nil {
		return err
	}

	qFee := entry.FeeQuantity.StringFixedBank(entry.AddSymbol.RoundingPlace())
	qAdd := entry.AddQuantity.StringFixedBank(entry.AddSymbol.RoundingPlace())
	qSub := entry.SubQuantity.Mul(decimal.NewFromInt(-1)).StringFixedBank(entry.SubSymbol.RoundingPlace())

	acct := &persist.Account{ID: a.ID.String()}
	trepo := m.acct.Transactions(acct)
	err = trepo.SetTransaction(&persist.Transaction{
		Type:      persist.OrderTransactionType,
		OrderID:   entry.Order.ID.String(),
		Symbol:    entry.AddSymbol.String(),
		Quantity:  qAdd,
		Fee:       qFee,
		Timestamp: persist.NanoTime(time.Now()),
	})
	if err != nil {
		return err
	}

	err = m.PostAmtToBalance(a, entry.SubSymbol, entry.SubQuantity.Mul(decimal.NewFromInt(-1)))
	if err != nil {
		return err
	}

	err = trepo.SetTransaction(&persist.Transaction{
		Type:      persist.OrderTransactionType,
		OrderID:   entry.Order.ID.String(),
		Symbol:    entry.SubSymbol.String(),
		Quantity:  qSub,
		Fee:       "",
		Timestamp: persist.NanoTime(time.Now()),
	})
	if err != nil {
		return err
	}

	err = m.ledger.RecordFee(entry.AddSymbol, entry.FeeQuantity)
	if err != nil {
		return err
	}

	// update the order status and transaction list
	orepo := m.acct.Orders(acct)

	stat := persist.StatusPartial
	if filled {
		stat = persist.StatusFilled
	}

	err = orepo.UpdateOrderStatus(entry.Order.ID, stat, []string{tm.String(), qFee, qAdd, qSub})
	if err != nil {
		return err
	}

	return nil
}

// CreateOrder inserts an order into the provided account as an open order
func (m *BalanceManager) CreateOrder(a *Account, req types.OrderRequest) (types.Order, error) {
	rep := m.acct.Orders(&persist.Account{ID: a.ID.String()})

	order := types.NewOrderFromRequest(req)
	err := rep.SetOrder(&persist.Order{Status: persist.StatusOpen, Base: order})

	return order, err
}

// CancelOrder cancels an order and removes any associated holds
func (m *BalanceManager) CancelOrder(order types.Order) error {
	var err error

	rep := m.acct.Orders(&persist.Account{ID: order.Account.String()})
	if rep == nil {
		return errors.New("CancelOrder: unknown order acount")
	}

	err = rep.UpdateOrderStatus(order.ID, persist.StatusCanceled, []string{})
	if err != nil {
		err = fmt.Errorf("CancelOrder::OrderRepository::%w", err)
		return err
	}

	smb, _ := order.Type.HoldAmount(order.Action, order.Base, order.Target)
	err = m.RemoveHoldOnAccount(&Account{ID: order.Account}, smb, ky(order.HoldID))
	if err != nil {
		err = fmt.Errorf("CancelOrder::RemoveHoldOnAccount::%w", err)
		return err
	}

	return nil
}
