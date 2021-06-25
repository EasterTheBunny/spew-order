package firebase

import (
	"testing"
	"time"

	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/pkg/types"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestDocumentToEntry(t *testing.T) {
	doc := map[string]interface{}{
		"account":   "sales",
		"entry":     "debit",
		"symbol":    "BTC",
		"amount":    "0.00438920",
		"timestamp": 5000000000,
	}

	expected := persist.LedgerEntry{
		Account:   persist.Sales,
		Entry:     persist.Debit,
		Symbol:    types.SymbolBitcoin,
		Amount:    decimal.NewFromFloat(0.0043892),
		Timestamp: persist.NanoTime(time.Unix(0, 5000000000)),
	}

	entry := documentToEntry(doc)

	assert.Equal(t, expected, *entry)
}
