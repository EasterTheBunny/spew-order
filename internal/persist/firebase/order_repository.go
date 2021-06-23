package firebase

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"github.com/easterthebunny/spew-order/internal/persist"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OrderRepository struct {
	client  *firestore.Client
	account *persist.Account
}

func NewOrderRepository(client *firestore.Client, account *persist.Account) *OrderRepository {
	return &OrderRepository{client: client, account: account}
}

// /root/account/{accountid}/order/{orderid}
func (or *OrderRepository) GetOrder(ctx context.Context, k persist.Key) (*persist.Order, error) {
	col := fmt.Sprintf("accounts/%s/orders", or.account.ID)
	dsnap, err := or.getClient(ctx).Collection(col).Doc(k.String()).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, persist.ErrObjectNotExist
		}
		return nil, err
	}

	var od persist.Order
	err = dsnap.DataTo(&od)
	if err != nil {
		return nil, err
	}

	return &od, nil
}

func (or *OrderRepository) SetOrder(ctx context.Context, o *persist.Order) error {
	col := fmt.Sprintf("accounts/%s/orders", or.account.ID)
	_, err := or.getClient(ctx).Collection(col).Doc(o.Base.ID.String()).Set(ctx, *o)
	if err != nil {
		return err
	}

	return nil
}

func (or *OrderRepository) GetOrdersByStatus(ctx context.Context, s ...persist.FillStatus) (orders []*persist.Order, err error) {
	col := fmt.Sprintf("accounts/%s/orders", or.account.ID)
	iter := or.getClient(ctx).Collection(col).Where("status", "in", s).Documents(ctx)
	var doc *firestore.DocumentSnapshot
	for {
		doc, err = iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return
		}

		var od persist.Order
		err = doc.DataTo(&od)
		if err != nil {
			return
		}
		orders = append(orders, &od)
	}
	return
}

func (or *OrderRepository) UpdateOrderStatus(ctx context.Context, k persist.Key, s persist.FillStatus, tr []string) error {
	order, err := or.GetOrder(ctx, k)
	if err != nil {
		return err
	}

	order.Status = s
	if len(tr) > 0 {
		order.Transactions = append(order.Transactions, tr)
	}
	return or.SetOrder(ctx, order)
}

func (or *OrderRepository) getClient(ctx context.Context) *firestore.Client {

	var client *firestore.Client
	if or.client == nil {
		client = clientFromContext(ctx)
	} else {
		client = or.client
	}
	return client
}
