package account

import (
	"errors"

	"github.com/easterthebunny/spew-order/internal/account"
	"github.com/easterthebunny/spew-order/pkg/types"
	"github.com/shopspring/decimal"
)

var (
	ErrInsufficientBalanceForHold = errors.New("account balance too low for hold")
)

func NewBalanceService(repo account.AccountRepository) *BalanceService {
	return &BalanceService{acct: repo}
}

type BalanceService struct {
	acct account.AccountRepository
}

// GetAvailableBalance returns the total spendable balance for a single Symbol and includes all active holds
func (m *BalanceService) GetAvailableBalance(a *types.Account, s types.Symbol) (balance decimal.Decimal, err error) {

	r := m.acct.Balances(a, s)
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
func (m *BalanceService) GetPostedBalance(a *types.Account, s types.Symbol) (balance decimal.Decimal, err error) {

	r := m.acct.Balances(a, s)
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
func (m *BalanceService) SetHoldOnAccount(a *types.Account, s types.Symbol, amt decimal.Decimal) (holdid string, err error) {

	r := m.acct.Balances(a, s)
	newHold := account.NewBalanceItem(amt)
	err = r.CreateHold(&newHold)
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
		if hold.ID.String() == newHold.ID.String() {
			break
		}
	}

	if balance.LessThan(decimal.NewFromInt(0)) {

		err = r.DeleteHold(&newHold)
		if err != nil {
			return
		}

		err = ErrInsufficientBalanceForHold
		return
	}

	holdid = newHold.ID.String()
	return
}

// PostToBalance places a balance change record on the account and Symbol provided
// does not roll posting up to the balance and is a thread safe operation.
func (m *BalanceService) PostToBalance(a *types.Account, s types.Symbol, amt decimal.Decimal) error {

	r := m.acct.Balances(a, s)
	newPost := account.NewBalanceItem(amt)
	err := r.CreatePost(&newPost)
	if err != nil {
		return err
	}

	return nil
}
