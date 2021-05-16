package kv

import (
	"fmt"

	"github.com/easterthebunny/spew-order/internal/key"
	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/pkg/types"
)

type AccountRepository struct {
	kvstore persist.KVStore
}

func NewAccountRepository(store persist.KVStore) *AccountRepository {
	return &AccountRepository{kvstore: store}
}

var _ persist.AccountRepository = &AccountRepository{}
var _ persist.BalanceRepository = &BalanceRepository{}

const (
	bookSub int = iota
	authzSub
	accountSub
	symbolsSub
	balanceSub
	holdSub
	postSub
)

var (
	gsRoot    = key.FromBytes([]byte{0xFE})
	gsBook    = gsRoot.Sub(bookSub)
	gsAccount = gsRoot.Sub(accountSub)
	gsAuthz   = gsRoot.Sub(authzSub)
)

func (r *AccountRepository) Find(id persist.Key) (a *persist.Account, err error) {

	b, err := r.kvstore.Get(accountKey(id.String()))
	if err != nil {
		return
	}

	attr, err := r.kvstore.Attrs(accountKey(id.String()))
	if err != nil {
		return
	}

	a = &persist.Account{}
	err = a.Decode(b, encodingFromStr(attr.ContentEncoding))
	if err != nil {
		return
	}

	return a, nil
}

func (r *AccountRepository) Save(a *persist.Account) error {
	if a == nil {
		return fmt.Errorf("%w for account", persist.ErrCannotSaveNilValue)
	}

	enc := persist.JSON
	b, err := a.Encode(enc)
	if err != nil {
		return err
	}

	attrs := persist.KVStoreObjectAttrsToUpdate{
		ContentEncoding: encodingToStr(enc),
		Metadata:        make(map[string]string),
	}

	return r.kvstore.Set(accountKey(a.ID), b, &attrs)
}

func (r *AccountRepository) Balances(a *persist.Account, s types.Symbol) persist.BalanceRepository {
	return NewBalanceRepository(r.kvstore, a, s)
}

func NewBalanceRepository(kv persist.KVStore, a *persist.Account, s types.Symbol) *BalanceRepository {
	return &BalanceRepository{
		kvstore: kv,
		account: a,
		symbol:  s,
	}
}

func accountKey(id string) string {
	return gsAccount.Pack(key.Tuple{id}).String()
}

func encodingFromStr(str string) persist.EncodingType {
	var encoding persist.EncodingType
	switch str {
	case persist.JSONEncodingTypeName:
		encoding = persist.JSON
	case persist.GOBEncodingTypeName:
		encoding = persist.GOB
	default:
		encoding = persist.JSON
	}

	return encoding
}

func encodingToStr(encoding persist.EncodingType) string {
	switch encoding {
	case persist.JSON:
		return persist.JSONEncodingTypeName
	case persist.GOB:
		return persist.GOBEncodingTypeName
	default:
		return persist.JSONEncodingTypeName
	}
}
