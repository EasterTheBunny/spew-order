package queue

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/easterthebunny/spew-order/internal/contexts"
	"github.com/easterthebunny/spew-order/pkg/domain"
	"github.com/easterthebunny/spew-order/pkg/types"
	"github.com/shopspring/decimal"
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

func (o *OrderQueue) CancelOrder(ctx context.Context, order types.Order) (err error) {

	om := domain.OrderMessage{
		Action: domain.CancelOrderMessageType,
		Order:  order}

	b, err := json.Marshal(om)
	if err != nil {
		return
	}

	_, err = o.client.Publish(ctx, OrderTopic, b)

	return
}

func (o *OrderQueue) PublishOrderRequest(ctx context.Context, or types.OrderRequest) (order types.Order, err error) {

	aID, err := contexts.GetAccountID(ctx)
	if err != nil {
		return
	}

	acct, err := o.balance.GetAccount(ctx, aID)
	if err != nil {
		return
	}

	if acct == nil {
		err = contexts.ErrAccountNotFoundInContext
		return
	}

	// place hold on account
	symbol, hold := or.Type.HoldAmount(or.Action, or.Base, or.Target)
	if hold.LessThanOrEqual(decimal.NewFromInt(0)) {
		err = errors.New("order type not supported")
		return
	}

	// set a hold on the traded amount
	holdid, err := o.balance.SetHoldOnAccount(ctx, acct, symbol, hold)
	if err != nil {
		return
	}

	or.HoldID = holdid

	// set a hold on the fee amount if not dealing with native token
	if or.Target != types.SymbolCipherMtn {
		var feeHoldId string
		feeHoldId, err = o.balance.SetHoldOnAccount(ctx, acct, types.SymbolCipherMtn, types.StandardFee)
		if err != nil {
			return
		}

		or.FeeHoldID = feeHoldId
	}

	order, err = o.balance.CreateOrder(ctx, acct, or)
	if err != nil {
		return
	}

	om := domain.OrderMessage{
		Action: domain.OpenOrderMessageType,
		Order:  order}

	b, err := json.Marshal(om)
	if err != nil {
		return
	}

	_, err = o.client.Publish(ctx, OrderTopic, b)

	return
}
