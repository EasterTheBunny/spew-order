package handlers

import (
	"github.com/easterthebunny/spew-order/internal/account"
	"github.com/easterthebunny/spew-order/internal/auth"
	"github.com/easterthebunny/spew-order/internal/middleware"
	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/internal/queue"
	"github.com/easterthebunny/spew-order/pkg/book"
)

func NewGoogleOrderBook(kv persist.KVStore) book.OrderBook {
	return account.NewKVBookRepository(kv)
}

func NewGoogleKVStore(bucket *string) (persist.KVStore, error) {
	return persist.NewGoogleKVStore(bucket)
}

func NewGooglePubSub(projectID string) queue.PubSub {
	return queue.NewGooglePubSub(projectID)
}

func NewJWTAuth(url string) (auth.AuthenticationProvider, error) {
	return middleware.NewJWTAuth(url)
}

func NewDefaultRouter(kv persist.KVStore, ps queue.PubSub, pr auth.AuthenticationProvider) (*Router, error) {
	a := account.NewKVAccountRepository(kv)
	bs := account.NewBalanceService(a)

	r := Router{
		AuthStore: account.NewKVAuthzRepository(kv),
		Balance:   bs,
		AuthProv:  pr,
		Orders: &OrderHandler{
			queue: queue.NewOrderQueue(ps, bs),
		},
		Accounts: &AccountHandler{repo: a},
	}

	return &r, nil
}
