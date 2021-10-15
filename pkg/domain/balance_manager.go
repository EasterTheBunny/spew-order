package domain

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/easterthebunny/spew-order/internal/funding"
	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/pkg/types"
	uuid "github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
)

var (
	ErrInsufficientBalanceForHold = errors.New("account balance too low for hold")
	FundNewAccounts               = true
	NewAccountFunds               = decimal.NewFromInt(5000)
)

func NewBalanceManager(repo persist.AccountRepository, l persist.LedgerRepository, f ...funding.Source) *BalanceManager {
	return &BalanceManager{acct: repo, ledger: l, funding: f}
}

type BalanceManager struct {
	acct    persist.AccountRepository
	ledger  persist.LedgerRepository
	funding []funding.Source
}

// GetAccount searches the persistance layer for an account. If one doesn't
// exist, it creates one.
func (m *BalanceManager) GetAccount(ctx context.Context, id string) (a *Account, err error) {

	a = NewAccount()
	uid, err := uuid.FromString(id)
	if err != nil {
		return
	}
	a.ID = uid

	dirty := false
	accountCreated := false

	var p *persist.Account
	p, err = m.acct.Find(ctx, a.ID)

	// create the account if it wasn't found
	if err != nil {
		if errors.Is(err, persist.ErrObjectNotExist) {
			p = &persist.Account{ID: a.ID.String()}
			dirty = true
			accountCreated = true
		} else {
			err = fmt.Errorf("BalanceManager::GetAccount::%w", err)
			return
		}
	}

	for _, k := range p.Addresses {
		a.Addresses[k.Symbol] = k.Address
	}

	// TODO: very inefficient method of collecting account balances; refactor
	var bal decimal.Decimal
	for _, s := range a.ActiveSymbols() {
		bal, err = m.GetAvailableBalance(ctx, a, s)
		if err != nil {
			err = fmt.Errorf("BalanceManager::GetAccount.GetAvailableBalance::%w", err)
			return nil, err
		}

		a.Balances[s] = bal
	}

	// only save the value once; protect against rapid back to back updates
	if dirty {
		if err := m.acct.Save(ctx, p); err != nil {
			err = fmt.Errorf("BalanceManager::GetAccount.Save::%w", err)
			return nil, err
		}

		// fund new accounts on create
		if accountCreated && FundNewAccounts {

			acct := &persist.Account{ID: a.ID.String()}
			if err := m.fundAccount(ctx, acct, types.SymbolCipherMtn, NewAccountFunds); err != nil {
				return nil, err
			}
		}
	}

	return
}

func (m *BalanceManager) GetFundingAddress(ctx context.Context, a *Account, s types.Symbol) (addr *funding.Address, err error) {
	// check for funding address; if it doesn't exist of that symbol or it
	// is blank, create a new one
	x, ok := a.Addresses[s]
	if !ok || x == "" {

		var src funding.Source
		for _, check := range m.funding {
			if check.Supports(s) {
				src = check
				break
			}
		}

		if src == nil {
			return nil, fmt.Errorf("supported funding source not available for %s", s)
		}

		addr, err = src.CreateAddress(s)
		if err == nil {
			a.Addresses[s] = addr.Hash

			var p *persist.Account
			p, err = m.acct.Find(ctx, a.ID)
			if err != nil {
				return nil, err
			}

			p.Addresses = append(p.Addresses, persist.FundingAddress{Symbol: s, Address: addr.Hash})

			err = m.acct.Save(ctx, p)
			if err != nil {
				err = fmt.Errorf("BalanceManager::GetAccount.Save::%w", err)
				return nil, err
			}
		}

		// log errors instead of bubbling them up
		if err != nil {
			return nil, err
		}
	} else {
		addr = &funding.Address{
			ID:   "",
			Hash: x,
		}
	}

	return
}

// GetAvailableBalance returns the total spendable balance for a single Symbol and includes all active holds
func (m *BalanceManager) GetAvailableBalance(ctx context.Context, a *Account, s types.Symbol) (balance decimal.Decimal, err error) {

	acct := &persist.Account{ID: a.ID.String()}
	balance, err = m.GetPostedBalance(ctx, a, s)
	if err != nil {
		err = fmt.Errorf("BalanceManager::GetAvailableBalance::%w", err)
		return
	}

	r := m.acct.Balances(acct, s)
	h, err := r.FindHolds(ctx)
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
// returns both a balance and/or an error. This function relies on the repository
// to be thread safe
func (m *BalanceManager) GetPostedBalance(ctx context.Context, a *Account, s types.Symbol) (balance decimal.Decimal, err error) {

	acct := &persist.Account{ID: a.ID.String()}
	r := m.acct.Balances(acct, s)
	balance, err = r.GetBalance(ctx)
	if err != nil {
		err = fmt.Errorf("BalanceManager.GetPostedBalance::%w", err)
		return
	}

	return
}

// SetHoldOnAccount places a hold on the account and Symbol specified
// in the case of an insufficient balance, the hold will be removed and
// an error returned.
func (m *BalanceManager) SetHoldOnAccount(ctx context.Context, a *Account, s types.Symbol, amt decimal.Decimal) (holdid string, err error) {

	acct := &persist.Account{ID: a.ID.String()}
	r := m.acct.Balances(acct, s)
	newHold := persist.NewBalanceItem(amt)
	err = r.CreateHold(ctx, newHold)
	if err != nil {
		return
	}

	balance, err := m.GetPostedBalance(ctx, a, s)
	if err != nil {
		return
	}

	activeHolds, err := r.FindHolds(ctx)
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

		err = r.DeleteHold(ctx, ky(newHold.ID))
		if err != nil {
			return
		}

		err = ErrInsufficientBalanceForHold
		return
	}

	holdid = newHold.ID
	return
}

func (m *BalanceManager) UpdateHoldOnAccount(ctx context.Context, a *Account, s types.Symbol, amt decimal.Decimal, id persist.Key) error {

	acct := &persist.Account{ID: a.ID.String()}
	r := m.acct.Balances(acct, s)

	return r.UpdateHold(ctx, id, amt)
}

func (m *BalanceManager) RemoveHoldOnAccount(ctx context.Context, a *Account, s types.Symbol, id persist.Key) error {

	acct := &persist.Account{ID: a.ID.String()}
	r := m.acct.Balances(acct, s)

	return r.DeleteHold(ctx, id)
}

// PostAmtToBalance places a balance change record on the account and Symbol provided
// does not roll posting up to the balance and is a thread safe operation.
func (m *BalanceManager) PostAmtToBalance(ctx context.Context, a *Account, s types.Symbol, amt decimal.Decimal) error {

	acct := &persist.Account{ID: a.ID.String()}
	r := m.acct.Balances(acct, s)
	return r.AddToBalance(ctx, amt)
}

func (m *BalanceManager) WithdrawFunds(ctx context.Context, a *Account, s types.Symbol, amt decimal.Decimal, hash string) (t *persist.Transaction, err error) {

	switch s {
	case types.SymbolCipherMtn:
		return nil, fmt.Errorf("unsupported withdrawal: %s", s)
	}

	hid, err := m.SetHoldOnAccount(ctx, a, s, amt)
	if err != nil {
		return
	}

	tr := funding.Transaction{
		Symbol:  s,
		Address: hash,
		Amount:  amt,
	}

	var src funding.Source
	for _, check := range m.funding {
		if check.Supports(s) {
			src = check
			break
		}
	}

	if src == nil {
		return nil, fmt.Errorf("supported funding source not available for %s", s)
	}

	trhash, err := src.Withdraw(&tr)
	if err != nil {
		return
	}

	err = m.PostAmtToBalance(ctx, a, s, amt.Mul(decimal.NewFromInt(-1)))
	if err != nil {
		return
	}

	err = m.RemoveHoldOnAccount(ctx, a, s, ky(hid))
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
	err = trepo.SetTransaction(ctx, t)
	if err != nil {
		return
	}

	err = m.ledger.RecordTransfer(ctx, s, amt)
	if err != nil {
		return
	}

	return
}

func (m *BalanceManager) FundAccountByID(ctx context.Context, id uuid.UUID, s types.Symbol, amt decimal.Decimal) error {

	a, err := m.acct.Find(ctx, id)
	if err != nil {
		return err
	}

	return m.fundAccount(ctx, a, s, amt)
}

func (m *BalanceManager) fundAccount(ctx context.Context, a *persist.Account, s types.Symbol, amt decimal.Decimal) error {
	var err error

	r := m.acct.Balances(a, s)
	err = r.AddToBalance(ctx, amt)
	if err != nil {
		return err
	}

	var tm = persist.NanoTime(time.Now())
	trepo := m.acct.Transactions(a)
	err = trepo.SetTransaction(ctx, &persist.Transaction{
		Type:      persist.DepositTransactionType,
		Symbol:    s.String(),
		Quantity:  amt.StringFixedBank(s.RoundingPlace()),
		Timestamp: tm,
	})
	if err != nil {
		return err
	}

	err = m.ledger.RecordDeposit(ctx, s, amt)
	if err != nil {
		return err
	}

	return nil
}

func (m *BalanceManager) FundAccountByAddress(ctx context.Context, hash string, transaction string, s types.Symbol, amt decimal.Decimal) error {

	a, err := m.acct.FindByAddress(ctx, hash, s)
	if err != nil {
		return err
	}

	r := m.acct.Balances(a, s)
	err = r.AddToBalance(ctx, amt)
	if err != nil {
		return err
	}

	var tm = persist.NanoTime(time.Now())
	trepo := m.acct.Transactions(a)
	err = trepo.SetTransaction(ctx, &persist.Transaction{
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

	err = m.ledger.RecordDeposit(ctx, s, amt)
	if err != nil {
		return err
	}

	return nil
}

// PostTransactionToBalance creates balance updates and transaction records in the appropriate
// accounts and adds fee payments to the general ledger
func (m *BalanceManager) PostTransactionToBalance(ctx context.Context, t *types.Transaction) error {

	var err error
	var tm = time.Now()
	var filled bool

	// post amounts to balance and transactions list for account a
	for _, order := range t.Filled {
		if order.ID.String() == t.A.Order.ID.String() {
			filled = true
		}
	}
	err = m.makeOrderRecords(ctx, t.A, tm, filled)
	if err != nil {
		return err
	}

	filled = false
	// post amounts to balance and transactions list for account b
	for _, order := range t.Filled {
		if order.ID.String() == t.B.Order.ID.String() {
			filled = true
		}
	}
	err = m.makeOrderRecords(ctx, t.B, tm, filled)
	if err != nil {
		return err
	}

	return nil
}

func (m *BalanceManager) makeOrderRecords(ctx context.Context, entry types.BalanceEntry, t time.Time, filled bool) error {

	var err error
	var tm = persist.NanoTime(t)

	// post amounts to balance and transactions list for account a
	a := &Account{ID: entry.AccountID}
	err = m.PostAmtToBalance(ctx, a, entry.AddSymbol, entry.AddQuantity)
	if err != nil {
		return err
	}

	qFee := entry.FeeQuantity.StringFixedBank(types.SymbolCipherMtn.RoundingPlace())
	qAdd := entry.AddQuantity.StringFixedBank(entry.AddSymbol.RoundingPlace())
	qSub := entry.SubQuantity.Mul(decimal.NewFromInt(-1)).StringFixedBank(entry.SubSymbol.RoundingPlace())

	acct := &persist.Account{ID: a.ID.String()}
	trepo := m.acct.Transactions(acct)
	err = trepo.SetTransaction(ctx, &persist.Transaction{
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

	err = m.PostAmtToBalance(ctx, a, entry.SubSymbol, entry.SubQuantity.Mul(decimal.NewFromInt(-1)))
	if err != nil {
		return err
	}

	err = trepo.SetTransaction(ctx, &persist.Transaction{
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

	if entry.FeeQuantity.GreaterThan(decimal.NewFromInt(0)) {

		// for flat fees apply the amount
		err = m.PostAmtToBalance(ctx, a, types.SymbolCipherMtn, entry.FeeQuantity.Mul(decimal.NewFromInt(-1)))
		if err != nil {
			return err
		}

		err = m.ledger.RecordFee(ctx, types.SymbolCipherMtn, entry.FeeQuantity)
		if err != nil {
			return err
		}
	}

	// update the order status and transaction list
	orepo := m.acct.Orders(acct)

	stat := persist.StatusPartial
	if filled {
		stat = persist.StatusFilled
	}

	err = orepo.UpdateOrderStatus(ctx, entry.Order.ID, stat, []string{tm.String(), qFee, qAdd, qSub})
	if err != nil {
		return err
	}

	return nil
}

// CreateOrder inserts an order into the provided account as an open order
func (m *BalanceManager) CreateOrder(ctx context.Context, a *Account, req types.OrderRequest) (types.Order, error) {
	rep := m.acct.Orders(&persist.Account{ID: a.ID.String()})

	order := types.NewOrderFromRequest(req)
	err := rep.SetOrder(ctx, &persist.Order{Status: persist.StatusOpen, Base: order, Transactions: [][]string{}})

	return order, err
}

// CancelOrder cancels an order and removes any associated holds
func (m *BalanceManager) CancelOrder(ctx context.Context, order types.Order) error {
	var err error

	rep := m.acct.Orders(&persist.Account{ID: order.Account.String()})
	if rep == nil {
		return errors.New("CancelOrder: unknown order acount")
	}

	err = rep.UpdateOrderStatus(context.Background(), order.ID, persist.StatusCanceled, []string{})
	if err != nil {
		err = fmt.Errorf("CancelOrder::OrderRepository::%w", err)
		return err
	}

	smb, _ := order.Type.HoldAmount(order.Action, order.Base, order.Target)
	err = m.RemoveHoldOnAccount(ctx, &Account{ID: order.Account}, smb, ky(order.HoldID))
	if err != nil {
		err = fmt.Errorf("CancelOrder::RemoveHoldOnAccount::%w", err)
		return err
	}

	// attempt to remove fee hold
	err = m.RemoveHoldOnAccount(ctx, &Account{ID: order.Account}, types.SymbolCipherMtn, ky(order.FeeHoldID))
	if err != nil {
		err = fmt.Errorf("CancelOrder::RemoveHoldOnAccount::%w", err)
		return err
	}

	return nil
}
