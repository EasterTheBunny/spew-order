package firebase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"cloud.google.com/go/firestore"
	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/pkg/types"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// /root/account/{accountid}
// /root/addresses/{symbol}/{addr}
type AccountRepository struct {
	client *firestore.Client
}

func NewAccountRepository(client *firestore.Client) *AccountRepository {
	return &AccountRepository{client: client}
}

func (r *AccountRepository) Find(ctx context.Context, id persist.Key) (account *persist.Account, err error) {

	dsnap, err := r.getClient(ctx).Collection("accounts").Doc(id.String()).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, persist.ErrObjectNotExist
		}

		return nil, err
	}

	account = documentToAccount(dsnap.Data())
	return
}

func (r *AccountRepository) FindByAddress(ctx context.Context, addr string, sym types.Symbol) (acct *persist.Account, err error) {
	iter := r.getClient(ctx).Collection("addresses").Where("symbol", "==", sym.String()).Where("address", "==", addr).Documents(ctx)
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

		m := doc.Data()
		if v, ok := m["account"]; ok {
			return r.Find(ctx, ky(v.(string)))
		} else {
			err = errors.New("incorrect value for account id")
			break
		}
	}
	return
}

func (r *AccountRepository) Save(ctx context.Context, account *persist.Account) error {
	if account == nil {
		return fmt.Errorf("%w for account", persist.ErrCannotSaveNilValue)
	}

	var err error
	for _, addr := range account.Addresses {
		doc := map[string]interface{}{
			"symbol":  addr.Symbol.String(),
			"address": addr.Address,
			"account": account.ID,
		}

		_, _, err := r.getClient(ctx).Collection("addresses").Add(ctx, doc)
		if err != nil {
			return err
		}
	}

	_, err = r.getClient(ctx).Collection("accounts").Doc(account.ID).Set(ctx, accountToDocument(account))
	if err != nil {
		return err
	}

	return nil
}

func (r *AccountRepository) Balances(a *persist.Account, s types.Symbol) persist.BalanceRepository {
	return NewBalanceRepository(r.client, a, s)
}

func (r *AccountRepository) Transactions(a *persist.Account) persist.TransactionRepository {
	return NewTransactionRepository(r.client, a)
}

func (r *AccountRepository) Orders(a *persist.Account) persist.OrderRepository {
	return NewOrderRepository(r.client, a)
}

type ky string

func (k ky) String() string {
	return string(k)
}

func accountToDocument(a *persist.Account) map[string]interface{} {
	m := map[string]interface{}{
		"id": a.ID,
	}

	addr := []interface{}{}
	for _, v := range a.Addresses {
		addr = append(addr, map[string]interface{}{
			"symbol":  v.Symbol.String(),
			"address": v.Address,
		})
	}

	m["addresses"] = addr

	return m
}

func documentToAccount(m map[string]interface{}) *persist.Account {
	acct := &persist.Account{}

	if v, ok := m["id"]; ok {
		acct.ID = v.(string)
	}

	if v, ok := m["addresses"]; ok {
		addrs := []persist.FundingAddress{}

		for _, a := range v.([]interface{}) {
			vals := a.(map[string]interface{})
			addr := persist.FundingAddress{}

			if s, ok := vals["symbol"]; ok {
				json.Unmarshal([]byte(`"`+s.(string)+`"`), &addr.Symbol)
			}

			if s, ok := vals["address"]; ok {
				addr.Address = s.(string)
			}

			addrs = append(addrs, addr)
		}

		acct.Addresses = addrs
	}

	return acct
}

func (r *AccountRepository) getClient(ctx context.Context) *firestore.Client {

	var client *firestore.Client
	if r.client == nil {
		client = clientFromContext(ctx)
	} else {
		client = r.client
	}
	return client
}

func clientFromContext(ctx context.Context) *firestore.Client {
	v := ctx.Value(ClientContextKey)
	if val, ok := v.(*firestore.Client); ok {
		return val
	}

	return nil
}

type contextKey int

const (
	ClientContextKey contextKey = iota
)
