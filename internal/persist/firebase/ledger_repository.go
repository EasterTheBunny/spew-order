package firebase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/pkg/types"
	"github.com/shopspring/decimal"
	"google.golang.org/api/iterator"
)

type LedgerRepository struct {
	client *firestore.Client
}

func NewLedgerRepository(client *firestore.Client) *LedgerRepository {
	return &LedgerRepository{client: client}
}

func (r *LedgerRepository) RecordDeposit(ctx context.Context, s types.Symbol, amt decimal.Decimal) error {
	record := map[string]interface{}{
		"entry":     persist.Debit.String(),
		"account":   persist.Transfers.String(),
		"symbol":    s.String(),
		"amount":    amt.StringFixedBank(s.RoundingPlace()),
		"timestamp": time.Now().UnixNano(),
	}

	_, _, err := r.getClient(ctx).Collection(r.ledgerAccountSubspace(persist.Transfers)).Add(ctx, record)
	if err != nil {
		return err
	}

	record["entry"] = persist.Credit.String()
	record["account"] = persist.TransfersPayable.String()

	_, _, err = r.getClient(ctx).Collection(r.ledgerAccountSubspace(persist.TransfersPayable)).Add(ctx, record)
	if err != nil {
		return err
	}

	return nil
}

func (r *LedgerRepository) RecordTransfer(ctx context.Context, s types.Symbol, amt decimal.Decimal) error {
	record := map[string]interface{}{
		"entry":     persist.Credit.String(),
		"account":   persist.Transfers.String(),
		"symbol":    s.String(),
		"amount":    amt.StringFixedBank(s.RoundingPlace()),
		"timestamp": time.Now().UnixNano(),
	}

	_, _, err := r.getClient(ctx).Collection(r.ledgerAccountSubspace(persist.Transfers)).Add(ctx, record)
	if err != nil {
		return err
	}

	record["entry"] = persist.Debit.String()
	record["account"] = persist.TransfersPayable.String()

	_, _, err = r.getClient(ctx).Collection(r.ledgerAccountSubspace(persist.TransfersPayable)).Add(ctx, record)
	if err != nil {
		return err
	}

	return nil
}

func (r *LedgerRepository) GetLiabilityBalance(ctx context.Context, a persist.LedgerAccount) (balances map[types.Symbol]decimal.Decimal, err error) {
	balances = make(map[types.Symbol]decimal.Decimal)

	iter := r.getClient(ctx).Collection(r.ledgerAccountSubspace(a)).Documents(ctx)
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

		entry := documentToEntry(doc.Data())

		amt := entry.Amount
		if entry.Entry == persist.Debit {
			amt = decimal.NewFromInt(0).Sub(entry.Amount)
		}

		bal, ok := balances[entry.Symbol]
		if !ok {
			balances[entry.Symbol] = amt
		} else {
			balances[entry.Symbol] = bal.Add(amt)
		}
	}

	return
}

func (r *LedgerRepository) GetAssetBalance(ctx context.Context, a persist.LedgerAccount) (balances map[types.Symbol]decimal.Decimal, err error) {
	balances = make(map[types.Symbol]decimal.Decimal)

	iter := r.getClient(ctx).Collection(r.ledgerAccountSubspace(a)).Documents(ctx)
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

		entry := documentToEntry(doc.Data())

		amt := entry.Amount
		if entry.Entry == persist.Credit {
			amt = decimal.NewFromInt(0).Sub(entry.Amount)
		}

		bal, ok := balances[entry.Symbol]
		if !ok {
			balances[entry.Symbol] = amt
		} else {
			balances[entry.Symbol] = bal.Add(amt)
		}
	}

	return
}

func (r *LedgerRepository) RecordFee(ctx context.Context, s types.Symbol, amt decimal.Decimal) error {
	record := map[string]interface{}{
		"entry":     persist.Credit.String(),
		"account":   persist.Transfers.String(),
		"symbol":    s.String(),
		"amount":    amt.StringFixedBank(s.RoundingPlace()),
		"timestamp": time.Now().UnixNano(),
	}

	_, _, err := r.getClient(ctx).Collection(r.ledgerAccountSubspace(persist.Transfers)).Add(ctx, record)
	if err != nil {
		return err
	}

	record["entry"] = persist.Debit.String()
	record["account"] = persist.TransfersPayable.String()

	_, _, err = r.getClient(ctx).Collection(r.ledgerAccountSubspace(persist.TransfersPayable)).Add(ctx, record)
	if err != nil {
		return err
	}

	record["entry"] = persist.Debit.String()
	record["account"] = persist.Cash.String()

	_, _, err = r.getClient(ctx).Collection(r.ledgerAccountSubspace(persist.Cash)).Add(ctx, record)
	if err != nil {
		return err
	}

	record["entry"] = persist.Credit.String()
	record["account"] = persist.Sales.String()

	_, _, err = r.getClient(ctx).Collection(r.ledgerAccountSubspace(persist.Sales)).Add(ctx, record)
	if err != nil {
		return err
	}

	return nil
}

func (r *LedgerRepository) getClient(ctx context.Context) *firestore.Client {

	var client *firestore.Client
	if r.client == nil {
		client = clientFromContext(ctx)
	} else {
		client = r.client
	}
	return client
}

func (r *LedgerRepository) ledgerAccountSubspace(a persist.LedgerAccount) string {
	switch a {
	case persist.Cash:
		return fmt.Sprintf("ledger/%s/%s", persist.Asset, persist.Cash)
	case persist.Sales:
		return fmt.Sprintf("ledger/%s/%s", persist.Liability, persist.Sales)
	case persist.TransfersPayable:
		return fmt.Sprintf("ledger/%s/%s", persist.Liability, persist.TransfersPayable)
	case persist.Transfers:
		return fmt.Sprintf("ledger/%s/%s", persist.Asset, persist.Transfers)
	default:
		return "ledger"
	}
}

func documentToEntry(m map[string]interface{}) *persist.LedgerEntry {
	entry := &persist.LedgerEntry{}

	if v, ok := m["account"]; ok {
		entry.Account.FromString(v.(string))
	}

	if v, ok := m["entry"]; ok {
		entry.Entry.FromString(v.(string))
	}

	if v, ok := m["symbol"]; ok {
		json.Unmarshal([]byte(fmt.Sprintf(`"%s"`, v.(string))), &entry.Symbol)
	}

	if v, ok := m["amount"]; ok {
		amt, _ := decimal.NewFromString(v.(string))
		entry.Amount = amt
	}

	entry.Timestamp = persist.NanoTime(time.Unix(0, m["timestamp"].(int64)))

	return entry
}
