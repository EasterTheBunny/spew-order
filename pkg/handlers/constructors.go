package handlers

import (
	"io"

	"cloud.google.com/go/firestore"
	"github.com/easterthebunny/spew-order/internal/funding"
	"github.com/easterthebunny/spew-order/internal/middleware"
	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/internal/persist/firebase"
	"github.com/easterthebunny/spew-order/internal/queue"
	"github.com/easterthebunny/spew-order/pkg/domain"
)

func NewGoogleOrderBook(client *firestore.Client, f ...funding.Source) *domain.OrderBook {
	br := firebase.NewBookRepository(client)
	a := firebase.NewAccountRepository(client)
	l := firebase.NewLedgerRepository(client)
	bs := domain.NewBalanceManager(a, l, f...)
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
	case "CMTN":
		return funding.NewAirdropSource(*apiKey)
	default:
		return funding.NewMockSource()
	}
}

func NewDefaultRouter(client *firestore.Client, ps queue.PubSub, pr middleware.AuthenticationProvider, f ...funding.Source) (*Router, error) {
	a := firebase.NewAccountRepository(client)
	l := firebase.NewLedgerRepository(client)
	bs := domain.NewBalanceManager(a, l, f...)

	r := Router{
		AuthStore: firebase.NewAuthorizationRepository(client),
		Balance:   bs,
		AuthProv:  pr,
		Orders:    NewOrderHandler(queue.NewOrderQueue(ps, bs)),
		Accounts:  NewAccountHandler(a),
	}

	return &r, nil
}

func NewWebhookRouter(client *firestore.Client, f funding.Source, d funding.Source) *WebhookRouter {
	a := firebase.NewAccountRepository(client)
	l := firebase.NewLedgerRepository(client)

	return &WebhookRouter{
		Funding: NewFundingHandler(a, l, f),
		Airdrop: NewFundingHandler(a, l, d)}
}

func NewAuditRouter(client *firestore.Client) *AuditRouter {
	a := firebase.NewAccountRepository(client)
	u := firebase.NewAuthorizationRepository(client)
	l := firebase.NewLedgerRepository(client)
	b := firebase.NewBookRepository(client)

	return &AuditRouter{Audit: NewAuditHandler(a, u, l, b)}
}
