package queue

import (
	"context"
	"encoding/json"

	"github.com/easterthebunny/spew-order/internal/contexts"
	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/internal/persist/kv"
	"github.com/easterthebunny/spew-order/pkg/domain"
	"github.com/easterthebunny/spew-order/pkg/types"
)

var (
	OrderTopic = "OrderRequests"
)

func NewGoogleOrderQueue(projectID string, bucket string) (*OrderQueue, error) {
	q := NewGooglePubSub(projectID)
	s, err := persist.NewGoogleKVStore(&bucket)
	if err != nil {
		return nil, err
	}
	r := kv.NewAccountRepository(s)
	bs := domain.NewBalanceManager(r)

	oq := OrderQueue{
		client:  q,
		balance: bs}

	return &oq, nil
}

func NewOrderQueue(pubsub PubSub, bs *domain.BalanceManager) *OrderQueue {
	return &OrderQueue{
		client:  pubsub,
		balance: bs}
}

type OrderQueue struct {
	client  PubSub
	balance *domain.BalanceManager
}

func (o *OrderQueue) PublishOrderRequest(ctx context.Context, or types.OrderRequest) (id string, err error) {

	b, err := json.Marshal(or)
	if err != nil {
		return
	}

	aID, err := contexts.GetAccountID(ctx)
	if err != nil {
		return
	}

	acct, err := o.balance.GetAccount(aID)
	if err != nil {
		return
	}

	if acct == nil {
		err = contexts.ErrAccountNotFoundInContext
		return
	}

	// place hold on account
	symbol, hold := or.Type.HoldAmount(or.Action, or.Base, or.Target)
	holdid, err := o.balance.SetHoldOnAccount(acct, symbol, hold)
	if err != nil {
		return
	}

	or.HoldID = holdid

	return o.client.Publish(ctx, OrderTopic, b)
}
