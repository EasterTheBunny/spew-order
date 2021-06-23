package firebase

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/easterthebunny/spew-order/internal/persist"
	"google.golang.org/api/iterator"
)

type TransactionRepository struct {
	client  *firestore.Client
	account *persist.Account
}

func NewTransactionRepository(client *firestore.Client, account *persist.Account) *TransactionRepository {
	return &TransactionRepository{client: client, account: account}
}

// /root/account/{accountid}/transaction/{timestamp}
func (tr *TransactionRepository) SetTransaction(ctx context.Context, t *persist.Transaction) error {
	collection := tr.getClient(ctx).Collection("accounts").Doc(tr.account.ID).Collection("transactions")
	_, _, err := collection.Add(ctx, transactionToDocument(t))
	if err != nil {
		return err
	}

	return nil
}

func (tr *TransactionRepository) GetTransactions(ctx context.Context) (t []*persist.Transaction, err error) {
	collection := tr.getClient(ctx).Collection("accounts").Doc(tr.account.ID).Collection("transactions")
	iter := collection.OrderBy("timestamp", firestore.Desc).Documents(ctx)
	var doc *firestore.DocumentSnapshot
	for {
		doc, err = iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return
		}

		t = append(t, documentToTransaction(doc.Data()))
	}
	return
}

func (tr *TransactionRepository) getClient(ctx context.Context) *firestore.Client {

	var client *firestore.Client
	if tr.client == nil {
		client = clientFromContext(ctx)
	} else {
		client = tr.client
	}
	return client
}

func transactionToDocument(tr *persist.Transaction) map[string]interface{} {

	m := map[string]interface{}{
		"type":             string(tr.Type),
		"address_hash":     tr.AddressHash,
		"transaction_hash": tr.TransactionHash,
		"order_id":         tr.OrderID,
		"symbol":           tr.Symbol,
		"quantity":         tr.Quantity,
		"fee":              tr.Fee,
		"timestamp":        tr.Timestamp.Value(),
	}

	return m
}

func documentToTransaction(m map[string]interface{}) *persist.Transaction {
	return &persist.Transaction{
		Type:            persist.TransactionType(m["type"].(string)),
		AddressHash:     m["address_hash"].(string),
		TransactionHash: m["transaction_hash"].(string),
		OrderID:         m["order_id"].(string),
		Symbol:          m["symbol"].(string),
		Quantity:        m["quantity"].(string),
		Fee:             m["fee"].(string),
		Timestamp:       persist.NanoTime(time.Unix(0, m["timestamp"].(int64))),
	}
}
