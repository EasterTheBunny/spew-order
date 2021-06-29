package firebase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
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

type bookItemDocument struct {
	Version    int64     `firestore:"version"`
	Delete     bool      `firestore:"remove"`
	Created    time.Time `firestore:"created"`
	SortKey    string    `firestore:"sort_key"`
	Timestamp  time.Time `firestore:"timestamp"`
	ActionType []byte    `firestore:"action_type"`
	Order      []byte    `firestore:"order"`
}

// BookItemExists searches the data store by the item key. Returns true if the item does exist.
func (br *BookRepository) BookItemExists(ctx context.Context, item *persist.BookItem) (bool, error) {

	_, doc, err := br.getBookItemDocument(ctx, item)

	if err != nil {
		err = fmt.Errorf("BookItemExists::%w", err)
	}
	return doc != nil, err
}

func (br *BookRepository) SetBookItem(ctx context.Context, item *persist.BookItem) error {

	_, doc, err := br.getBookItemDocument(ctx, item)
	if err != nil {
		return err
	}

	newDoc := bookitemToDocument(item)
	if doc != nil {
		newDoc.Version = doc.Version + 1
	}

	_, _, err = br.getClient(ctx).Collection("book").
		Doc(market(item)).
		Collection(item.Order.Action.String()).Add(ctx, &newDoc)

	if err != nil {
		err = fmt.Errorf("SetBookItem::%w", err)
	}

	return err
}

func (br *BookRepository) getBookItemDocument(ctx context.Context, item *persist.BookItem) (*firestore.DocumentRef, *bookItemDocument, error) {

	iter := br.getClient(ctx).Collection("book").
		Doc(market(item)).
		Collection(item.Order.Action.String()).
		Where("sort_key", "==", itemKey(item)).
		Documents(ctx)

	var err error
	var document *bookItemDocument
	var ref *firestore.DocumentRef
	batch := br.getClient(ctx).Batch()
	batchedItems := false

	var doc *firestore.DocumentSnapshot
	for {
		doc, err = iter.Next()
		if err != nil {
			if errors.Is(err, iterator.Done) {
				err = nil
			} else {
				return nil, nil, err
			}

			break
		}

		var current bookItemDocument
		doc.DataTo(&current)

		if document == nil || current.Version > document.Version {
			document = &current
			ref = doc.Ref
		} else if canChange(current.Created) {
			// delete current from database if possible
			batch.Delete(doc.Ref)
			batchedItems = true
		}

		if document != nil && document.Delete {
			batch.Delete(doc.Ref)
			batchedItems = true
			document = nil
			ref = nil
		}
	}

	if batchedItems {
		_, err = batch.Commit(ctx)
	}

	return ref, document, err

}

func (br *BookRepository) GetHeadBatch(ctx context.Context, item *persist.BookItem, limit int) (items []*persist.BookItem, err error) {

	iter := br.getClient(ctx).Collection("book").
		Doc(market(item)).
		Collection(item.ActionType.String()).
		OrderBy("sort_key", firestore.Asc).
		Limit(limit * 3).
		Documents(ctx)

	keep := make(map[string][]*firestore.DocumentSnapshot)
	remove := make(map[string][]*firestore.DocumentSnapshot)

	batch := br.getClient(ctx).Batch()
	batchedItems := false

	var doc *firestore.DocumentSnapshot
snapshots:
	for {
		doc, err = iter.Next()
		if err != nil {
			if errors.Is(err, iterator.Done) {
				err = nil
			} else {
				err = fmt.Errorf("GetHeadBatch: %w", err)
			}

			break
		}

		var interfaceKey interface{}
		interfaceKey, err = doc.DataAt("sort_key")
		if err != nil {
			return
		}
		sortKey := interfaceKey.(string)

		kp, inKeep := keep[sortKey]
		dl, inDel := remove[sortKey]

		switch {
		case inKeep:
			// if the item id exists in the keep or delete bucket, add the item to the list for processing
			keep[sortKey] = append(kp, doc)

			var deleteVal interface{}
			deleteVal, err = doc.DataAt("remove")
			if err != nil {
				return
			}
			deleteKey := deleteVal.(bool)

			if deleteKey {
				remove[sortKey] = keep[sortKey]
				delete(keep, sortKey)
			}
			continue
		case inDel:
			remove[sortKey] = append(dl, doc)
			continue
		case len(keep) < limit:
			// if the length of the keep bucket is less than the limit add the item to the list for processing
			keep[sortKey] = []*firestore.DocumentSnapshot{doc}
			continue
		default:
			// if the length of the keep bucket is equal to the limit, break from the loop
			break snapshots
		}
	}

	keys := make([]string, 0, len(keep))
	for k := range keep {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] > keys[j]
	})

	for _, key := range keys {
		docs := keep[key]
		var current *bookItemDocument
		var ref *firestore.DocumentRef

		for _, doc = range docs {
			var item bookItemDocument
			doc.DataTo(&item)

			switch {
			case current == nil:
				current = &item
				ref = doc.Ref
			case item.Version > current.Version:
				current = &item
				batch.Delete(ref)
				ref = doc.Ref
				batchedItems = true
			}
		}

		items = append(items, documentToBookItem(current))
	}

	for _, docs := range remove {
		for _, doc = range docs {
			batch.Delete(doc.Ref)
			batchedItems = true
		}
	}

	if batchedItems {
		_, err = batch.Commit(ctx)
	}

	if err != nil {
		err = fmt.Errorf("GetHeadBatch::%w", err)
	}

	return
}

func (br *BookRepository) DeleteBookItem(ctx context.Context, item *persist.BookItem) error {

	ref, doc, err := br.getBookItemDocument(ctx, item)
	if err != nil {
		return err
	}

	batch := br.getClient(ctx).Batch()
	batchedItems := false
	if doc != nil && canChange(doc.Created) {
		batch.Delete(ref)
		batchedItems = true
	}

	if batchedItems {
		_, err = batch.Commit(ctx)
	} else {
		newDoc := bookitemToDocument(item)
		if doc != nil {
			newDoc.Version = doc.Version + 1
			newDoc.Delete = true
		}

		_, _, err = br.getClient(ctx).Collection("book").
			Doc(market(item)).
			Collection(item.Order.Action.String()).Add(ctx, &newDoc)
	}

	if err != nil {
		err = fmt.Errorf("DeleteBookItem::%w", err)
	}

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

// itemKey generates a key that will sort ASC lexicographically, but remain in
// type sorted order: buys are sorted largest/oldest to smallest/newest and sells
// are sorted smallest/oldest to largest/newest
func itemKey(item *persist.BookItem) string {
	price := item.Order.Type.KeyString(item.Order.Action)
	timestamp := item.Order.Timestamp.UnixNano()
	return fmt.Sprintf("%s.%011d", price, timestamp)
}

func market(item *persist.BookItem) string {
	return fmt.Sprintf("%s-%s", item.Order.Base, item.Order.Target)
}

func bookitemToDocument(b *persist.BookItem) *bookItemDocument {
	at, _ := json.Marshal(b.ActionType)
	order, _ := json.Marshal(b.Order)
	return &bookItemDocument{
		Version:    0,
		Delete:     false,
		Created:    time.Now(),
		SortKey:    itemKey(b),
		Timestamp:  time.Time(b.Timestamp),
		ActionType: at,
		Order:      order,
	}
}

func documentToBookItem(doc *bookItemDocument) *persist.BookItem {

	b := &persist.BookItem{
		Timestamp: persist.NanoTime(doc.Timestamp),
	}

	json.Unmarshal(doc.ActionType, &b.ActionType)
	json.Unmarshal(doc.Order, &b.Order)

	return b
}
