package queue

import (
	"context"
	"encoding/json"

	"github.com/easterthebunny/spew-order/internal/contexts"
	"github.com/easterthebunny/spew-order/pkg/domain"
	"github.com/easterthebunny/spew-order/pkg/types"
)

var (
	OrderTopic = "OrderRequests"
)

func NewGoogleOrderQueue(projectID string, manager *domain.BalanceManager) (*OrderQueue, error) {
	q := NewGooglePubSub(projectID)
	oq := OrderQueue{
		client:  q,
		balance: manager}

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

func (o *OrderQueue) PublishOrderRequest(ctx context.Context, or types.OrderRequest) (order types.Order, err error) {

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

	order, err = o.balance.CreateOrder(acct, or)
	if err != nil {
		return
	}

	b, err := json.Marshal(order)
	if err != nil {
		return
	}

	_, err = o.client.Publish(ctx, OrderTopic, b)

	return
}
