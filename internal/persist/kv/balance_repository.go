package kv

import (
	"encoding/json"
	"fmt"

	"github.com/easterthebunny/spew-order/internal/key"
	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/pkg/types"
	"github.com/shopspring/decimal"
)

type BalanceRepository struct {
	kvstore persist.KVStore
	account *persist.Account
	symbol  types.Symbol
}

func NewBalanceRepository(kv persist.KVStore, a *persist.Account, s types.Symbol) *BalanceRepository {
	return &BalanceRepository{
		kvstore: kv,
		account: a,
		symbol:  s,
	}
}

func (b *BalanceRepository) GetBalance() (balance decimal.Decimal, err error) {

	k := balanceKey(*b.account, b.symbol)
	var byt []byte
	byt, err = b.kvstore.Get(k)
	if err != nil {
		if err == persist.ErrObjectNotExist {
			return balance, nil
		}
		return balance, err
	}

	err = json.Unmarshal(byt, &balance)
	return
}

func (b *BalanceRepository) UpdateBalance(bal decimal.Decimal) error {

	k := balanceKey(*b.account, b.symbol)
	val, err := json.Marshal(bal)
	if err != nil {
		return err
	}

	attrs := persist.KVStoreObjectAttrsToUpdate{
		ContentEncoding: persist.JSONEncodingTypeName,
		Metadata:        make(map[string]string),
	}

	return b.kvstore.Set(k, val, &attrs)
}

func (b *BalanceRepository) FindHolds() (holds []*persist.BalanceItem, err error) {

	q := persist.KVStoreQuery{
		StartOffset: holdSubspace(*b.account, b.symbol).Pack(key.Tuple{}).String(),
	}

	attr, err := b.kvstore.RangeGet(&q, 0)
	if err != nil {
		return
	}

	for _, at := range attr {
		var bts []byte
		bts, err = b.kvstore.Get(at.Name)
		if err != nil {
			err = fmt.Errorf("Balace::FindHolds -- %w", err)
			return
		}

		bal := &persist.BalanceItem{}
		err = bal.Decode(bts, encodingFromStr(at.ContentEncoding))
		if err != nil {
			return
		}

		holds = append(holds, bal)
	}

	return
}

func (b *BalanceRepository) CreateHold(hold *persist.BalanceItem) error {
	if hold == nil {
		return fmt.Errorf("%w for hold", persist.ErrCannotSaveNilValue)
	}

	k := holdKey(*b.account, b.symbol, stringer(hold.ID))

	enc := persist.JSON
	bts, err := hold.Encode(enc)
	if err != nil {
		return err
	}

	attrs := persist.KVStoreObjectAttrsToUpdate{
		ContentEncoding: encodingToStr(enc),
		Metadata:        make(map[string]string),
	}

	return b.kvstore.Set(k, bts, &attrs)
}

func (b *BalanceRepository) UpdateHold(id persist.Key, amt decimal.Decimal) error {

	k := holdKey(*b.account, b.symbol, id)

	bts, err := b.kvstore.Get(k)
	if err != nil {
		err = fmt.Errorf("Balace::UpdateHold -- %w", err)
		return err
	}

	attrs, err := b.kvstore.Attrs(k)
	if err != nil {
		return err
	}

	item := &persist.BalanceItem{}
	enc := encodingFromStr(attrs.ContentEncoding)
	err = item.Decode(bts, enc)
	if err != nil {
		return err
	}

	item.Amount = amt
	return b.CreateHold(item)
}

func (b *BalanceRepository) DeleteHold(id persist.Key) error {
	return b.kvstore.Delete(holdKey(*b.account, b.symbol, id))
}

func (b *BalanceRepository) FindPosts() (posts []*persist.BalanceItem, err error) {

	q := persist.KVStoreQuery{
		StartOffset: postSubspace(*b.account, b.symbol).Pack(key.Tuple{}).String(),
	}

	attr, err := b.kvstore.RangeGet(&q, 0)
	if err != nil {
		return
	}

	for _, at := range attr {
		var bts []byte
		bts, err = b.kvstore.Get(at.Name)
		if err != nil {
			err = fmt.Errorf("Balace::FindPosts -- %w", err)
			return
		}

		bal := &persist.BalanceItem{}
		err = bal.Decode(bts, encodingFromStr(at.ContentEncoding))
		if err != nil {
			return
		}

		posts = append(posts, bal)
	}

	return
}

func (b *BalanceRepository) CreatePost(post *persist.BalanceItem) error {
	if post == nil {
		return fmt.Errorf("%w for post", persist.ErrCannotSaveNilValue)
	}

	k := postKey(*b.account, b.symbol, stringer(post.ID))

	enc := persist.JSON
	bts, err := post.Encode(enc)
	if err != nil {
		return err
	}

	attrs := persist.KVStoreObjectAttrsToUpdate{
		ContentEncoding: encodingToStr(enc),
		Metadata:        make(map[string]string),
	}

	return b.kvstore.Set(k, bts, &attrs)
}

func (b *BalanceRepository) DeletePost(post *persist.BalanceItem) error {
	if post == nil {
		return fmt.Errorf("%w for post", persist.ErrCannotSaveNilValue)
	}

	return b.kvstore.Delete(postKey(*b.account, b.symbol, stringer(post.ID)))
}
