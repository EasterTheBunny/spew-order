package kv

import (
	"context"
	"fmt"

	"github.com/easterthebunny/spew-order/internal/key"
	"github.com/easterthebunny/spew-order/internal/persist"
)

type BookRepository struct {
	kvstore persist.KVStore
}

func NewBookRepository(store persist.KVStore) *BookRepository {
	return &BookRepository{kvstore: store}
}

func (br *BookRepository) BookItemExists(ctx context.Context, item *persist.BookItem) (bool, error) {
	query := &persist.KVStoreQuery{
		StartOffset: bookItemKey(*item)}

	attrs, err := br.kvstore.RangeGet(query, 1)
	if err != nil {
		return false, err
	}

	if len(attrs) == 1 {
		return true, nil
	}

	return false, nil
}

func (br *BookRepository) SetBookItem(ctx context.Context, bi *persist.BookItem) error {
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

func (br *BookRepository) GetHeadBatch(ctx context.Context, bi *persist.BookItem, limit int, offset *persist.BookItem) (items []*persist.BookItem, err error) {
	query := &persist.KVStoreQuery{
		StartOffset: bookItemSubspace(*bi, &bi.ActionType).Pack(key.Tuple{}).String()}
	attrs, err := br.kvstore.RangeGet(query, 10)
	if err != nil {
		return
	}

	for _, attr := range attrs {
		var data []byte
		data, err = br.kvstore.Get(attr.Name)
		if err != nil {
			err = fmt.Errorf("Book::GetHeadBatch -- %w", err)
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

func (br *BookRepository) DeleteBookItem(ctx context.Context, bi *persist.BookItem) error {
	return br.kvstore.Delete(bookItemKey(*bi))
}
