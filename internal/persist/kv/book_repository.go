package kv

import (
	"fmt"

	"github.com/easterthebunny/spew-order/internal/key"
	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/pkg/types"
)

type BookRepository struct {
	kvstore persist.KVStore
}

func NewBookRepository(store persist.KVStore) *BookRepository {
	return &BookRepository{kvstore: store}
}

var _ persist.BookRepository = &BookRepository{}

func (br *BookRepository) SetBookItem(bi *persist.BookItem) error {
	if bi == nil {
		return fmt.Errorf("%w for book item", persist.ErrCannotSaveNilValue)
	}

	enc := persist.JSON
	b, err := bi.Encode(enc)
	if err != nil {
		return err
	}

	attrs := persist.KVStoreObjectAttrsToUpdate{
		ContentEncoding: encodingToStr(enc),
		Metadata:        make(map[string]string),
	}

	ky := k(*bi).String()
	return br.kvstore.Set(ky, b, &attrs)
}

func (br *BookRepository) GetHeadBatch(ref *persist.BookItem, limit int) (items []*persist.BookItem, err error) {
	query := getStorageQuery(actionKey(subspace(*ref), ref.ActionType).String())
	attrs, err := br.kvstore.RangeGet(query, 10)
	if err != nil {
		return
	}

	for _, attr := range attrs {
		var data []byte
		data, err = br.kvstore.Get(attr.Name)
		if err != nil {
			return
		}

		item := &persist.BookItem{}
		err = item.Decode(data, encodingFromStr(attr.ContentEncoding))
		if err != nil {
			return
		}

		items = append(items, item)
	}

	return
}

func (br *BookRepository) DeleteBookItem(bi *persist.BookItem) error {
	ky := k(*bi).String()
	return br.kvstore.Delete(ky)
}

func getStorageQuery(offset string) *persist.KVStoreQuery {
	// get the head of the list for the opposite action type
	query := &persist.KVStoreQuery{
		StartOffset: offset}

	return query
}

// k generates a key that will sort ASC lexigraphically, but remain in type
// sorted order: buys are sorted largest/oldest to smallest/newest and sells
// are sorted smallest/oldest to largest/newest
func k(o persist.BookItem) key.Key {
	t := o.Order.Type.KeyTuple(o.Order.Action)
	t = append(t, key.Tuple{o.Order.Timestamp.UnixNano()}...)
	return actionSubspace(subspace(o), o.Order.Action).Pack(t)
}

// actionTypeKey generates a key that will find a sorted match in the opposite order book
func actionTypeKey(o persist.BookItem) key.Key {
	return actionSubspace(subspace(o), o.ActionType).Pack(o.Order.Type.KeyTuple(o.Order.Action))
}

// HeadKey returns a key that can be used to range query a lexigraphically sorted set
func headKey(o persist.BookItem) key.Key {
	return subspace(o).Pack(key.Tuple{uint(o.Order.Action)})
}

func subspace(o persist.BookItem) key.Subspace {
	return gsBook.Sub(uint(o.Order.Base)).Sub(uint(o.Order.Target))
}

func actionKey(sub key.Subspace, action types.ActionType) key.Key {
	return sub.Pack(key.Tuple{uint(action)})
}

func actionSubspace(sub key.Subspace, action types.ActionType) key.Subspace {
	return sub.Sub(uint(action))
}
