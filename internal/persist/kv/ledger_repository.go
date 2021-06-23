package kv

import (
	"context"
	"fmt"
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

// RecordDeposit is run for the case when a new customer sends funds to their account.
// This makes a credit in a Transfers Payable account and a debit in a Transfers account.
func (r *LedgerRepository) RecordDeposit(ctx context.Context, s types.Symbol, amt decimal.Decimal) error {
	entry := &persist.LedgerEntry{
		Symbol:    s,
		Amount:    amt,
		Timestamp: persist.NanoTime(time.Now()),
	}

	key1 := r.ledgerAccountSubspace(persist.Transfers).Sub(int(persist.Debit))
	entry.Account = persist.Transfers
	entry.Entry = persist.Debit
	err := r.record(entry, key1)
	if err != nil {
		return err
	}

	key2 := r.ledgerAccountSubspace(persist.TransfersPayable).Sub(int(persist.Credit))
	entry.Account = persist.TransfersPayable
	entry.Entry = persist.Credit

	return r.record(entry, key2)
}

func (r *LedgerRepository) RecordTransfer(ctx context.Context, s types.Symbol, amt decimal.Decimal) error {

	entry := &persist.LedgerEntry{
		Symbol:    s,
		Amount:    amt,
		Timestamp: persist.NanoTime(time.Now()),
	}

	key1 := r.ledgerAccountSubspace(persist.Transfers).Sub(int(persist.Credit))
	entry.Account = persist.Transfers
	entry.Entry = persist.Credit
	err := r.record(entry, key1)
	if err != nil {
		return err
	}

	key2 := r.ledgerAccountSubspace(persist.TransfersPayable).Sub(int(persist.Debit))
	entry.Account = persist.TransfersPayable
	entry.Entry = persist.Debit

	return r.record(entry, key2)
}

func (r *LedgerRepository) GetLiabilityBalance(ctx context.Context, a persist.LedgerAccount) (balances map[types.Symbol]decimal.Decimal, err error) {
	balances = make(map[types.Symbol]decimal.Decimal)

	q := persist.KVStoreQuery{
		StartOffset: r.ledgerAccountSubspace(a).Pack(key.Tuple{}).String(),
	}

	attr, err := r.kvstore.RangeGet(&q, 0)
	if err != nil {
		return
	}

	for _, at := range attr {
		var bts []byte
		bts, err = r.kvstore.Get(at.Name)
		if err != nil {
			err = fmt.Errorf("Ledger::getLiabilityBalance -- %w", err)
			return
		}

		entry := &persist.LedgerEntry{}
		err = entry.Decode(bts, encodingFromStr(at.ContentEncoding))
		if err != nil {
			return
		}

		amt := entry.Amount
		if entry.Entry == persist.Debit {
			amt = decimal.NewFromInt(0).Sub(entry.Amount)
		}

		bal, ok := balances[entry.Symbol]
		if !ok {
			balances[entry.Symbol] = amt
		} else {
			balances[entry.Symbol] = bal.Add(amt)
		}
	}

	return
}

func (r *LedgerRepository) GetAssetBalance(ctx context.Context, a persist.LedgerAccount) (balances map[types.Symbol]decimal.Decimal, err error) {

	balances = make(map[types.Symbol]decimal.Decimal)

	q := persist.KVStoreQuery{
		StartOffset: r.ledgerAccountSubspace(a).Pack(key.Tuple{}).String(),
	}

	attr, err := r.kvstore.RangeGet(&q, 0)
	if err != nil {
		return
	}

	for _, at := range attr {
		var bts []byte
		bts, err = r.kvstore.Get(at.Name)
		if err != nil {
			err = fmt.Errorf("Ledger::getAssetBalance -- %w", err)
			return
		}

		entry := &persist.LedgerEntry{}
		err = entry.Decode(bts, encodingFromStr(at.ContentEncoding))
		if err != nil {
			return
		}

		amt := entry.Amount
		if entry.Entry == persist.Credit {
			amt = decimal.NewFromInt(0).Sub(entry.Amount)
		}

		bal, ok := balances[entry.Symbol]
		if !ok {
			balances[entry.Symbol] = amt
		} else {
			balances[entry.Symbol] = bal.Add(amt)
		}
	}

	return
}

func (r *LedgerRepository) RecordFee(ctx context.Context, s types.Symbol, amt decimal.Decimal) error {

	entry := &persist.LedgerEntry{
		Symbol:    s,
		Amount:    amt,
		Timestamp: persist.NanoTime(time.Now()),
	}

	key1 := r.ledgerAccountSubspace(persist.Transfers).Sub(int(persist.Credit))
	entry.Account = persist.Transfers
	entry.Entry = persist.Credit
	err := r.record(entry, key1)
	if err != nil {
		return err
	}

	key2 := r.ledgerAccountSubspace(persist.TransfersPayable).Sub(int(persist.Debit))
	entry.Account = persist.TransfersPayable
	entry.Entry = persist.Debit
	err = r.record(entry, key2)
	if err != nil {
		return err
	}

	key3 := r.ledgerAccountSubspace(persist.Cash).Sub(int(persist.Debit))
	entry.Account = persist.Cash
	entry.Entry = persist.Debit
	err = r.record(entry, key3)
	if err != nil {
		return err
	}

	key4 := r.ledgerAccountSubspace(persist.Sales).Sub(int(persist.Credit))
	entry.Account = persist.Sales
	entry.Entry = persist.Credit
	return r.record(entry, key4)
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

func (r *LedgerRepository) accountTypeSubspace(t persist.AccountType) key.Subspace {

	s := ledgerSubspace()
	switch t {
	case persist.Liability:
		return s.Sub(int(persist.Liability))
	case persist.Asset:
		return s.Sub(int(persist.Asset))
	}

	return s
}

func (r *LedgerRepository) ledgerAccountSubspace(a persist.LedgerAccount) key.Subspace {

	switch a {
	case persist.Cash:
		return r.accountTypeSubspace(persist.Asset).Sub(int(persist.Cash))
	case persist.Sales:
		return r.accountTypeSubspace(persist.Liability).Sub(int(persist.Sales))
	case persist.TransfersPayable:
		return r.accountTypeSubspace(persist.Liability).Sub(int(persist.TransfersPayable))
	case persist.Transfers:
		return r.accountTypeSubspace(persist.Asset).Sub(int(persist.Transfers))
	}

	return ledgerSubspace()
}
