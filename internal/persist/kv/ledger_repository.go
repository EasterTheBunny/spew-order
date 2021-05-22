package kv

import (
	"time"

	"github.com/easterthebunny/spew-order/internal/key"
	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/pkg/types"
	"github.com/shopspring/decimal"
)

type LedgerRepository struct {
	kvstore persist.KVStore
}

func NewLedgerRepository(store persist.KVStore) *LedgerRepository {
	return &LedgerRepository{kvstore: store}
}

var _ persist.LedgerRepository = &LedgerRepository{}

type AccountType int

const (
	Liability AccountType = iota
	Asset
)

type LedgerAccount int

const (
	Cash LedgerAccount = iota
	Sales
	TransfersPayable
	Transfers
)

type EntryType int

const (
	Credit EntryType = iota
	Debit
)

// RecordDeposit is run for the case when a new customer sends funds to their account.
// This makes a credit in a Transfers Payable account and a debit in a Transfers account.
func (r *LedgerRepository) RecordDeposit(s types.Symbol, amt decimal.Decimal) error {
	d := time.Now()
	//
	key1 := r.ledgerAccountSubspace(Transfers).Sub(int(Debit))
	key2 := r.ledgerAccountSubspace(TransfersPayable).Sub(int(Credit))

	entry := &persist.LedgerEntry{
		Symbol:    s,
		Amount:    amt,
		Timestamp: persist.NanoTime(d),
	}

	return r.record(entry, key1, key2)
}

func (r *LedgerRepository) RecordTransfer(s types.Symbol, amt decimal.Decimal) error {

	d := time.Now()
	key2 := r.ledgerAccountSubspace(TransfersPayable).Sub(int(Debit))
	key1 := r.ledgerAccountSubspace(Transfers).Sub(int(Credit))

	entry := &persist.LedgerEntry{
		Symbol:    s,
		Amount:    amt,
		Timestamp: persist.NanoTime(d),
	}

	return r.record(entry, key1, key2)
}

func (r *LedgerRepository) RecordFee(s types.Symbol, amt decimal.Decimal) error {

	d := time.Now()
	key1 := r.ledgerAccountSubspace(Transfers).Sub(int(Debit))
	key2 := r.ledgerAccountSubspace(TransfersPayable).Sub(int(Credit))
	key3 := r.ledgerAccountSubspace(Cash).Sub(int(Debit))
	key4 := r.ledgerAccountSubspace(Sales).Sub(int(Credit))

	entry := &persist.LedgerEntry{
		Symbol:    s,
		Amount:    amt,
		Timestamp: persist.NanoTime(d),
	}

	return r.record(entry, key1, key2, key3, key4)
}

func (r *LedgerRepository) record(e *persist.LedgerEntry, keys ...key.Subspace) error {
	enc := persist.JSON
	b, err := e.Encode(enc)
	if err != nil {
		return err
	}

	attrs := persist.KVStoreObjectAttrsToUpdate{
		ContentEncoding: encodingToStr(enc),
		Metadata:        make(map[string]string),
	}

	p := key.Tuple{e.Timestamp.Value()}

	for _, k := range keys {
		err = r.kvstore.Set(k.Sub(e.Symbol.String()).Pack(p).String(), b, &attrs)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *LedgerRepository) accountTypeSubspace(t AccountType) key.Subspace {

	s := ledgerSubspace()
	switch t {
	case Liability:
		return s.Sub(int(Liability))
	case Asset:
		return s.Sub(int(Asset))
	}

	return s
}

func (r *LedgerRepository) ledgerAccountSubspace(a LedgerAccount) key.Subspace {

	switch a {
	case Cash:
		return r.accountTypeSubspace(Asset).Sub(int(Cash))
	case Sales:
		return r.accountTypeSubspace(Liability).Sub(int(Sales))
	case TransfersPayable:
		return r.accountTypeSubspace(Liability).Sub(int(TransfersPayable))
	case Transfers:
		return r.accountTypeSubspace(Asset).Sub(int(Transfers))
	}

	return ledgerSubspace()
}
