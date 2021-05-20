package kv

import (
	"fmt"

	"github.com/easterthebunny/spew-order/internal/persist"
)

type TransactionRepository struct {
	kvstore persist.KVStore
	account *persist.Account
}

func NewTransactionRepository(store persist.KVStore, account *persist.Account) *TransactionRepository {
	return &TransactionRepository{kvstore: store, account: account}
}

func (tr *TransactionRepository) SetTransaction(t *persist.Transaction) error {
	if t == nil {
		return fmt.Errorf("%w for transaction", persist.ErrCannotSaveNilValue)
	}

	enc := persist.JSON
	b, err := t.Encode(enc)
	if err != nil {
		return err
	}

	attrs := persist.KVStoreObjectAttrsToUpdate{
		ContentEncoding: encodingToStr(enc),
		Metadata:        make(map[string]string),
	}

	return tr.kvstore.Set(transactionKey(*tr.account, *t), b, &attrs)
}
