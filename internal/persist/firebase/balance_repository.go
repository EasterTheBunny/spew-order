package firebase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/pkg/types"
	"github.com/shopspring/decimal"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrBalanceDocumentNotFound = errors.New("balance document not found")
	ErrHoldNotFound            = errors.New("hold not found")
)

type BalanceRepository struct {
	client  *firestore.Client
	account *persist.Account
	symbol  types.Symbol
}

func NewBalanceRepository(client *firestore.Client, a *persist.Account, s types.Symbol) *BalanceRepository {
	return &BalanceRepository{
		client:  client,
		account: a,
		symbol:  s,
	}
}

type balanceDocument struct {
	Balance string    `firestore:"balance"`
	Updated time.Time `firestore:"updated"`
	Created time.Time `firestore:"created"`
}

type balanceItemDocument struct {
	Version   int64     `firestore:"version"`
	ID        string    `firestore:"id"`
	Timestamp time.Time `firestore:"timestamp"`
	Created   time.Time `firestore:"created"`
	Amount    string    `firestore:"amount"`
}

// GetBalance gets the balance record for the defined repository parameters. If the record does
// not exist, it is created. This method continually rolls up all balance postings into the
// balance document and avoids the firestore write contention by checking the last update time
// on the balance document and the create date on the post document being rolled up. All updates
// are completed as a batch operation.
func (b *BalanceRepository) GetBalance(ctx context.Context) (balance decimal.Decimal, err error) {

	balance = decimal.NewFromInt(0)
	storeBalance := decimal.NewFromInt(0)

	balRef := b.getSymbolDocumentRef(ctx)
	err = b.getClient(ctx).RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		var txErr error
		execUpdate := false

		// get the balance document from storage
		// if the document is not found, attempt to create it
		balDoc, txErr := tx.Get(balRef)
		if txErr != nil {
			if status.Code(txErr) != codes.NotFound {
				return txErr
			}

			newDoc := &balanceDocument{
				Balance: "0",
				Created: time.Now(),
				Updated: time.Now(),
			}

			return tx.Set(balRef, newDoc)
		}

		docBalString, txErr := balDoc.DataAt("balance")
		if txErr != nil {
			return txErr
		}

		balance, txErr = decimal.NewFromString(docBalString.(string))
		if txErr != nil {
			return txErr
		}

		// collect all post items with their document refs
		iter := tx.Documents(b.getPostCollection(ctx))
		var doc *firestore.DocumentSnapshot
		for {
			doc, txErr = iter.Next()
			if txErr != nil {
				if errors.Is(txErr, iterator.Done) {
					txErr = nil
				}

				break
			}
			execUpdate = true

			amtStr, txErr := doc.DataAt("amount")
			if txErr != nil {
				return txErr
			}

			amt, txErr := decimal.NewFromString(amtStr.(string))
			if txErr != nil {
				return txErr
			}

			// the returned balance amount should have all posts applied to it
			balance = balance.Add(amt)

			// the post is allowed to be deleted only if the balance document AND
			// the post document are last updated more than 1 second ago
			// if they can both be edited, add the post amount to the change
			// amount and delete the post
			if canChange(doc.UpdateTime) && canChange(balDoc.UpdateTime) {
				storeBalance = storeBalance.Add(amt)
				txErr = tx.Delete(doc.Ref)
				if txErr != nil {
					return txErr
				}
			}
		}

		// if the balance document can be updated and it needs to be updated
		// execute the update
		if execUpdate && canChange(balDoc.UpdateTime) {
			return tx.Set(balRef, map[string]interface{}{
				"balance": storeBalance.StringFixedBank(b.symbol.RoundingPlace()),
				"updated": time.Now(),
			}, firestore.MergeAll)
		} else {
			return nil
		}
	})

	return balance, err
}

func (b *BalanceRepository) AddToBalance(ctx context.Context, amt decimal.Decimal) error {
	t := time.Now()
	item := balanceItemDocument{
		Version:   0,
		ID:        fmt.Sprintf("%d", t.Unix()),
		Timestamp: t,
		Created:   t,
		Amount:    amt.StringFixedBank(b.symbol.RoundingPlace()),
	}

	col := fmt.Sprintf("accounts/%s/symbols/%s/posts", b.account.ID, b.symbol)
	_, _, err := b.getClient(ctx).Collection(col).Add(ctx, &item)

	return err
}

func (b *BalanceRepository) getPostCollection(ctx context.Context) *firestore.CollectionRef {
	return b.getSymbolDocumentRef(ctx).Collection("posts")
}

func (b *BalanceRepository) getHoldCollection(ctx context.Context) *firestore.CollectionRef {
	return b.getSymbolDocumentRef(ctx).Collection("holds")
}

func (b *BalanceRepository) getSymbolDocumentRef(ctx context.Context) *firestore.DocumentRef {
	return b.getClient(ctx).Collection("accounts").Doc(b.account.ID).Collection("symbols").Doc(b.symbol.String())
}

func (b *BalanceRepository) getBalanceItemDocuments(ctx context.Context, collection *firestore.CollectionRef) (refs []*firestore.DocumentRef, items []*balanceItemDocument, err error) {
	versions := make(map[string]*balanceItemDocument)
	vRefs := make(map[string]*firestore.DocumentRef)

	err = b.getClient(ctx).RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		var txErr error
		iter := collection.OrderBy("timestamp", firestore.Asc).Documents(ctx)
		var doc *firestore.DocumentSnapshot
		for {
			doc, txErr = iter.Next()
			if txErr != nil {
				if errors.Is(txErr, iterator.Done) {
					txErr = nil
				}
				break
			}

			var current balanceItemDocument
			doc.DataTo(&current)

			// build key for looking up last encountered item version
			vStr := current.ID
			version, ok := versions[vStr]
			if !ok { // add the current item to the version lookup if its not there
				versions[vStr] = &current
				vRefs[vStr] = doc.Ref
			} else {
				// if the current item version is greater than the lookup version
				// update the lookup version with the current version
				// and delete the lesser version
				if current.Version > version.Version {
					// delete old version if possible
					r, ok := vRefs[vStr]
					if ok && canChange(version.Created) {
						tx.Delete(r)
					}

					versions[vStr] = &current
					vRefs[vStr] = doc.Ref
				} else if canChange(current.Created) {
					// delete old version if possible
					tx.Delete(doc.Ref)
				}
			}
		}

		zero := decimal.NewFromInt(0)
		for k, v := range versions {
			r, ok := vRefs[k]
			if ok {

				// the latest version could still be a 0 value item
				// in this case, the version should be deleted
				// if it cannot be deleted, do not include it in the
				// returned slice
				var amt decimal.Decimal
				amt, txErr = decimal.NewFromString(v.Amount)
				if txErr != nil {
					return txErr
				}

				if amt.Equal(zero) && canChange(v.Created) {
					tx.Delete(r)
				} else if !amt.Equal(zero) {
					refs = append(refs, r)
					items = append(items, v)
				}
			}
		}

		return nil
	})

	return
}

// FindHolds returns a list of balance items associated with holds. This function deletes older versions
// of the same hold to prevent write contention.
func (b *BalanceRepository) FindHolds(ctx context.Context) (holds []*persist.BalanceItem, err error) {

	_, docs, err := b.getBalanceItemDocuments(ctx, b.getHoldCollection(ctx))

	for _, doc := range docs {
		var amt decimal.Decimal
		amt, err = decimal.NewFromString(doc.Amount)
		if err != nil {
			return
		}

		hold := persist.BalanceItem{
			ID:        doc.ID,
			Timestamp: persist.NanoTime(doc.Created),
			Amount:    amt,
		}
		holds = append(holds, &hold)
	}

	return
}

// /root/account/{accountid}/symbol/{symbol}/hold/{holdid}
func (b *BalanceRepository) CreateHold(ctx context.Context, hold *persist.BalanceItem) error {

	item := balanceItemDocument{
		Version:   0,
		ID:        hold.ID,
		Timestamp: time.Time(hold.Timestamp),
		Created:   time.Now(),
		Amount:    hold.Amount.StringFixedBank(b.symbol.RoundingPlace()),
	}

	col := fmt.Sprintf("accounts/%s/symbols/%s/holds", b.account.ID, b.symbol)
	_, _, err := b.getClient(ctx).Collection(col).Add(ctx, &item)

	return err
}

func (b *BalanceRepository) UpdateHold(ctx context.Context, id persist.Key, amt decimal.Decimal) error {

	_, docs, err := b.getBalanceItemDocuments(ctx, b.getHoldCollection(ctx))
	if err != nil {
		return err
	}

	for _, doc := range docs {
		if doc.ID == id.String() {
			// insert new version of item
			item := balanceItemDocument{
				Version:   doc.Version + 1,
				ID:        doc.ID,
				Timestamp: doc.Timestamp,
				Created:   time.Now(),
				Amount:    amt.StringFixedBank(b.symbol.RoundingPlace()),
			}

			col := fmt.Sprintf("accounts/%s/symbols/%s/holds", b.account.ID, b.symbol)
			_, _, err := b.getClient(ctx).Collection(col).Add(ctx, &item)
			if err != nil {
				return err
			}

			return nil
		}
	}

	return fmt.Errorf("%w for account %s and id %s", ErrHoldNotFound, b.account.ID, id)
}

func (b *BalanceRepository) DeleteHold(ctx context.Context, id persist.Key) error {

	_, docs, err := b.getBalanceItemDocuments(ctx, b.getHoldCollection(ctx))
	if err != nil {
		return err
	}

	item := balanceItemDocument{
		Version:   99999,
		ID:        id.String(),
		Timestamp: time.Now(),
		Created:   time.Now(),
		Amount:    "0",
	}

	for _, doc := range docs {
		if doc.ID == id.String() {
			// insert new version of item as a zero amount item
			// this will allow it to be deleted later
			item.Version = doc.Version + 1
			item.Timestamp = doc.Timestamp

			break
		}
	}

	col := fmt.Sprintf("accounts/%s/symbols/%s/holds", b.account.ID, b.symbol)
	_, _, err = b.getClient(ctx).Collection(col).Add(ctx, &item)
	if err != nil {
		return err
	}

	return nil
}

func (b *BalanceRepository) getClient(ctx context.Context) *firestore.Client {

	var client *firestore.Client
	if b.client == nil {
		client = clientFromContext(ctx)
	} else {
		client = b.client
	}
	return client
}
