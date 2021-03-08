package types

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Symbol ...
type Symbol uint

const (
	// SymbolBitcoin ...
	SymbolBitcoin Symbol = iota
	// SymbolEthereum ...
	SymbolEthereum
)

const (
	symbolBitcoinName  = "BTC"
	symbolEthereumName = "ETH"
)

var (
	// ErrSymbolUnrecognized describes an error state where a provided Symbol
	// is not in the list of options provided by this package.
	ErrSymbolUnrecognized = errors.New("unrecognized symbol")
)

// String provides a string representation to an Symbol value. Defaults to
// empty string if value is unrecognized.
func (s Symbol) String() string {
	names := [...]string{
		symbolBitcoinName,
		symbolEthereumName}

	// default to blank string
	if !s.typeInRange() {
		return ""
	}
	return names[s]
}

func (s Symbol) typeInRange() bool {
	return s >= SymbolBitcoin && s <= SymbolEthereum
}

// MarshalJSON ...
func (s Symbol) MarshalJSON() ([]byte, error) {
	if !s.typeInRange() {
		return []byte(`""`), ErrSymbolUnrecognized
	}

	return []byte(fmt.Sprintf(`"%s"`, s.String())), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface. This implementation returns
// an error if the Sybol is not within the range of the values defined
// in this package.
func (s *Symbol) UnmarshalJSON(b []byte) error {
	var str string
	err := json.Unmarshal(b, &str)
	if err != nil {
		return err
	}

	switch str {
	case symbolBitcoinName:
		*s = SymbolBitcoin
	case symbolEthereumName:
		*s = SymbolEthereum
	default:
		return ErrSymbolUnrecognized
	}

	return nil
}
