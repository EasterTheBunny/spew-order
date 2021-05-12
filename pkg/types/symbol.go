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
	SymbolBitcoin Symbol = 2
	// SymbolEthereum ...
	SymbolEthereum Symbol = 4
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
	switch s {
	case SymbolBitcoin:
		return symbolBitcoinName
	case SymbolEthereum:
		return symbolEthereumName
	default:
		return ""
	}
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

func (s Symbol) RoundingPlace() int32 {
	switch s {
	case SymbolBitcoin:
		return 8
	case SymbolEthereum:
		return 18
	default:
		return 8
	}
}
