package domain

import (
	"testing"

	"github.com/easterthebunny/spew-order/internal/funding"
	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/internal/persist/kv"
	"github.com/easterthebunny/spew-order/pkg/types"
	"github.com/shopspring/decimal"
)

func TestGetAvailableBalance(t *testing.T) {

	type balanceTest struct {
		action     *balanceTestItem
		actiontype int
		expected   decimal.Decimal
		err        error
	}

	tests := []balanceTest{
		{
			action:     &balanceTestItem{decimal.NewFromFloat(0.4329), a, types.SymbolBitcoin},
			actiontype: 0,
			expected:   decimal.NewFromFloat(6.79404568),
			err:        nil,
		},
		{
			action:     &balanceTestItem{decimal.NewFromFloat(0.55442), a, types.SymbolEthereum},
			actiontype: 1,
			expected:   decimal.NewFromFloat(4.22294869),
			err:        nil,
		},
		{
			action:     &balanceTestItem{decimal.NewFromFloat(0.4329), b, types.SymbolBitcoin},
			actiontype: 0,
			expected:   decimal.NewFromFloat(6.79404568),
			err:        nil,
		},
		{
			action:     &balanceTestItem{decimal.NewFromFloat(-0.2222), a, types.SymbolBitcoin},
			actiontype: 1,
			expected:   decimal.NewFromFloat(6.57184568),
			err:        nil,
		},
		{
			action:     &balanceTestItem{decimal.NewFromFloat(0.55442), b, types.SymbolEthereum},
			actiontype: 1,
			expected:   decimal.NewFromFloat(4.22294869),
			err:        nil,
		},
		{
			action:     &balanceTestItem{decimal.NewFromFloat(4.5), b, types.SymbolEthereum},
			actiontype: 0,
			expected:   decimal.NewFromFloat(4.22294869),
			err:        ErrInsufficientBalanceForHold,
		},
	}

	service := NewBalanceManager(newSeededRepo())

	for i, test := range tests {
		switch test.actiontype {
		case 0:
			_, err := service.SetHoldOnAccount(test.action.acct, test.action.sym, test.action.amt)
			if test.err == ErrInsufficientBalanceForHold && err == nil {
				t.Errorf("%d: error [%s] expected, none found", i, ErrInsufficientBalanceForHold)
				continue
			}

			if (test.err != ErrInsufficientBalanceForHold || test.err == nil) && err != nil {
				t.Errorf("%d: hold error encountered where none expected: %s", i, err)
				continue
			}
		case 1:
			err := service.PostAmtToBalance(test.action.acct, test.action.sym, test.action.amt)
			if err != nil {
				t.Errorf("%d: post error encountered where none expected: %s", i, err)
				continue
			}
		}

		bal, err := service.GetAvailableBalance(test.action.acct, test.action.sym)
		if test.err == nil && err != nil {
			t.Errorf("%d: balance error encountered where none expected: %s", i, err)
			continue
		}

		if !bal.Equal(test.expected) {
			t.Errorf("%d: unexpected balance %s; expected %s", i, bal.StringFixed(8), test.expected.StringFixed(8))
		}
	}
}

type balanceTestItem struct {
	amt  decimal.Decimal
	acct *Account
	sym  types.Symbol
}

var a = NewAccount()
var b = NewAccount()
var c = NewAccount()

var seedData = []balanceTestItem{
	{decimal.NewFromFloat(1.33452823), a, types.SymbolBitcoin},
	{decimal.NewFromFloat(5.89238922), a, types.SymbolBitcoin},
	{decimal.NewFromFloat(0.00002823), a, types.SymbolBitcoin}, // 7.22694568
	{decimal.NewFromFloat(0.00000023), a, types.SymbolEthereum},
	{decimal.NewFromFloat(2.33400023), a, types.SymbolEthereum},
	{decimal.NewFromFloat(1.33452823), a, types.SymbolEthereum}, // 3.66852869
	{decimal.NewFromFloat(1.33452823), b, types.SymbolBitcoin},
	{decimal.NewFromFloat(5.89238922), b, types.SymbolBitcoin},
	{decimal.NewFromFloat(0.00002823), b, types.SymbolBitcoin},
	{decimal.NewFromFloat(0.00000023), b, types.SymbolEthereum},
	{decimal.NewFromFloat(2.33400023), b, types.SymbolEthereum},
	{decimal.NewFromFloat(1.33452823), b, types.SymbolEthereum},
	{decimal.NewFromFloat(1.33452823), c, types.SymbolBitcoin},
	{decimal.NewFromFloat(5.89238922), c, types.SymbolBitcoin},
	{decimal.NewFromFloat(0.00002823), c, types.SymbolBitcoin},
	{decimal.NewFromFloat(0.00000023), c, types.SymbolEthereum},
	{decimal.NewFromFloat(2.33400023), c, types.SymbolEthereum},
	{decimal.NewFromFloat(1.33452823), c, types.SymbolEthereum},
}

func newSeededRepo() (persist.AccountRepository, persist.LedgerRepository, funding.Source) {
	store := persist.NewMockKVStore()
	repo := kv.NewAccountRepository(store)
	l := kv.NewLedgerRepository(store)
	f := funding.NewMockSource()

	for _, s := range seedData {
		acct := &persist.Account{ID: s.acct.ID.String()}
		b := repo.Balances(acct, s.sym)
		bal, _ := b.GetBalance()
		bal = bal.Add(s.amt)
		b.UpdateBalance(bal)
	}

	return repo, l, f
}
