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

var (
	errProcessLimitReached = errors.New("defined limit reached")
	ErrNotFound            = errors.New("book item does not exist")
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
	var highestVersion *bookItemDocument
	var highestVersionRef *firestore.DocumentRef
	var highestVersionUpdateTime time.Time

	// get the most recent version of the document from storage
	var snapshot *firestore.DocumentSnapshot
	for {
		snapshot, err = iter.Next()
		if err != nil {
			if errors.Is(err, iterator.Done) {
				err = nil
			} else {
				return nil, nil, err
			}

			break
		}

		var document bookItemDocument
		snapshot.DataTo(&document)

		// if the current has a version number larger than the last
		// document, replace it and delete the replaced document
		// otherwise delete the current document if possible
		if highestVersion == nil || document.Version > highestVersion.Version {
			var deleteable *firestore.DocumentRef
			// don't do the delete operation until after the version swap
			// is done in case of an error in the delete
			if highestVersion != nil && canChange(highestVersionUpdateTime) {
				deleteable = highestVersionRef
			}

			highestVersion = &document
			highestVersionRef = snapshot.Ref
			highestVersionUpdateTime = snapshot.UpdateTime

			if deleteable != nil {
				err = tx.Delete(deleteable)
				if err != nil {
					break
				}
			}
		} else if canChange(snapshot.UpdateTime) {
			// delete current from database if possible
			err = tx.Delete(snapshot.Ref)
			if err != nil {
				break
			}
		}
	}

	// if the document is supposed to be deleted, do so if possible
	// and set the return values as null
	if highestVersion != nil && highestVersion.Delete {
		if canChange(highestVersionUpdateTime) {
			err = tx.Delete(highestVersionRef)
		}
		highestVersion = nil
		highestVersionRef = nil
	}

	return highestVersionRef, highestVersion, err
}

func (br *BookRepository) sortBatchSnapshot(snapshot *firestore.DocumentSnapshot, keep, remove map[string][]*firestore.DocumentSnapshot, limit int) error {
	var err error

	var interfaceKey interface{}
	interfaceKey, err = snapshot.DataAt("sort_key")
	if err != nil {
		return err
	}
	sortKey := interfaceKey.(string)

	kp, inKeep := keep[sortKey]
	dl, inDel := remove[sortKey]

	switch {
	case inKeep:
		// if the item key already exists in the keep bucket, the incoming
		// snapshot is also a keep by default
		keep[sortKey] = append(kp, snapshot)

		// evaluate the remove flag for the incoming snapshot
		var deleteVal interface{}
		deleteVal, err = snapshot.DataAt("remove")
		if err != nil {
			return err
		}
		deleteKey := deleteVal.(bool)

		// if the remove flag is set, the book item is closed and should be
		// removed from the book along with all of its versions
		if deleteKey {
			remove[sortKey] = keep[sortKey]
			delete(keep, sortKey)
		}
		return nil
	case inDel:
		// if the item key exists in the remove bucket, add this version
		// for cleanup
		remove[sortKey] = append(dl, snapshot)
		return nil
	case len(keep) < limit:
		// for an item key that doesn't exist in either bucket and the total
		// limit has not been reached, add the item to the keep bucket to start

		// evaluate the remove flag for the incoming snapshot
		var deleteVal interface{}
		deleteVal, err = snapshot.DataAt("remove")
		if err != nil {
			return err
		}
		deleteKey := deleteVal.(bool)

		// if the item is to be removed, put it in the remove bucket
		if deleteKey {
			remove[sortKey] = []*firestore.DocumentSnapshot{snapshot}
		} else {
			keep[sortKey] = []*firestore.DocumentSnapshot{snapshot}
		}

		return nil
	default:
		// if the length of the keep bucket is equal to the limit, break from the loop
		return errProcessLimitReached
	}
}

func (br *BookRepository) selectHighestVersionSnapshot(docs []*firestore.DocumentSnapshot, tx *firestore.Transaction) (*bookItemDocument, error) {
	var err error
	var highestVersion *bookItemDocument
	var highestVersionRef *firestore.DocumentRef
	var highestVersionUpdateTime time.Time

	for _, snapshot := range docs {
		var currentItem bookItemDocument
		snapshot.DataTo(&currentItem)

		switch {
		case highestVersion == nil:
			highestVersion = &currentItem
			highestVersionRef = snapshot.Ref
			highestVersionUpdateTime = snapshot.UpdateTime
		case currentItem.Version > highestVersion.Version:
			var deletable *firestore.DocumentRef
			if canChange(highestVersionUpdateTime) {
				deletable = highestVersionRef
			}

			highestVersion = &currentItem
			highestVersionRef = snapshot.Ref

			if deletable != nil {
				err = tx.Delete(deletable)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	return highestVersion, err
}

func (br *BookRepository) GetHeadBatch(ctx context.Context, item *persist.BookItem, limit int, offset *persist.BookItem) (items []*persist.BookItem, err error) {

	err = br.getClient(ctx).RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		var txErr error

		// TODO: getting the entire order book is inefficient; refactor into a cursor
		col := br.baseCollection(ctx, item, item.ActionType)
		var qry firestore.Query

		if offset != nil {
			qry = col.Where("sort_key", ">", itemKey(offset)).OrderBy("sort_key", firestore.Asc)
		} else {
			qry = col.OrderBy("sort_key", firestore.Asc)
			//Limit(limit * 3)
		}

		iter := tx.Documents(qry)

		keep := make(map[string][]*firestore.DocumentSnapshot)
		remove := make(map[string][]*firestore.DocumentSnapshot)

		var snapshot *firestore.DocumentSnapshot
		for {
			snapshot, txErr = iter.Next()
			if txErr != nil {
				if errors.Is(txErr, iterator.Done) {
					txErr = nil
				} else {
					txErr = fmt.Errorf("GetHeadBatch: %w", txErr)
				}

				break
			}

			err = br.sortBatchSnapshot(snapshot, keep, remove, limit)
			if errors.Is(err, errProcessLimitReached) {
				err = nil
				break
			}
		}

		keys := make([]string, 0, len(keep))
		for k := range keep {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, key := range keys {
			var highestVersion *bookItemDocument
			highestVersion, txErr = br.selectHighestVersionSnapshot(keep[key], tx)
			if txErr != nil {
				return txErr
			}

			items = append(items, documentToBookItem(highestVersion))
		}

		for _, docs := range remove {
			for _, snapshot = range docs {
				if canChange(snapshot.UpdateTime) {
					txErr = tx.Delete(snapshot.Ref)
					if txErr != nil {
						return txErr
					}
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

// DeleteBookItem attempts to delete the book item if it exists. Returns an error if
// no book item is found.
func (br *BookRepository) DeleteBookItem(ctx context.Context, item *persist.BookItem) error {
	var err error

	err = br.getClient(ctx).RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		var txErr error

		ref, doc, txErr := br.getBookItemDocument(ctx, tx, item)
		if txErr != nil {
			return txErr
		}

		if doc == nil {
			return ErrNotFound
		}

		if canChange(doc.Created) {
			txErr = tx.Delete(ref)
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
