package kv

import (
	"fmt"

	"github.com/easterthebunny/spew-order/internal/persist"
)

type BookRepository struct {
	kvstore persist.KVStore
}

func NewBookRepository(store persist.KVStore) *BookRepository {
	return &BookRepository{kvstore: store}
}

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

	return br.kvstore.Set(bookItemKey(*bi), b, &attrs)
}

func (br *BookRepository) GetHeadBatch(bi *persist.BookItem, limit int) (items []*persist.BookItem, err error) {
	query := &persist.KVStoreQuery{
		StartOffset: string(bookItemSubspace(*bi, &bi.ActionType).Bytes())}
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
	return br.kvstore.Delete(bookItemKey(*bi))
}
