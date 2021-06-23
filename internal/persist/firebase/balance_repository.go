package firebase

import (
	"context"
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

// /root/account/{accountid}/symbol/{symbol}/balance
func (b *BalanceRepository) GetBalance(ctx context.Context) (balance decimal.Decimal, err error) {
	balance = decimal.NewFromInt(0)
	collection := b.getClient(ctx).Collection("accounts").Doc(b.account.ID).Collection("symbols")

	dsnap, err := collection.Doc(b.symbol.String()).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			err = b.UpdateBalance(ctx, balance)
		}
		return
	}

	if !dsnap.Exists() {
		err = b.UpdateBalance(ctx, balance)
		return
	}

	m := dsnap.Data()
	if v, ok := m["balance"]; ok {
		balance, err = decimal.NewFromString(v.(string))
	}

	return
}

func (b *BalanceRepository) UpdateBalance(ctx context.Context, bal decimal.Decimal) error {
	collection := b.getClient(ctx).Collection("accounts").Doc(b.account.ID).Collection("symbols")

	doc := map[string]interface{}{
		"balance": bal.StringFixedBank(b.symbol.RoundingPlace()),
	}

	_, err := collection.Doc(b.symbol.String()).Set(ctx, doc)
	if err != nil {
		return err
	}

	return nil
}

// /root/account/{accountid}/symbol/{symbol}/hold
func (b *BalanceRepository) FindHolds(ctx context.Context) (holds []*persist.BalanceItem, err error) {
	collection := b.getClient(ctx).Collection("accounts").Doc(b.account.ID).Collection("symbols").Doc(b.symbol.String()).Collection("holds")
	return b.getBalanceItems(ctx, collection)
}

// /root/account/{accountid}/symbol/{symbol}/hold/{holdid}
func (b *BalanceRepository) CreateHold(ctx context.Context, hold *persist.BalanceItem) error {
	col := fmt.Sprintf("accounts/%s/symbols/%s/holds", b.account.ID, b.symbol)
	_, err := b.getClient(ctx).Collection(col).Doc(hold.ID).Set(ctx, balanceItemToDocument(b.symbol, hold))
	if err != nil {
		return err
	}

	return nil
}

func (b *BalanceRepository) UpdateHold(ctx context.Context, id persist.Key, amt decimal.Decimal) error {
	col := fmt.Sprintf("accounts/%s/symbols/%s/holds/%s", b.account.ID, b.symbol, id)

	amtStr := amt.StringFixedBank(b.symbol.RoundingPlace())
	_, err := b.getClient(ctx).Doc(col).Set(ctx, amtStr, firestore.Merge([]string{"amount"}))
	return err
}

func (b *BalanceRepository) DeleteHold(ctx context.Context, id persist.Key) error {
	col := fmt.Sprintf("accounts/%s/symbols/%s/holds/%s", b.account.ID, b.symbol, id)
	_, err := b.getClient(ctx).Doc(col).Delete(ctx)
	return err
}

func (b *BalanceRepository) FindPosts(ctx context.Context) (posts []*persist.BalanceItem, err error) {
	collection := b.getClient(ctx).Collection("accounts").Doc(b.account.ID).Collection("symbols").Doc(b.symbol.String()).Collection("posts")
	return b.getBalanceItems(ctx, collection)
}

func (b *BalanceRepository) CreatePost(ctx context.Context, post *persist.BalanceItem) error {
	col := fmt.Sprintf("accounts/%s/symbols/%s/posts", b.account.ID, b.symbol)
	_, err := b.getClient(ctx).Collection(col).Doc(post.ID).Set(ctx, balanceItemToDocument(b.symbol, post))
	if err != nil {
		return err
	}

	return nil
}

func (b *BalanceRepository) DeletePost(ctx context.Context, post *persist.BalanceItem) error {
	col := fmt.Sprintf("accounts/%s/symbols/%s/posts/%s", b.account.ID, b.symbol, post.ID)
	_, err := b.getClient(ctx).Doc(col).Delete(ctx)
	return err
}

func (b *BalanceRepository) getBalanceItems(ctx context.Context, collection *firestore.CollectionRef) (items []*persist.BalanceItem, err error) {
	iter := collection.OrderBy("timestamp", firestore.Desc).Documents(ctx)
	var doc *firestore.DocumentSnapshot
	for {
		doc, err = iter.Next()
		if err == iterator.Done {
			err = nil
			break
		}
		if err != nil {
			return
		}

		items = append(items, documentToBalanceItem(doc.Data()))
	}
	return
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

func balanceItemToDocument(sym types.Symbol, a *persist.BalanceItem) map[string]interface{} {
	m := map[string]interface{}{
		"id":        a.ID,
		"timestamp": a.Timestamp.Value(),
		"amount":    a.Amount.StringFixedBank(sym.RoundingPlace()),
	}

	return m
}

func documentToBalanceItem(doc map[string]interface{}) *persist.BalanceItem {
	item := &persist.BalanceItem{}

	if v, ok := doc["id"]; ok {
		item.ID = v.(string)
	}

	if v, ok := doc["timestamp"]; ok {
		t := v.(int64)
		item.Timestamp = persist.NanoTime(time.Unix(0, t))
	}

	if v, ok := doc["amount"]; ok {
		amt, _ := decimal.NewFromString(v.(string))
		item.Amount = amt
	}

	return item
}
