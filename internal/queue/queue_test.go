package queue

import (
	"context"
	"testing"
	"time"

	"github.com/easterthebunny/spew-order/internal/contexts"
	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/internal/persist/kv"
	"github.com/easterthebunny/spew-order/pkg/domain"
	"github.com/easterthebunny/spew-order/pkg/types"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestPublishOrderRequest(t *testing.T) {

	// set up the mocked pub sub and establish a subscription to the topic
	subscription := make(chan domain.OrderMessage)
	mps := NewMockPubSub()
	mps.Subscribe(OrderTopic, subscription)

	acct := domain.NewAccount()
	repo := kv.NewAccountRepository(persist.NewMockKVStore())
	err := repo.Save(&persist.Account{ID: acct.ID.String()})
	if err != nil {
		t.FailNow()
	}
	svc := domain.NewBalanceManager(repo)
	svc.PostToBalance(acct, types.SymbolBitcoin, decimal.NewFromFloat(2.0))

	// account is required in the context
	ctx := contexts.AttachAccountID(context.Background(), acct.ID.String())

	q := NewOrderQueue(mps, svc)

	// PublishOrder requires an account in the context
	t.Run("MissingAccount", func(t *testing.T) {
		or := types.OrderRequest{
			Base:   types.SymbolBitcoin,
			Target: types.SymbolEthereum,
			Action: types.ActionTypeBuy,
			Type: &types.LimitOrderType{
				Base:     types.SymbolBitcoin,
				Price:    decimal.NewFromFloat(0.25),
				Quantity: decimal.NewFromFloat(4.0)}}

		_, err := q.PublishOrderRequest(context.Background(), or)
		assert.Error(t, contexts.ErrAccountNotFoundInContext, err, "account must exist in context")

		// nothing should be sent to pubsub
		select {
		case <-time.After(100 * time.Millisecond):
			return
		case <-subscription:
			t.Errorf("data found on the queue subscription")
		}

	})

	// PublishOrder should place hold on account, publish to pubsub, and return an id
	t.Run("Success", func(t *testing.T) {
		or := types.OrderRequest{
			Base:   types.SymbolBitcoin,
			Target: types.SymbolEthereum,
			Action: types.ActionTypeBuy,
			Type: &types.LimitOrderType{
				Base:     types.SymbolBitcoin,
				Price:    decimal.NewFromFloat(0.25),
				Quantity: decimal.NewFromFloat(4.0)}}

		id, err := q.PublishOrderRequest(ctx, or)
		assert.NoError(t, err)
		assert.NotEqual(t, "", id, "id returned is not blank")

		// the new order should be published to the order queue within the
		// handler. wait for the posting and fail if not found
		select {
		case <-time.After(100 * time.Millisecond):
			t.Errorf("no data found on the queue subscription")
		case <-subscription:
			// happy case
			return
		}

	})

	// PublishOrder should place hold on account, generate an error, and release the hold
	// the hold from the previous run should still be active and register this run as an error
	t.Run("InsuffientFunds", func(t *testing.T) {
		or := types.OrderRequest{
			Base:   types.SymbolBitcoin,
			Target: types.SymbolEthereum,
			Action: types.ActionTypeBuy,
			Type: &types.LimitOrderType{
				Base:     types.SymbolBitcoin,
				Price:    decimal.NewFromFloat(0.25),
				Quantity: decimal.NewFromFloat(5.0)}}

		_, err := q.PublishOrderRequest(ctx, or)
		assert.Error(t, domain.ErrInsufficientBalanceForHold, err, "must produce an error for insufficient funds")

		// nothing should be sent to pubsub
		select {
		case <-time.After(100 * time.Millisecond):
			return
		case <-subscription:
			t.Errorf("data found on the queue subscription")
		}

	})
}
