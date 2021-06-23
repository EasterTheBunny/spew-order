package firebase

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/pkg/types"
	"github.com/shopspring/decimal"
)

type LedgerRepository struct {
	client *firestore.Client
}

func NewLedgerRepository(client *firestore.Client) *LedgerRepository {
	return &LedgerRepository{client: client}
}

func (r *LedgerRepository) RecordDeposit(ctx context.Context, s types.Symbol, amt decimal.Decimal) error {
	return nil
}

func (r *LedgerRepository) RecordTransfer(ctx context.Context, s types.Symbol, amt decimal.Decimal) error {
	return nil
}

func (r *LedgerRepository) GetLiabilityBalance(ctx context.Context, a persist.LedgerAccount) (balances map[types.Symbol]decimal.Decimal, err error) {
	return
}

func (r *LedgerRepository) GetAssetBalance(ctx context.Context, a persist.LedgerAccount) (balances map[types.Symbol]decimal.Decimal, err error) {
	return
}

func (r *LedgerRepository) RecordFee(ctx context.Context, s types.Symbol, amt decimal.Decimal) error {
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
