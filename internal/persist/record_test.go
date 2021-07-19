package persist

import (
	"testing"
	"time"

	"github.com/easterthebunny/spew-order/pkg/types"
	uuid "github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestAccountByteEncoding(t *testing.T) {

	var b []byte
	var err error
	account := Account{
		ID: "test2",
	}

	t.Run("EncodeWithoutError", func(t *testing.T) {
		b, err = account.Encode(JSON)
		if err != nil {
			t.Fatalf("no error expected; encountered: %s", err)
		}
	})

	t.Run("DecodeFromBytes", func(t *testing.T) {
		dec := &Account{}
		err := dec.Decode(b, JSON)
		if err != nil {
			t.Fatalf("no error expected; encountered: %s", err)
		}

		assert.Equal(t, account.ID, dec.ID, "keys must be equal")
	})
}

func TestBalanceItemByteEncoding(t *testing.T) {

	var b []byte
	var err error
	bi := BalanceItem{
		ID:        "test2",
		Timestamp: NanoTime(time.Now()),
		Amount:    decimal.NewFromFloat(180.998329),
	}

	t.Run("EncodeWithoutError", func(t *testing.T) {
		b, err = bi.Encode(JSON)
		if err != nil {
			t.Fatalf("no error expected; encountered: %s", err)
		}
	})

	t.Run("DecodeFromBytes", func(t *testing.T) {
		dec := &BalanceItem{}
		err := dec.Decode(b, JSON)
		if err != nil {
			t.Fatalf("no error expected; encountered: %s", err)
		}

		assert.Equal(t, bi.ID, dec.ID, "keys must be equal")
		assert.Equal(t, time.Time(bi.Timestamp).UnixNano(), time.Time(dec.Timestamp).UnixNano(), "timestamps must be equal")
		assert.Equal(t, bi.Amount, dec.Amount, "amount must be equal")
	})
}

func TestAuthorizationByteEncoding(t *testing.T) {

	var b []byte
	var err error
	authz := Authorization{
		ID:       "test2",
		Username: "test3",
		Email:    "test4",
		Name:     "test5",
		Avatar:   "test6",
		Accounts: []string{"test7", "test8"},
	}

	t.Run("EncodeWithoutError", func(t *testing.T) {
		b, err = authz.Encode(JSON)
		if err != nil {
			t.Fatalf("no error expected; encountered: %s", err)
		}
	})

	t.Run("DecodeFromBytes", func(t *testing.T) {
		dec := &Authorization{}
		err := dec.Decode(b, JSON)
		if err != nil {
			t.Fatalf("no error expected; encountered: %s", err)
		}

		assert.Equal(t, authz.ID, dec.ID, "keys must be equal")
	})
}

func TestOrderByteEncoding(t *testing.T) {

	var b []byte
	var err error
	order := Order{
		Status: StatusFilled,
		Base: types.Order{
			OrderRequest: types.OrderRequest{
				Base:    types.SymbolBitcoin,
				Target:  types.SymbolEthereum,
				Action:  types.ActionTypeBuy,
				HoldID:  "holdid",
				Owner:   "ownerid",
				Account: uuid.NewV4(),
				Type: &types.LimitOrderType{
					Base:     types.SymbolBitcoin,
					Price:    decimal.NewFromInt(5),
					Quantity: decimal.NewFromInt(3),
				},
			},
			ID:        uuid.NewV4(),
			Timestamp: time.Now(),
		},
	}

	t.Run("EncodeWithoutError", func(t *testing.T) {
		b, err = order.Encode(JSON)
		if err != nil {
			t.Fatalf("no error expected; encountered: %s", err)
		}
	})

	t.Run("DecodeFromBytes", func(t *testing.T) {
		dec := &Order{}
		err := dec.Decode(b, JSON)
		if err != nil {
			t.Fatalf("no error expected; encountered: %s", err)
		}

		assert.Equal(t, order.Status, dec.Status, "status must be equal")
		assert.Equal(t, order.Base.ID, dec.Base.ID, "status must be equal")
	})
}

func TestFillStatusMarshalJSON(t *testing.T) {
	fs := StatusCanceled

	b, err := fs.MarshalJSON()
	if err != nil {
		t.Fatalf("no error expected; encountered: %s", err)
	}

	assert.Equal(t, `"canceled"`, string(b))
}

func TestFillStatusUnmarshalJSON(t *testing.T) {
	bts := []byte(`"filled"`)

	var fs FillStatus
	err := fs.UnmarshalJSON(bts)
	if err != nil {
		t.Fatalf("no error expected; encountered: %s", err)
	}

	assert.Equal(t, StatusFilled, fs)
}

func TestNanoTimeMarshalJSON(t *testing.T) {
	tm := NanoTime(time.Unix(0, 1000001))

	b, err := tm.MarshalJSON()
	if err != nil {
		t.Fatalf("no error expected; encountered: %s", err)
	}

	assert.Equal(t, "1000001", string(b))
}

func TestNanoTimeUnmarshalJSON(t *testing.T) {
	bts := []byte("1000001")

	var tm NanoTime
	err := tm.UnmarshalJSON(bts)
	if err != nil {
		t.Fatalf("no error expected; encountered: %s", err)
	}

	assert.Equal(t, int64(1000001), time.Time(tm).UnixNano())
}

// BenchmarkAccountByteEncoding runs a benchmark test for account byte encoding
func BenchmarkAccountByteEncoding(b *testing.B) {

	account := &Account{
		ID: "test2",
	}

	b.Run("JSON", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			b, _ := account.Encode(JSON)
			account.Decode(b, JSON)
		}
	})

	b.Run("GOB", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			b, _ := account.Encode(GOB)
			account.Decode(b, GOB)
		}
	})
}

// BenchmarkBalanceItemByteEncoding runs a benchmark test for account byte encoding
func BenchmarkBalanceItemByteEncoding(b *testing.B) {
	b.ReportAllocs()

	bi := &BalanceItem{
		ID:        "test2",
		Timestamp: NanoTime(time.Now()),
		Amount:    decimal.NewFromFloat(180.998329),
	}

	b.Run("JSON", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			b, _ := bi.Encode(JSON)
			bi.Decode(b, JSON)
		}
	})

	b.Run("GOB", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			b, _ := bi.Encode(GOB)
			bi.Decode(b, GOB)
		}
	})
}

// BenchmarkAuthorizationByteEncoding runs a benchmark test for account byte encoding
func BenchmarkAuthorizationByteEncoding(b *testing.B) {
	b.ReportAllocs()

	bi := &Authorization{
		ID:       "test2",
		Username: "test3",
		Email:    "test4",
		Name:     "test5",
		Avatar:   "test6",
		Accounts: []string{"test7", "test8"},
	}

	b.Run("JSON", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			b, _ := bi.Encode(JSON)
			bi.Decode(b, JSON)
		}
	})

	b.Run("GOB", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			b, _ := bi.Encode(GOB)
			bi.Decode(b, GOB)
		}
	})
}
