package types

import (
	"bytes"
	"encoding/gob"
	"testing"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestGobEncodeAccount(t *testing.T) {
	expected := Account{
		ID: uuid.NewV4()}

	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	dec := gob.NewDecoder(&b)

	// Encode (send) some values.
	err := enc.Encode(expected)
	if err != nil {
		t.Fatalf("encode error: %s", err)
	}

	// Decode (receive) and print the values.
	var item Account
	err = dec.Decode(&item)
	if err != nil {
		t.Fatalf("decode error: %s", err)
	}

	assert.Equal(t, expected.ID.String(), item.ID.String())
}

func TestGobEncodeBalanceItem(t *testing.T) {
	expected := BalanceItem{
		ID:        uuid.NewV4(),
		Timestamp: time.Now(),
		Amount:    decimal.NewFromFloat(0.0234),
	}

	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	dec := gob.NewDecoder(&b)

	// Encode (send) some values.
	err := enc.Encode(expected)
	if err != nil {
		t.Fatalf("encode error: %s", err)
	}

	// Decode (receive) and print the values.
	var item BalanceItem
	err = dec.Decode(&item)
	if err != nil {
		t.Fatalf("decode error: %s", err)
	}

	assert.Equal(t, expected.ID.String(), item.ID.String())
	if !expected.Timestamp.Equal(item.Timestamp) {
		assert.FailNow(t, "timestamp does not match")
	}
	assert.Equal(t, expected.Amount.String(), item.Amount.String())
}
