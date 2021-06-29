package firebase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"cloud.google.com/go/firestore"
	"github.com/easterthebunny/spew-order/internal/persist"
	"google.golang.org/api/iterator"
)

var (
	ErrOrderNotFound = errors.New("order not found")
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
	order, _, err := or.getOrder(ctx, k)
	if err != nil {
		err = fmt.Errorf("GetOrder: %w", err)
	}
	return order, err
}

func (or *OrderRepository) SetOrder(ctx context.Context, o *persist.Order) error {

	_, version, err := or.getOrder(ctx, o.Base.ID)
	if err != nil {
		if errors.Is(err, ErrOrderNotFound) {
			err = nil
			version = 0
		} else {
			return fmt.Errorf("SetOrder: %w", err)
		}
	}
	version++

	col := fmt.Sprintf("accounts/%s/orders", or.account.ID)
	_, _, err = or.getClient(ctx).Collection(col).Add(ctx, orderToDocument(o, 0))
	if err != nil {
		return fmt.Errorf("SetOrder: %w", err)
	}

	return nil
}

func (or *OrderRepository) GetOrdersByStatus(ctx context.Context, s ...persist.FillStatus) (orders []*persist.Order, err error) {

	ops := make([]string, len(s))
	for i, v := range s {
		ops[i] = v.String()
	}

	client := or.getClient(ctx)
	col := fmt.Sprintf("accounts/%s/orders", or.account.ID)
	iter := client.Collection(col).
		Where("status", "in", ops).
		OrderBy("version", firestore.Desc).
		Documents(ctx)

	versionMap := make(map[string]bool)

	batch := client.Batch()
	batchedItems := false
	var doc *firestore.DocumentSnapshot
	var order *persist.Order
	for {
		doc, err = iter.Next()
		if err != nil {
			if errors.Is(err, iterator.Done) {
				err = nil
			} else {
				err = fmt.Errorf("GetOrdersByStatus: %w", err)
			}

			break
		}

		order = documentToOrder(doc.Data())
		if _, ok := versionMap[order.Base.ID.String()]; !ok {
			versionMap[order.Base.ID.String()] = true
			orders = append(orders, order)
		} else if canChange(doc.UpdateTime) {
			// delete the older version
			batch.Delete(doc.Ref)
			batchedItems = true
		}
	}
	iter.Stop()

	if batchedItems {
		_, err = batch.Commit(ctx)
		if err != nil {
			err = fmt.Errorf("GetOrdersByStatus: %w", err)
		}
	}

	return
}

func (or *OrderRepository) UpdateOrderStatus(ctx context.Context, k persist.Key, s persist.FillStatus, tr []string) error {
	o, version, err := or.getOrder(ctx, k)
	if err != nil {
		return fmt.Errorf("UpdateOrderStatus: %w", err)
	}
	version++

	o.Status = s
	if len(tr) > 0 {
		o.Transactions = append(o.Transactions, tr)
	}

	col := fmt.Sprintf("accounts/%s/orders", or.account.ID)
	_, _, err = or.getClient(ctx).Collection(col).Add(ctx, orderToDocument(o, version))
	if err != nil {
		return fmt.Errorf("UpdateOrderStatus: %w", err)
	}

	return nil

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

func (or *OrderRepository) getOrder(ctx context.Context, k persist.Key) (*persist.Order, int, error) {
	var err error
	client := or.getClient(ctx)
	col := fmt.Sprintf("accounts/%s/orders", or.account.ID)
	iter := client.Collection(col).
		Where("id", "==", k.String()).
		OrderBy("version", firestore.Desc).
		Documents(ctx)

	batch := client.Batch()
	batchedItems := false
	var version int = 0
	var order *persist.Order
	var doc *firestore.DocumentSnapshot
	for {
		doc, err = iter.Next()
		if err != nil {
			if errors.Is(err, iterator.Done) {
				err = nil
				if order == nil {
					err = fmt.Errorf("%w for id %s", ErrOrderNotFound, k)
				}
			}

			break
		}

		if order == nil {
			m2 := doc.Data()
			if v, ok := m2["version"]; ok {
				version, _ = strconv.Atoi(v.(string))
			}
			order = documentToOrder(m2)
		} else if canChange(doc.UpdateTime) {
			batch.Delete(doc.Ref)
			batchedItems = true
		}
	}

	if batchedItems {
		_, err = batch.Commit(ctx)
	}

	return order, version, err
}

func orderToDocument(order *persist.Order, version int) map[string]interface{} {
	base, _ := json.Marshal(order.Base)
	tr, _ := json.Marshal(order.Transactions)

	m := map[string]interface{}{
		"base":         base,
		"id":           order.Base.ID.String(),
		"version":      strconv.Itoa(version),
		"status":       order.Status.String(),
		"transactions": tr,
	}

	return m
}

func documentToOrder(m map[string]interface{}) *persist.Order {
	order := &persist.Order{}

	if v, ok := m["status"]; ok {
		order.Status.FromString(v.(string))
	}

	if v, ok := m["transactions"]; ok {
		json.Unmarshal(v.([]byte), &order.Transactions)
	}

	if v, ok := m["base"]; ok {
		json.Unmarshal(v.([]byte), &order.Base)
	}

	return order
}
