package persist

import (
	"encoding/json"
	"fmt"

	"github.com/easterthebunny/spew-order/internal/key"
	"github.com/easterthebunny/spew-order/pkg/types"
	uuid "github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
)

type AccountRepository struct {
	kvstore KVStore
}

func NewAccountRepository(store KVStore) *AccountRepository {
	return &AccountRepository{kvstore: store}
}

func (r *AccountRepository) Find(id uuid.UUID) (*types.Account, error) {

	key := gsAccount.Pack(key.Tuple{id.String()})
	b, err := r.kvstore.Get(key.String())
	if err != nil {
		return nil, err
	}

	var account types.Account
	err = json.Unmarshal(b, &account)
	if err != nil {
		return nil, err
	}

	return &account, nil
}

func (r *AccountRepository) Save(a *types.Account) error {
	if a == nil {
		return fmt.Errorf("no action available for nil value")
	}

	key := gsAccount.Pack(key.Tuple{a.ID.String()})

	b, err := json.Marshal(*a)
	if err != nil {
		return err
	}

	attrs := KVStoreObjectAttrsToUpdate{
		Metadata: make(map[string]string),
	}

	r.kvstore.Set(key.String(), b, &attrs)

	return nil
}

func (r *AccountRepository) Balances(a *types.Account, s types.Symbol) types.BalanceRepo {
	return &balanceRepository{
		kvstore: r.kvstore,
		account: a,
		symbol:  s,
	}
}

type balanceRepository struct {
	kvstore KVStore
	account *types.Account
	symbol  types.Symbol
}

func (b *balanceRepository) GetBalance() (decimal.Decimal, error) {

	var balance decimal.Decimal

	s := gsAccount.Sub(b.account.ID.String())
	s = s.Sub(symbolsSub).Sub(b.symbol.String())
	key := s.Pack(key.Tuple{balanceSub})

	byt, err := b.kvstore.Get(key.String())
	if err != nil {
		if err == ErrObjectNotExist {
			return balance, nil
		}
		return balance, err
	}

	err = json.Unmarshal(byt, &balance)
	return balance, err
}

func (b *balanceRepository) UpdateBalance(bal decimal.Decimal) error {

	s := gsAccount.Sub(b.account.ID.String())
	s = s.Sub(symbolsSub).Sub(b.symbol.String())
	key := s.Pack(key.Tuple{balanceSub})

	val, err := json.Marshal(bal)
	if err != nil {
		return err
	}

	attrs := KVStoreObjectAttrsToUpdate{
		Metadata: make(map[string]string),
	}

	return b.kvstore.Set(key.String(), val, &attrs)
}

func (b *balanceRepository) FindHolds() (holds []*types.BalanceItem, err error) {

	s := gsAccount.Sub(b.account.ID.String())
	s = s.Sub(symbolsSub).Sub(b.symbol.String())
	key := s.Pack(key.Tuple{holdSub})

	q := KVStoreQuery{
		StartOffset: key.String(),
	}

	attr, err := b.kvstore.RangeGet(&q, 0)
	if err != nil {
		return
	}

	for _, at := range attr {
		var bts []byte
		bts, err = b.kvstore.Get(at.Name)
		if err != nil {
			return
		}

		var bal types.BalanceItem
		err = json.Unmarshal(bts, &bal)
		if err != nil {
			return
		}

		holds = append(holds, &bal)
	}

	return
}

func (b *balanceRepository) CreateHold(hold *types.BalanceItem) error {
	if hold == nil {
		return fmt.Errorf("no action available for nil value")
	}

	s := gsAccount.Sub(b.account.ID.String())
	s = s.Sub(symbolsSub).Sub(b.symbol.String())
	s = s.Sub(holdSub)
	key := s.Pack(key.Tuple{hold.Timestamp.UnixNano(), hold.ID.String()})

	bts, err := json.Marshal(hold)
	if err != nil {
		return err
	}

	attrs := KVStoreObjectAttrsToUpdate{
		Metadata: make(map[string]string),
	}

	return b.kvstore.Set(key.String(), bts, &attrs)
}

func (b *balanceRepository) DeleteHold(hold *types.BalanceItem) error {
	if hold == nil {
		return fmt.Errorf("no action available for nil value")
	}

	s := gsAccount.Sub(b.account.ID.String())
	s = s.Sub(symbolsSub).Sub(b.symbol.String())
	s = s.Sub(holdSub)
	key := s.Pack(key.Tuple{hold.Timestamp.UnixNano(), hold.ID.String()})

	return b.kvstore.Delete(key.String())
}

/*
	/account/{accountid}/symbols/{symbol}/post/{timestamp_nano}{orderid}
*/
func (b *balanceRepository) FindPosts() (posts []*types.BalanceItem, err error) {

	s := gsAccount.Sub(b.account.ID.String())
	s = s.Sub(symbolsSub).Sub(b.symbol.String())
	key := s.Pack(key.Tuple{postSub})

	q := KVStoreQuery{
		StartOffset: key.String(),
	}

	attr, err := b.kvstore.RangeGet(&q, 0)
	if err != nil {
		return
	}

	for _, at := range attr {
		var bts []byte
		bts, err = b.kvstore.Get(at.Name)
		if err != nil {
			return
		}

		var bal types.BalanceItem
		err = json.Unmarshal(bts, &bal)
		if err != nil {
			return
		}

		posts = append(posts, &bal)
	}

	return
}

func (b *balanceRepository) CreatePost(post *types.BalanceItem) error {
	if post == nil {
		return fmt.Errorf("no action available for nil value")
	}

	s := gsAccount.Sub(b.account.ID.String())
	s = s.Sub(symbolsSub).Sub(b.symbol.String())
	s = s.Sub(postSub)
	key := s.Pack(key.Tuple{post.Timestamp.UnixNano(), post.ID.String()})

	bts, err := json.Marshal(post)
	if err != nil {
		return err
	}

	attrs := KVStoreObjectAttrsToUpdate{
		Metadata: make(map[string]string),
	}

	return b.kvstore.Set(key.String(), bts, &attrs)
}

func (b *balanceRepository) DeletePost(post *types.BalanceItem) error {
	if post == nil {
		return fmt.Errorf("no action available for nil value")
	}

	s := gsAccount.Sub(b.account.ID.String())
	s = s.Sub(symbolsSub).Sub(b.symbol.String())
	s = s.Sub(postSub)
	key := s.Pack(key.Tuple{post.Timestamp.UnixNano(), post.ID.String()})

	return b.kvstore.Delete(key.String())
}
