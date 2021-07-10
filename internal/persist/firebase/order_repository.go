package firebase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"cloud.google.com/go/firestore"
	"github.com/easterthebunny/spew-order/internal/persist"
	uuid "github.com/satori/go.uuid"
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
	var err error
	var order *persist.Order

	err = or.getClient(ctx).RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		var txErr error

		order, _, txErr = or.getOrder(ctx, tx, k)
		if txErr != nil {
			txErr = fmt.Errorf("GetOrder: %w", txErr)
		}

		return txErr
	})

	return order, err
}

func (or *OrderRepository) SetOrder(ctx context.Context, o *persist.Order) error {
	return or.getClient(ctx).RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		var txErr error

		_, version, txErr := or.getOrder(ctx, tx, o.Base.ID)
		if txErr != nil {
			if errors.Is(txErr, ErrOrderNotFound) {
				txErr = nil
				version = 0
			} else {
				return fmt.Errorf("SetOrder: %w", txErr)
			}
		}
		version++

		col := fmt.Sprintf("accounts/%s/orders", or.account.ID)
		d := or.getClient(ctx).Collection(col).Doc(uuid.NewV4().String())
		txErr = tx.Create(d, orderToDocument(o, 0))
		if txErr != nil {
			return fmt.Errorf("SetOrder: %w", txErr)
		}

		return txErr
	})
}

func (or *OrderRepository) GetOrdersByStatus(ctx context.Context, s ...persist.FillStatus) (orders []*persist.Order, err error) {

	ops := make([]string, len(s))
	for i, v := range s {
		ops[i] = v.String()
	}

	client := or.getClient(ctx)
	col := fmt.Sprintf("accounts/%s/orders", or.account.ID)

	versionMap := make(map[string]bool)

	err = client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		var txErr error
		iter := tx.Documents(client.Collection(col).
			Where("status", "in", ops).
			OrderBy("version", firestore.Desc))

		var doc *firestore.DocumentSnapshot
		var order *persist.Order
		for {
			doc, txErr = iter.Next()
			if txErr != nil {
				if errors.Is(txErr, iterator.Done) {
					txErr = nil
				} else {
					txErr = fmt.Errorf("GetOrdersByStatus: %w", txErr)
				}

				break
			}

			order = documentToOrder(doc.Data())
			if _, ok := versionMap[order.Base.ID.String()]; !ok {
				versionMap[order.Base.ID.String()] = true
				orders = append(orders, order)
			} else if canChange(doc.UpdateTime) {
				// delete the older version
				txErr = tx.Delete(doc.Ref)
				if txErr != nil {
					return txErr
				}
			}
		}
		iter.Stop()

		return txErr
	})

	return
}

func (or *OrderRepository) UpdateOrderStatus(ctx context.Context, k persist.Key, s persist.FillStatus, tr []string) error {
	return or.getClient(ctx).RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		var txErr error
		var o *persist.Order
		var version int

		o, version, txErr = or.getOrder(ctx, tx, k)
		if txErr != nil {
			return fmt.Errorf("UpdateOrderStatus: %w", txErr)
		}
		version++

		o.Status = s
		if len(tr) > 0 {
			o.Transactions = append(o.Transactions, tr)
		}

		col := fmt.Sprintf("accounts/%s/orders", or.account.ID)
		d := or.getClient(ctx).Collection(col).Doc(uuid.NewV4().String())
		txErr = tx.Create(d, orderToDocument(o, version))
		if txErr != nil {
			return fmt.Errorf("UpdateOrderStatus: %w", txErr)
		}

		return txErr
	})
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

func (or *OrderRepository) getOrder(ctx context.Context, tx *firestore.Transaction, k persist.Key) (*persist.Order, int, error) {
	var err error
	client := or.getClient(ctx)
	var order *persist.Order
	var version int = 0

	col := fmt.Sprintf("accounts/%s/orders", or.account.ID)
	iter := tx.Documents(client.Collection(col).Where("id", "==", k.String()).OrderBy("version", firestore.Desc))

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
			err = tx.Delete(doc.Ref)
			if err != nil {
				return order, version, err
			}
		}
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
