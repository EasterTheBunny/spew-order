package firebase

import (
	"context"
	"fmt"

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

// /root/book/{base}/{target}/{BUY|SELL}/{decimal_price}{timestamp}
func (br *BookRepository) BookItemExists(ctx context.Context, item *persist.BookItem) (bool, error) {
	price := item.Order.Type.KeyString(item.Order.Action)
	timestamp := item.Order.Timestamp.UnixNano()

	iter := br.getClient(ctx).Collection("book").
		Doc(fmt.Sprintf("%s-%s", item.Order.Base, item.Order.Target)).
		Collection(item.Order.Action.String()).
		Where("sort_key", "==", fmt.Sprintf("%s%d", price, timestamp)).
		Limit(1).
		Documents(ctx)

	var err error

	_, err = iter.Next()
	if err == iterator.Done {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

func (br *BookRepository) SetBookItem(ctx context.Context, item *persist.BookItem) error {
	// price := item.Order.Type.KeyString(item.Order.Action)
	// timestamp := item.Order.Timestamp.UnixNano()

	collection := br.getClient(ctx).Collection("book").
		Doc(fmt.Sprintf("%s-%s", item.Order.Base, item.Order.Target)).
		Collection(item.Order.Action.String())

	_, _, err := collection.Add(ctx, item)
	if err != nil {
		return err
	}

	return nil
}

func (br *BookRepository) GetHeadBatch(bi *persist.BookItem, limit int) (items []*persist.BookItem, err error) {
	return
}

func (br *BookRepository) DeleteBookItem(bi *persist.BookItem) error {
	return nil
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
