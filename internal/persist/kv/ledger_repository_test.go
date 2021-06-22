package kv

import (
	"testing"

	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/pkg/types"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestRecordDeposit(t *testing.T) {
	s := persist.NewMockKVStore()
	r := NewLedgerRepository(s)
	b := types.SymbolBitcoin

	var err error
	err = r.RecordDeposit(b, decimal.NewFromFloat(0.3))
	assert.NoError(t, err)

	var m map[types.Symbol]decimal.Decimal
	var v decimal.Decimal
	var ok bool

	m, err = r.GetAssetBalance(persist.Transfers)
	assert.NoError(t, err)

	v, ok = m[types.SymbolBitcoin]
	assert.Equal(t, true, ok)

	if ok {
		assert.Equal(t, "0.30000000", v.StringFixedBank(b.RoundingPlace()))
	}

	m, err = r.GetLiabilityBalance(persist.TransfersPayable)
	assert.NoError(t, err)

	v, ok = m[types.SymbolBitcoin]
	assert.Equal(t, true, ok)

	if ok {
		assert.Equal(t, "0.30000000", v.StringFixedBank(b.RoundingPlace()))
	}
}

func TestRecordTransfer(t *testing.T) {
	s := persist.NewMockKVStore()
	r := NewLedgerRepository(s)
	b := types.SymbolBitcoin

	var err error
	err = r.RecordTransfer(b, decimal.NewFromFloat(0.3))
	assert.NoError(t, err)

	var m map[types.Symbol]decimal.Decimal
	var v decimal.Decimal
	var ok bool

	m, err = r.GetAssetBalance(persist.Transfers)
	assert.NoError(t, err)

	v, ok = m[types.SymbolBitcoin]
	assert.Equal(t, true, ok)

	if ok {
		assert.Equal(t, "-0.30000000", v.StringFixedBank(b.RoundingPlace()))
	}

	m, err = r.GetLiabilityBalance(persist.TransfersPayable)
	assert.NoError(t, err)

	v, ok = m[types.SymbolBitcoin]
	assert.Equal(t, true, ok)

	if ok {
		assert.Equal(t, "-0.30000000", v.StringFixedBank(b.RoundingPlace()))
	}
}

func TestIntegratedLedger(t *testing.T) {
	s := persist.NewMockKVStore()
	r := NewLedgerRepository(s)
	btc := types.SymbolBitcoin
	eth := types.SymbolEthereum

	startETH := decimal.NewFromInt(20)
	startBTC := decimal.NewFromInt(2)

	var err error

	err = r.RecordDeposit(eth, startETH)
	assert.NoError(t, err)

	err = r.RecordDeposit(btc, startBTC)
	assert.NoError(t, err)

	err = r.RecordFee(btc, decimal.NewFromFloat(0.00002500))
	assert.NoError(t, err)

	err = r.RecordFee(eth, decimal.NewFromFloat(0.00075000))
	assert.NoError(t, err)

	err = r.RecordFee(btc, decimal.NewFromFloat(0.00001050))
	assert.NoError(t, err)

	err = r.RecordFee(eth, decimal.NewFromFloat(0.00210000))
	assert.NoError(t, err)

	type symbs struct {
		sym types.Symbol
		amt string
	}

	type balanceTest struct {
		act persist.LedgerAccount
		typ string
		val []symbs
	}

	assets := map[types.Symbol]decimal.Decimal{
		btc: decimal.NewFromInt(0),
		eth: decimal.NewFromInt(0),
	}

	liabilities := map[types.Symbol]decimal.Decimal{
		btc: decimal.NewFromInt(0),
		eth: decimal.NewFromInt(0),
	}

	tests := []balanceTest{
		{persist.Transfers, "asset", []symbs{{btc, "1.99996450"}, {eth, "19.997150000000000000"}}},
		{persist.TransfersPayable, "liability", []symbs{{btc, "1.99996450"}, {eth, "19.997150000000000000"}}},
		{persist.Cash, "asset", []symbs{{btc, "0.00003550"}, {eth, "0.002850000000000000"}}},
		{persist.Sales, "liability", []symbs{{btc, "0.00003550"}, {eth, "0.002850000000000000"}}},
	}

	for _, test := range tests {
		var mp map[types.Symbol]decimal.Decimal

		switch test.typ {
		case "asset":
			mp, err = r.GetAssetBalance(test.act)
		case "liability":
			mp, err = r.GetLiabilityBalance(test.act)
		}

		assert.NoError(t, err)

		for _, sm := range test.val {
			v, ok := mp[sm.sym]
			assert.Equal(t, true, ok)

			if ok {
				assert.Equal(t, sm.amt, v.StringFixedBank(sm.sym.RoundingPlace()))
				switch test.typ {
				case "asset":
					assets[sm.sym] = assets[sm.sym].Add(v)
				case "liability":
					liabilities[sm.sym] = liabilities[sm.sym].Add(v)
				}
			}
		}
	}

	assert.Equal(t, true, startBTC.Equal(assets[btc]))
	assert.Equal(t, true, startETH.Equal(assets[eth]))
	assert.Equal(t, true, startBTC.Equal(liabilities[btc]))
	assert.Equal(t, true, startETH.Equal(liabilities[eth]))
}
