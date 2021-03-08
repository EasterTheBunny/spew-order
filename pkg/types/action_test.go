package types

import (
	"encoding/json"
	"testing"
)

func TestActionTypeMarshalJSON(t *testing.T) {
	cases := []ActionType{
		ActionTypeBuy,
		ActionTypeSell}

	expect := []string{
		`"BUY"`,
		`"SELL"`}

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

	var notValidActionType ActionType = 100000000
	r, err := notValidActionType.MarshalJSON()

	if err != ErrActionTypeUnrecognized {
		t.Errorf("error expected: action type is not in the valid set; received %v", err)
	}

	if string(r) != `""` {
		t.Errorf(`unexpected return value %s for invalid action type: expected ""`, string(r))
	}
}

func TestActionTypeUnmarshalJSON(t *testing.T) {
	cases := []string{
		`"BUY"`,
		`"SELL"`}

	expect := []ActionType{
		ActionTypeBuy,
		ActionTypeSell}

	for i, c := range cases {
		exp := expect[i]

		// test built in go marshaling
		var result ActionType
		err := json.Unmarshal([]byte(c), &result)
		if err != nil {
			t.Errorf("unexpected error occurred: %s", err.Error())
		}

		if result != exp {
			t.Errorf("result did not match: %s; expected: %s", result, exp)
		}
	}

	invalid := `"INVALIDTYPE"`
	var r ActionType
	err := r.UnmarshalJSON([]byte(invalid))

	if err != ErrActionTypeUnrecognized {
		t.Errorf("error expected: action type is not in the valid set; received %v", err)
	}
}
