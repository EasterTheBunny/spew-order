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
	"github.com/easterthebunny/spew-order/pkg/types"
	uuid "github.com/satori/go.uuid"
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
	var doc *bookItemDocument
	var err error

	err = br.getClient(ctx).RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		var txErr error

		_, doc, txErr = br.getBookItemDocument(ctx, tx, item)

		return txErr
	})

	if err != nil {
		err = fmt.Errorf("BookItemExists::%w", err)
	}
	return doc != nil, err
}

func (br *BookRepository) SetBookItem(ctx context.Context, item *persist.BookItem) error {
	var err error

	err = br.getClient(ctx).RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		var txErr error

		_, doc, txErr := br.getBookItemDocument(ctx, tx, item)
		if txErr != nil {
			return txErr
		}

		newDoc := bookitemToDocument(item)
		if doc != nil {
			newDoc.Version = doc.Version + 1
		}

		ref := br.baseCollection(ctx, item, item.Order.Action).Doc(uuid.NewV4().String())
		txErr = tx.Create(ref, &newDoc)
		if txErr != nil {
			return txErr
		}

		return txErr
	})

	if err != nil {
		err = fmt.Errorf("SetBookItem::%w", err)
	}

	return err
}

func (br *BookRepository) baseCollection(ctx context.Context, item *persist.BookItem, action types.ActionType) *firestore.CollectionRef {
	return br.getClient(ctx).Collection("book").
		Doc(market(item)).
		Collection(action.String())
}

func (br *BookRepository) getBookItemDocument(ctx context.Context, tx *firestore.Transaction, item *persist.BookItem) (*firestore.DocumentRef, *bookItemDocument, error) {

	col := br.baseCollection(ctx, item, item.Order.Action).
		Where("sort_key", "==", itemKey(item))

	iter := tx.Documents(col)

	var err error
	var document *bookItemDocument
	var ref *firestore.DocumentRef

	// get the most recent version of the document from storage
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

		// if the current has a version number larger than the last
		// document, replace it and delete the replaced document
		// otherwise delete the current document if possible
		if document == nil || current.Version > document.Version {
			if document != nil && canChange(document.Created) {
				err = tx.Delete(doc.Ref)
				if err != nil {
					break
				}
			}

			document = &current
			ref = doc.Ref
		} else if canChange(current.Created) {
			// delete current from database if possible
			err = tx.Delete(doc.Ref)
			if err != nil {
				break
			}
		}
	}

	// if the document is supposed to be deleted, do so if possible
	// and set the return values as null
	if document != nil && document.Delete {
		if canChange(document.Created) {
			err = tx.Delete(doc.Ref)
		}
		document = nil
		ref = nil
	}

	return ref, document, err
}

func (br *BookRepository) GetHeadBatch(ctx context.Context, item *persist.BookItem, limit int) (items []*persist.BookItem, err error) {

	err = br.getClient(ctx).RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		var txErr error

		col := br.baseCollection(ctx, item, item.ActionType).
			OrderBy("sort_key", firestore.Asc).
			Limit(limit * 3)
		iter := tx.Documents(col)

		keep := make(map[string][]*firestore.DocumentSnapshot)
		remove := make(map[string][]*firestore.DocumentSnapshot)

		var doc *firestore.DocumentSnapshot
	snapshots:
		for {
			doc, txErr = iter.Next()
			if txErr != nil {
				if errors.Is(txErr, iterator.Done) {
					txErr = nil
				} else {
					txErr = fmt.Errorf("GetHeadBatch: %w", txErr)
				}

				break
			}

			var interfaceKey interface{}
			interfaceKey, txErr = doc.DataAt("sort_key")
			if err != nil {
				return txErr
			}
			sortKey := interfaceKey.(string)

			kp, inKeep := keep[sortKey]
			dl, inDel := remove[sortKey]

			switch {
			case inKeep:
				// if the item id exists in the keep or delete bucket, add the item to the list for processing
				keep[sortKey] = append(kp, doc)

				var deleteVal interface{}
				deleteVal, txErr = doc.DataAt("remove")
				if err != nil {
					return txErr
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
		sort.Strings(keys)

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
					txErr = tx.Delete(ref)
					if txErr != nil {
						break
					}
					ref = doc.Ref
				}
			}

			items = append(items, documentToBookItem(current))
		}

		for _, docs := range remove {
			for _, doc = range docs {
				txErr = tx.Delete(doc.Ref)
				if txErr != nil {
					return txErr
				}
			}
		}

		return txErr
	})

	if err != nil {
		err = fmt.Errorf("GetHeadBatch::%w", err)
	}

	return
}

func (br *BookRepository) DeleteBookItem(ctx context.Context, item *persist.BookItem) error {
	var err error

	err = br.getClient(ctx).RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		var txErr error

		ref, doc, txErr := br.getBookItemDocument(ctx, tx, item)
		if err != nil {
			return err
		}

		if doc != nil && canChange(doc.Created) {
			tx.Delete(ref)
		} else {
			newDoc := bookitemToDocument(item)
			if doc != nil {
				newDoc.Version = doc.Version + 1
				newDoc.Delete = true
			}

			col := br.baseCollection(ctx, item, item.Order.Action).Doc(uuid.NewV4().String())

			txErr = tx.Create(col, &newDoc)
		}

		return txErr
	})

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
