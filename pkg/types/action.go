package types

import (
	"encoding/json"
	"errors"
	"fmt"
)

// ActionType defines a set list of actions that an order can take.
type ActionType uint

const (
	// ActionTypeBuy represents a BUY order action.
	ActionTypeBuy ActionType = iota
	// ActionTypeSell represents a SELL order action.
	ActionTypeSell
)

const (
	actionTypeBuyName  = "BUY"
	actionTypeSellName = "SELL"
)

var (
	// ErrActionTypeUnrecognized describes an error state where a provided ActionType
	// is not in the list of options provided by this package.
	ErrActionTypeUnrecognized = errors.New("unrecognized action type")
)

// String provides a string representation to an ActionType value. Defaults to
// empty string if value is unrecognized.
func (at ActionType) String() string {
	names := [...]string{
		actionTypeBuyName,
		actionTypeSellName}

	// default to blank string
	if !at.typeInRange() {
		return ""
	}
	return names[at]
}

func (at ActionType) typeInRange() bool {
	return at >= ActionTypeBuy && at <= ActionTypeSell
}

// MarshalJSON implements the json.Marshaler interface. This implementation returns
// an error if the ActionType is not within the range of the values defined
// in this package.
func (at ActionType) MarshalJSON() ([]byte, error) {
	if !at.typeInRange() {
		return []byte(`""`), ErrActionTypeUnrecognized
	}

	return []byte(fmt.Sprintf(`"%s"`, at.String())), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface. This implementation returns
// an error if the ActionType is not within the range of the values defined
// in this package.
func (at *ActionType) UnmarshalJSON(b []byte) error {
	var str string
	err := json.Unmarshal(b, &str)
	if err != nil {
		return err
	}

	switch str {
	case actionTypeBuyName:
		*at = ActionTypeBuy
	case actionTypeSellName:
		*at = ActionTypeSell
	default:
		return ErrActionTypeUnrecognized
	}

	return nil
}
