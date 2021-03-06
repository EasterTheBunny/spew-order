package kv

import (
	"context"
	"encoding/json"
	"errors"
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

func (b *BalanceRepository) GetBalance(ctx context.Context) (balance decimal.Decimal, err error) {

	k := balanceKey(*b.account, b.symbol)
	var byt []byte
	b.kvstore.Attrs(k)
	byt, err = b.kvstore.Get(k)
	if err != nil {
		if errors.Is(err, persist.ErrObjectNotExist) {
			return balance, nil
		}
		return balance, err
	}

	err = json.Unmarshal(byt, &balance)
	return
}

func (b *BalanceRepository) AddToBalance(ctx context.Context, amt decimal.Decimal) error {
	bal, _ := b.GetBalance(ctx)
	bal = bal.Add(amt)
	return b.UpdateBalance(ctx, bal)
}

func (b *BalanceRepository) UpdateBalance(ctx context.Context, bal decimal.Decimal) error {

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

func (b *BalanceRepository) FindHolds(ctx context.Context) (holds []*persist.BalanceItem, err error) {

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
			err = fmt.Errorf("Balance::FindHolds -- %w", err)
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

// CreateHold stores a new hold in a time sorted list
func (b *BalanceRepository) CreateHold(ctx context.Context, hold *persist.BalanceItem) error {
	if hold == nil {
		return fmt.Errorf("%w for hold", persist.ErrCannotSaveNilValue)
	}

	k := holdKey(*b.account, b.symbol, hold.Timestamp)

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

func (b *BalanceRepository) UpdateHold(ctx context.Context, id persist.Key, amt decimal.Decimal) error {

	holds, err := b.FindHolds(ctx)
	if err != nil {
		return err
	}

	for _, hold := range holds {
		if hold.ID == id.String() {
			hold.Amount = amt
			return b.CreateHold(ctx, hold)
		}
	}

	return persist.ErrObjectNotExist
}

func (b *BalanceRepository) DeleteHold(ctx context.Context, id persist.Key) error {

	holds, err := b.FindHolds(ctx)
	if err != nil {
		return err
	}

	for _, hold := range holds {
		if hold.ID == id.String() {
			return b.kvstore.Delete(holdKey(*b.account, b.symbol, hold.Timestamp))
		}
	}

	return persist.ErrObjectNotExist
}

func (b *BalanceRepository) FindPosts(ctx context.Context) (posts []*persist.BalanceItem, err error) {

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

func (b *BalanceRepository) CreatePost(ctx context.Context, post *persist.BalanceItem) error {
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

func (b *BalanceRepository) DeletePost(ctx context.Context, post *persist.BalanceItem) error {
	if post == nil {
		return fmt.Errorf("%w for post", persist.ErrCannotSaveNilValue)
	}

	return b.kvstore.Delete(postKey(*b.account, b.symbol, stringer(post.ID)))
}
