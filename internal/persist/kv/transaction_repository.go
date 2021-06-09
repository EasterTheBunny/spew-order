package kv

import (
	"fmt"

	"github.com/easterthebunny/spew-order/internal/key"
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

func (tr *TransactionRepository) GetTransactions() (t []*persist.Transaction, err error) {

	q := persist.KVStoreQuery{
		StartOffset: transactionSubspace(*tr.account).Pack(key.Tuple{}).String()}

	attrs, err := tr.kvstore.RangeGet(&q, 0)
	if err != nil {
		return
	}

	for _, attr := range attrs {
		var bts []byte
		bts, err = tr.kvstore.Get(attr.Name)
		if err != nil {
			return
		}

		ord := &persist.Transaction{}
		err = ord.Decode(bts, encodingFromStr(attr.ContentEncoding))
		if err != nil {
			return
		}

		t = append(t, ord)
	}

	return
}
