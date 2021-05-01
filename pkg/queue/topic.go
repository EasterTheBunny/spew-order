package queue

import (
	"context"
	"encoding/json"

	"github.com/easterthebunny/spew-order/internal/queue"
	"github.com/easterthebunny/spew-order/pkg/types"
)

var (
	OrderTopic = "OrderRequests"
)

type OrderQueue interface {
	PublishOrderRequest(context.Context, types.OrderRequest) (string, error)
}

func NewGoogleOrderQueue(projectID string) OrderQueue {
	return &orderPubSub{
		client: queue.NewGooglePubSub(projectID)}
}

type orderPubSub struct {
	client queue.PubSub
}

func (o *orderPubSub) PublishOrderRequest(ctx context.Context, or types.OrderRequest) (id string, err error) {

	b, err := json.Marshal(or)
	if err != nil {
		return
	}

	return o.client.Publish(ctx, OrderTopic, b)
}
