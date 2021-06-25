package firebase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/easterthebunny/spew-order/internal/persist"
	"google.golang.org/api/iterator"
)

type BookRepository struct {
	client *firestore.Client
}

func NewBookRepository(client *firestore.Client) *BookRepository {
	return &BookRepository{client: client}
}

// /root/book/{base}-{target}/{BUY|SELL}/{decimal_price}{timestamp}
func (br *BookRepository) BookItemExists(ctx context.Context, item *persist.BookItem) (bool, error) {

	iter := br.getClient(ctx).Collection("book").
		Doc(market(item)).
		Collection(item.Order.Action.String()).
		Where("sort_key", "==", itemKey(item)).
		Limit(1).
		Documents(ctx)

	var err error

	_, err = iter.Next()
	if err != nil {
		if errors.Is(err, iterator.Done) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (br *BookRepository) SetBookItem(ctx context.Context, item *persist.BookItem) error {
	collection := br.getClient(ctx).Collection("book").
		Doc(fmt.Sprintf("%s-%s", item.Order.Base, item.Order.Target)).
		Collection(item.Order.Action.String())

	_, _, err := collection.Add(ctx, bookitemToDocument(item))
	if err != nil {
		return err
	}

	return nil
}

func (br *BookRepository) GetHeadBatch(ctx context.Context, item *persist.BookItem, limit int) (items []*persist.BookItem, err error) {

	iter := br.getClient(ctx).Collection("book").
		Doc(market(item)).
		Collection(item.ActionType.String()).
		Where("sort_key", ">=", itemKey(item)).
		Limit(limit).
		Documents(ctx)

	var doc *firestore.DocumentSnapshot
	for {
		doc, err = iter.Next()
		if err != nil {
			if errors.Is(err, iterator.Done) {
				err = nil
			}

			break
		}

		items = append(items, documentToBookItem(doc.Data()))
	}

	return
}

func (br *BookRepository) DeleteBookItem(ctx context.Context, item *persist.BookItem) error {
	iter := br.getClient(ctx).Collection("book").
		Doc(market(item)).
		Collection(item.Order.Action.String()).
		Where("sort_key", "==", itemKey(item)).
		Limit(1).Documents(ctx)

	batch := br.client.Batch()
	for {
		fmt.Println("checkpoint")
		doc, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.Done) {
				err = nil
			}
			break
		}

		batch.Delete(doc.Ref)
	}
	_, err := batch.Commit(ctx)

	return err
}

func (br *BookRepository) getClient(ctx context.Context) *firestore.Client {

	var client *firestore.Client
	if br.client == nil {
		client = clientFromContext(ctx)
	} else {
		client = br.client
	}
	return client
}

func itemKey(item *persist.BookItem) string {
	price := item.Order.Type.KeyString(item.Order.Action)
	timestamp := item.Order.Timestamp.UnixNano()
	return fmt.Sprintf("%s.%d", price, timestamp)
}

func market(item *persist.BookItem) string {
	return fmt.Sprintf("%s-%s", item.Order.Base, item.Order.Target)
}

func bookitemToDocument(b *persist.BookItem) map[string]interface{} {
	at, _ := json.Marshal(b.ActionType)
	order, _ := json.Marshal(b.Order)

	m := map[string]interface{}{
		"sort_key":    itemKey(b),
		"timestamp":   b.Timestamp.Value(),
		"action_type": at,
		"order":       order,
	}

	return m
}

func documentToBookItem(m map[string]interface{}) *persist.BookItem {
	b := &persist.BookItem{}

	if v, ok := m["timestamp"]; ok {
		b.Timestamp = persist.NanoTime(time.Unix(0, v.(int64)))
	}

	if v, ok := m["action_type"]; ok {
		json.Unmarshal(v.([]byte), &b.ActionType)
	}

	if v, ok := m["order"]; ok {
		json.Unmarshal(v.([]byte), &b.Order)
	}

	return b
}
