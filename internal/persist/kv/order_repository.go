package kv

import (
	"fmt"

	"github.com/easterthebunny/spew-order/internal/key"
	"github.com/easterthebunny/spew-order/internal/persist"
)

type OrderRepository struct {
	kvstore persist.KVStore
	account *persist.Account
}

func NewOrderRepository(store persist.KVStore, account *persist.Account) *OrderRepository {
	return &OrderRepository{kvstore: store, account: account}
}

func (or *OrderRepository) GetOrder(k persist.Key) (order *persist.Order, err error) {

	b, err := or.kvstore.Get(orderIDKey(*or.account, k))
	if err != nil {
		return
	}

	attr, err := or.kvstore.Attrs(orderIDKey(*or.account, k))
	if err != nil {
		return
	}

	order = &persist.Order{}
	err = order.Decode(b, encodingFromStr(attr.ContentEncoding))

	return
}

func (or *OrderRepository) SetOrder(o *persist.Order) error {
	if o == nil {
		return fmt.Errorf("%w for order", persist.ErrCannotSaveNilValue)
	}

	enc := persist.JSON
	b, err := o.Encode(enc)
	if err != nil {
		return err
	}

	attrs := persist.KVStoreObjectAttrsToUpdate{
		ContentEncoding: encodingToStr(enc),
		Metadata:        make(map[string]string),
	}

	return or.kvstore.Set(orderKey(*or.account, *o), b, &attrs)
}

func (or *OrderRepository) GetOrdersByStatus(s ...persist.FillStatus) (orders []*persist.Order, err error) {

	m := make(map[persist.FillStatus]bool)
	for _, st := range s {
		m[st] = true
	}

	q := persist.KVStoreQuery{
		StartOffset: orderSubspace(*or.account).Pack(key.Tuple{}).String()}

	attrs, err := or.kvstore.RangeGet(&q, 20)
	if err != nil {
		return
	}

	for _, attr := range attrs {
		var bts []byte
		bts, err = or.kvstore.Get(attr.Name)
		if err != nil {
			return
		}

		ord := &persist.Order{}
		err = ord.Decode(bts, encodingFromStr(attr.ContentEncoding))
		if err != nil {
			return
		}

		if _, ok := m[ord.Status]; ok {
			orders = append(orders, ord)
		}
	}

	return
}

func (or *OrderRepository) UpdateOrderStatus(k persist.Key, s persist.FillStatus, tr []string) error {

	order, err := or.GetOrder(k)
	if err != nil {
		return err
	}

	order.Status = s
	if len(tr) > 0 {
		order.Transactions = append(order.Transactions, tr)
	}
	return or.SetOrder(order)
}
