package kv

import (
	"github.com/easterthebunny/spew-order/internal/key"
	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/pkg/types"
)

const (
	bookSub int = iota
	authzSub
	accountSub
	symbolsSub
	balanceSub
	holdSub
	postSub
	orderSub
	transactionSub
	ledgerSub
	addressSub
)

var (
	gsRoot = key.FromBytes([]byte{0xFE})
)

var _ persist.AccountRepository = &AccountRepository{}
var _ persist.BookRepository = &BookRepository{}
var _ persist.BalanceRepository = &BalanceRepository{}
var _ persist.AuthorizationRepository = &AuthorizationRepository{}
var _ persist.TransactionRepository = &TransactionRepository{}
var _ persist.OrderRepository = &OrderRepository{}

func ledgerSubspace() key.Subspace {
	// /root/ledger
	return gsRoot.Sub(ledgerSub)
}

func addressKey(addr string, sym types.Symbol) string {
	// /root/addresses/{symbol}/{addr}
	return gsRoot.Sub(accountSub).
		Sub(addressSub).
		Sub(sym.String()).
		Pack(key.Tuple{addr}).String()
}

func accountSubspace(acct *persist.Account) key.Subspace {
	// /root/account/{accountid}
	s := gsRoot.Sub(accountSub)
	if acct != nil {
		return s.Sub(acct.ID)
	}
	return s
}

func accountKey(id persist.Key) string {
	// /root/account/{accountid}
	return accountSubspace(nil).Pack(key.Tuple{id.String()}).String()
}

func authzSubspace() key.Subspace {
	// /root/authz
	return gsRoot.Sub(authzSub)
}

func authzKey(id persist.Key) string {
	// /root/authz/{authzid}
	return authzSubspace().Pack(key.Tuple{id.String()}).String()
}

func transactionSubspace(acct persist.Account) key.Subspace {
	// /root/account/{accountid}/transaction
	return accountSubspace(&acct).Sub(transactionSub)
}

func transactionKey(acct persist.Account, t persist.Transaction) string {
	// /root/account/{accountid}/transaction/{timestamp}
	return transactionSubspace(acct).
		Pack(key.Tuple{t.Timestamp.Value()}).String()
}

func orderSubspace(acct persist.Account) key.Subspace {
	// /root/account/{accountid}/order
	return accountSubspace(&acct).
		Sub(orderSub)
}

func orderKey(acct persist.Account, t persist.Order) string {
	// /root/account/{accountid}/order/{orderid}
	return orderSubspace(acct).
		Pack(key.Tuple{t.Base.ID.String()}).String()
}

func orderIDKey(acct persist.Account, k persist.Key) string {
	// /root/account/{accountid}/order/{orderid}
	return accountSubspace(&acct).
		Sub(orderSub).
		Pack(key.Tuple{k.String()}).String()
}

func balanceKey(acct persist.Account, sym types.Symbol) string {
	// /root/account/{accountid}/symbol/{symbol}/balance
	return accountSubspace(&acct).
		Sub(symbolsSub).
		Sub(sym.String()).
		Pack(key.Tuple{balanceSub}).String()
}

func holdSubspace(acct persist.Account, sym types.Symbol) key.Subspace {
	// /root/account/{accountid}/symbol/{symbol}/hold
	return accountSubspace(&acct).
		Sub(symbolsSub).
		Sub(sym.String()).
		Sub(holdSub)
}

func holdKey(acct persist.Account, sym types.Symbol, hold persist.Key) string {
	// /root/account/{accountid}/symbol/{symbol}/hold/{holdid}
	return holdSubspace(acct, sym).Pack(key.Tuple{hold.String()}).String()
}

func postSubspace(acct persist.Account, sym types.Symbol) key.Subspace {
	// /root/account/{accountid}/symbol/{symbol}/post
	return accountSubspace(&acct).
		Sub(symbolsSub).
		Sub(sym.String()).
		Sub(postSub)
}

func postKey(acct persist.Account, sym types.Symbol, post persist.Key) string {
	// /root/account/{accountid}/symbol/{symbol}/post/{postid}
	return postSubspace(acct, sym).Pack(key.Tuple{post.String()}).String()
}

func bookItemSubspace(b persist.BookItem, t *types.ActionType) key.Subspace {
	// /root/book/{base}/{target}/{BUY|SELL}
	tp := b.Order.Action
	if t != nil {
		tp = *t
	}

	return gsRoot.Sub(bookSub).
		Sub(uint(b.Order.Base)).
		Sub(uint(b.Order.Target)).
		Sub(uint(tp))
}

// bookItemKey generates a key that will sort ASC lexicographically, but remain in
// type sorted order: buys are sorted largest/oldest to smallest/newest and sells
// are sorted smallest/oldest to largest/newest
func bookItemKey(b persist.BookItem) string {
	// /root/book/{base}/{target}/{BUY|SELL}/{decimal_price}{timestamp}
	p := b.Order.Type.KeyTuple(b.Order.Action)
	p = append(p, key.Tuple{b.Order.Timestamp.UnixNano()}...)
	return bookItemSubspace(b, nil).Pack(p).String()
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

type stringer string

func (s stringer) String() string {
	return string(s)
}
