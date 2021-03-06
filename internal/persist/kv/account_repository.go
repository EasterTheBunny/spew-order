package kv

import (
	"context"
	"fmt"

	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/pkg/types"
)

type AccountRepository struct {
	kvstore persist.KVStore
}

func NewAccountRepository(store persist.KVStore) *AccountRepository {
	return &AccountRepository{kvstore: store}
}

func (r *AccountRepository) Find(ctx context.Context, id persist.Key) (account *persist.Account, err error) {

	b, err := r.kvstore.Get(accountKey(id))
	if err != nil {
		err = fmt.Errorf("Account::Find -- %w", err)
		return
	}

	attr, err := r.kvstore.Attrs(accountKey(id))
	if err != nil {
		return
	}

	account = &persist.Account{}
	err = account.Decode(b, encodingFromStr(attr.ContentEncoding))
	if err != nil {
		return
	}

	return
}

func (r *AccountRepository) FindByAddress(ctx context.Context, addr string, sym types.Symbol) (acct *persist.Account, err error) {

	b, err := r.kvstore.Get(addressKey(addr, sym))
	if err != nil {
		err = fmt.Errorf("Account::Address::Find -- %w", err)
		return
	}

	return r.Find(ctx, stringer(string(b)))
}

func (r *AccountRepository) Save(ctx context.Context, account *persist.Account) error {
	if account == nil {
		return fmt.Errorf("%w for account", persist.ErrCannotSaveNilValue)
	}

	for _, addr := range account.Addresses {
		ky := addressKey(addr.Address, addr.Symbol)
		err := r.kvstore.Set(ky, []byte(account.ID), &persist.KVStoreObjectAttrsToUpdate{
			ContentEncoding: "text/plain",
			Metadata:        make(map[string]string),
		})
		if err != nil {
			return err
		}
	}

	enc := persist.JSON
	b, err := account.Encode(enc)
	if err != nil {
		return err
	}

	attrs := persist.KVStoreObjectAttrsToUpdate{
		ContentEncoding: encodingToStr(enc),
		Metadata:        make(map[string]string),
	}

	return r.kvstore.Set(accountKey(stringer(account.ID)), b, &attrs)
}

func (r *AccountRepository) Balances(a *persist.Account, s types.Symbol) persist.BalanceRepository {
	return NewBalanceRepository(r.kvstore, a, s)
}

func (r *AccountRepository) Transactions(a *persist.Account) persist.TransactionRepository {
	return NewTransactionRepository(r.kvstore, a)
}

func (r *AccountRepository) Orders(a *persist.Account) persist.OrderRepository {
	return NewOrderRepository(r.kvstore, a)
}
