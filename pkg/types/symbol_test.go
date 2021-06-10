package types

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSymbolMarshalJSON(t *testing.T) {
	cases := []Symbol{
		SymbolBitcoin,
		SymbolEthereum}

	expect := []string{
		`"BTC"`,
		`"ETH"`}

	for i, c := range cases {
		exp := expect[i]

		// test built in go marshaling
		result, err := json.Marshal(c)
		if err != nil {
			t.Errorf("unexpected error occurred: %s", err.Error())
		}

		if string(result) != exp {
			t.Errorf("result did not match: %s; expected: %s", string(result), exp)
		}
	}

	var notValidSymbol Symbol = 100000000
	r, err := notValidSymbol.MarshalJSON()

	if err != ErrSymbolUnrecognized {
		t.Errorf("error expected: symbol is not in the valid set; received %v", err)
	}

	if string(r) != `""` {
		t.Errorf(`unexpected return value %s for invalid symbol: expected ""`, string(r))
	}
}

func TestSymbolUnmarshalJSON(t *testing.T) {
	cases := []string{
		`"BTC"`,
		`"ETH"`}

	expect := []Symbol{
		SymbolBitcoin,
		SymbolEthereum}

	for i, c := range cases {
		exp := expect[i]

		// test built in go marshaling
		var result Symbol
		err := json.Unmarshal([]byte(c), &result)
		if err != nil {
			t.Errorf("unexpected error occurred: %s", err.Error())
		}

		if result != exp {
			t.Errorf("result did not match: %s; expected: %s", result, exp)
		}
	}

	invalid := `"INVALIDTYPE"`
	var r Symbol
	err := r.UnmarshalJSON([]byte(invalid))

	if err != ErrSymbolUnrecognized {
		t.Errorf("error expected: symbol is not in the valid set; received %v", err)
	}
}

func TestValidateAddress(t *testing.T) {
	type testCase struct {
		s        Symbol
		h        string
		expected bool
	}

	tests := []testCase{
		{SymbolEthereum, "0x03a03cDE317214414fd314fA5105C78f1f342a15", true},
	}

	for _, test := range tests {
		result := test.s.ValidateAddress(test.h)
		assert.Equal(t, test.expected, result, fmt.Sprintf("%s: %s", test.s.String(), test.h))
	}
}
