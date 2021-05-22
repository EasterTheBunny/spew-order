package handlers

import (
	"github.com/easterthebunny/spew-order/internal/middleware"
	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/internal/persist/kv"
	"github.com/easterthebunny/spew-order/internal/queue"
	"github.com/easterthebunny/spew-order/pkg/domain"
)

func NewGoogleOrderBook(kvstore persist.KVStore) *domain.OrderBook {
	br := kv.NewBookRepository(kvstore)
	a := kv.NewAccountRepository(kvstore)
	l := kv.NewLedgerRepository(kvstore)
	bs := domain.NewBalanceManager(a, l)
	return domain.NewOrderBook(br, bs)
}

func NewGoogleKVStore(bucket *string) (persist.KVStore, error) {
	return persist.NewGoogleKVStore(bucket)
}

func NewGooglePubSub(projectID string) queue.PubSub {
	return queue.NewGooglePubSub(projectID)
}

func NewJWTAuth(url string) (middleware.AuthenticationProvider, error) {
	return middleware.NewJWTAuth(url)
}

func NewDefaultRouter(kvstore persist.KVStore, ps queue.PubSub, pr middleware.AuthenticationProvider) (*Router, error) {
	a := kv.NewAccountRepository(kvstore)
	l := kv.NewLedgerRepository(kvstore)
	bs := domain.NewBalanceManager(a, l)

	r := Router{
		AuthStore: kv.NewAuthorizationRepository(kvstore),
		Balance:   bs,
		AuthProv:  pr,
		Orders:    NewOrderHandler(queue.NewOrderQueue(ps, bs)),
		Accounts:  NewAccountHandler(a),
	}

	return &r, nil
}

func NewWebhookRouter(kvstore persist.KVStore) *WebhookRouter {
	a := kv.NewAccountRepository(kvstore)
	l := kv.NewLedgerRepository(kvstore)
	return &WebhookRouter{Funding: NewFundingHandler(a, l)}
}
