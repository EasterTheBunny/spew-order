package account

import (
	"encoding/json"
	"fmt"

	"github.com/easterthebunny/spew-order/internal/key"
	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/pkg/types"
	uuid "github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
)

type KVAccountRepository struct {
	kvstore persist.KVStore
}

func NewKVAccountRepository(store persist.KVStore) *KVAccountRepository {
	return &KVAccountRepository{kvstore: store}
}

var _ AccountRepository = &KVAccountRepository{}
var _ BalanceRepository = &balanceRepository{}

const (
	bookSub int = iota
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
)

func (r *KVAccountRepository) Find(id uuid.UUID) (*types.Account, error) {

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

func (r *KVAccountRepository) Save(a *types.Account) error {
	if a == nil {
		return fmt.Errorf("no action available for nil value")
	}

	key := gsAccount.Pack(key.Tuple{a.ID.String()})

	b, err := json.Marshal(*a)
	if err != nil {
		return err
	}

	attrs := persist.KVStoreObjectAttrsToUpdate{
		Metadata: make(map[string]string),
	}

	r.kvstore.Set(key.String(), b, &attrs)

	return nil
}

func (r *KVAccountRepository) Balances(a *types.Account, s types.Symbol) BalanceRepository {
	return &balanceRepository{
		kvstore: r.kvstore,
		account: a,
		symbol:  s,
	}
}

type balanceRepository struct {
	kvstore persist.KVStore
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
		if err == persist.ErrObjectNotExist {
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

	attrs := persist.KVStoreObjectAttrsToUpdate{
		Metadata: make(map[string]string),
	}

	return b.kvstore.Set(key.String(), val, &attrs)
}

func (b *balanceRepository) FindHolds() (holds []*BalanceItem, err error) {

	s := gsAccount.Sub(b.account.ID.String())
	s = s.Sub(symbolsSub).Sub(b.symbol.String())
	key := s.Pack(key.Tuple{holdSub})

	q := persist.KVStoreQuery{
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

		var bal BalanceItem
		err = json.Unmarshal(bts, &bal)
		if err != nil {
			return
		}

		holds = append(holds, &bal)
	}

	return
}

func (b *balanceRepository) CreateHold(hold *BalanceItem) error {
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

	attrs := persist.KVStoreObjectAttrsToUpdate{
		Metadata: make(map[string]string),
	}

	return b.kvstore.Set(key.String(), bts, &attrs)
}

func (b *balanceRepository) DeleteHold(hold *BalanceItem) error {
	if hold == nil {
		return fmt.Errorf("no action available for nil value")
	}

	s := gsAccount.Sub(b.account.ID.String())
	s = s.Sub(symbolsSub).Sub(b.symbol.String())
	s = s.Sub(holdSub)
	key := s.Pack(key.Tuple{hold.Timestamp.UnixNano(), hold.ID.String()})

	return b.kvstore.Delete(key.String())
}

func (b *balanceRepository) FindPosts() (posts []*BalanceItem, err error) {

	s := gsAccount.Sub(b.account.ID.String())
	s = s.Sub(symbolsSub).Sub(b.symbol.String())
	key := s.Pack(key.Tuple{postSub})

	q := persist.KVStoreQuery{
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

		var bal BalanceItem
		err = json.Unmarshal(bts, &bal)
		if err != nil {
			return
		}

		posts = append(posts, &bal)
	}

	return
}

func (b *balanceRepository) CreatePost(post *BalanceItem) error {
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

	attrs := persist.KVStoreObjectAttrsToUpdate{
		Metadata: make(map[string]string),
	}

	return b.kvstore.Set(key.String(), bts, &attrs)
}

func (b *balanceRepository) DeletePost(post *BalanceItem) error {
	if post == nil {
		return fmt.Errorf("no action available for nil value")
	}

	s := gsAccount.Sub(b.account.ID.String())
	s = s.Sub(symbolsSub).Sub(b.symbol.String())
	s = s.Sub(postSub)
	key := s.Pack(key.Tuple{post.Timestamp.UnixNano(), post.ID.String()})

	return b.kvstore.Delete(key.String())
}
