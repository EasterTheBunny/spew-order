package domain

import (
	"errors"

	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/pkg/types"
	uuid "github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
)

var (
	ErrInsufficientBalanceForHold = errors.New("account balance too low for hold")
)

func NewBalanceManager(repo persist.AccountRepository) *BalanceManager {
	return &BalanceManager{acct: repo}
}

type BalanceManager struct {
	acct persist.AccountRepository
}

func (m *BalanceManager) GetAccount(id string) (a *Account, err error) {

	uid, err := uuid.FromString(id)
	if err != nil {
		return
	}

	_, err = m.acct.Find(uid)
	if err != nil {
		return
	}

	a = NewAccount()
	a.ID = uid

	return
}

// GetAvailableBalance returns the total spendable balance for a single Symbol and includes all active holds
func (m *BalanceManager) GetAvailableBalance(a *Account, s types.Symbol) (balance decimal.Decimal, err error) {

	acct := &persist.Account{ID: a.ID.String()}
	r := m.acct.Balances(acct, s)
	balance, err = m.GetPostedBalance(a, s)
	if err != nil {
		return
	}

	h, err := r.FindHolds()
	if err != nil {
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
		return
	}

	p, err := r.FindPosts()
	if err != nil {
		return
	}

	// for each post, remove the posting from the account,
	// update the balance variable
	var deleteErrors []error
	for _, post := range p {
		balance = balance.Add(post.Amount)

		err = r.DeletePost(post)
		if err != nil {
			deleteErrors = append(deleteErrors, err)
		}
	}

	// if an error is encountered deleting balance postings, don't
	// allow the balance to be updated
	// this should be fault tolerant and potentially thread safe??
	// TODO: benchmark test this
	if len(deleteErrors) > 0 {
		err = deleteErrors[0]
		return
	}

	// update the balance if posts were found
	if len(p) > 0 {
		err = r.UpdateBalance(balance)
		if err != nil {
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

// PostToBalance places a balance change record on the account and Symbol provided
// does not roll posting up to the balance and is a thread safe operation.
func (m *BalanceManager) PostToBalance(a *Account, s types.Symbol, amt decimal.Decimal) error {

	acct := &persist.Account{ID: a.ID.String()}
	r := m.acct.Balances(acct, s)
	newPost := persist.NewBalanceItem(amt)
	err := r.CreatePost(newPost)
	if err != nil {
		return err
	}

	return nil
}
