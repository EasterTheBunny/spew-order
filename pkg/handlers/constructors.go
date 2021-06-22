package handlers

import (
	"io"

	"github.com/easterthebunny/spew-order/internal/funding"
	"github.com/easterthebunny/spew-order/internal/middleware"
	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/internal/persist/kv"
	"github.com/easterthebunny/spew-order/internal/queue"
	"github.com/easterthebunny/spew-order/pkg/domain"
)

func NewGoogleOrderBook(kvstore persist.KVStore, f funding.Source) *domain.OrderBook {
	br := kv.NewBookRepository(kvstore)
	a := kv.NewAccountRepository(kvstore)
	l := kv.NewLedgerRepository(kvstore)
	bs := domain.NewBalanceManager(a, l, f)
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

func NewFundingSource(t string, apiKey, apiSecret *string, audit io.Writer, pubkey io.Reader) funding.Source {
	switch t {
	case "COINBASE":
		if apiKey == nil || apiSecret == nil {
			return nil
		}

		if *apiKey == "" || *apiSecret == "" {
			return nil
		}

		return funding.NewCoinbaseSource(funding.SourceConfig{
			CallbackAudit: audit,
			PublicKey:     pubkey,
			APIKey:        *apiKey,
			APISecret:     *apiSecret,
		})
	default:
		return funding.NewMockSource()
	}
}

func NewDefaultRouter(kvstore persist.KVStore, ps queue.PubSub, pr middleware.AuthenticationProvider, f funding.Source) (*Router, error) {
	a := kv.NewAccountRepository(kvstore)
	l := kv.NewLedgerRepository(kvstore)
	bs := domain.NewBalanceManager(a, l, f)

	r := Router{
		AuthStore: kv.NewAuthorizationRepository(kvstore),
		Balance:   bs,
		AuthProv:  pr,
		Orders:    NewOrderHandler(queue.NewOrderQueue(ps, bs)),
		Accounts:  NewAccountHandler(a),
	}

	return &r, nil
}

func NewWebhookRouter(kvstore persist.KVStore, f funding.Source) *WebhookRouter {
	a := kv.NewAccountRepository(kvstore)
	l := kv.NewLedgerRepository(kvstore)

	return &WebhookRouter{Funding: NewFundingHandler(a, l, f)}
}

func NewAuditRouter(kvstore persist.KVStore) *AuditRouter {
	a := kv.NewAccountRepository(kvstore)
	u := kv.NewAuthorizationRepository(kvstore)
	l := kv.NewLedgerRepository(kvstore)

	return &AuditRouter{Audit: NewAuditHandler(a, u, l)}
}
