package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"

	"github.com/shopspring/decimal"
)

// Symbol ...
type Symbol int

const (
	// SymbolBitcoin ...
	SymbolBitcoin Symbol = 2
	// SymbolEthereum ...
	SymbolEthereum Symbol = 4
	// SymbolBitcoinCash ...
	SymbolBitcoinCash Symbol = 8
	// SymbolDogecoin ...
	SymbolDogecoin Symbol = 16
	// SymbolUniswap ...
	SymbolUniswap Symbol = 20
)

const (
	symbolBitcoinName     = "BTC"
	symbolEthereumName    = "ETH"
	symbolBitcoinCashName = "BCH"
	symbolDogecoinName    = "DOGE"
	symbolUniswapName     = "UNI"
)

var (
	// ErrSymbolUnrecognized describes an error state where a provided Symbol
	// is not in the list of options provided by this package.
	ErrSymbolUnrecognized = errors.New("unrecognized symbol")
	ValidPairs            = []string{
		fmt.Sprintf("%s%s", symbolBitcoinName, symbolEthereumName),
		fmt.Sprintf("%s%s", symbolBitcoinName, symbolBitcoinCashName),
		// fmt.Sprintf("%s%s", symbolEthereumName, symbolBitcoinCashName),
		fmt.Sprintf("%s%s", symbolBitcoinName, symbolDogecoinName),
		// fmt.Sprintf("%s%s", symbolEthereumName, symbolDogecoinName),
		fmt.Sprintf("%s%s", symbolBitcoinName, symbolUniswapName),
	}
)

// String provides a string representation to an Symbol value. Defaults to
// empty string if value is unrecognized.
func (s Symbol) String() string {
	switch s {
	case SymbolBitcoin:
		return symbolBitcoinName
	case SymbolEthereum:
		return symbolEthereumName
	case SymbolBitcoinCash:
		return symbolBitcoinCashName
	case SymbolDogecoin:
		return symbolDogecoinName
	case SymbolUniswap:
		return symbolUniswapName
	default:
		return ""
	}
}

func (s Symbol) typeInRange() bool {
	return s >= SymbolBitcoin && s <= SymbolUniswap
}

// RoundingPlace provides expected rounding values for each symbol
func (s Symbol) RoundingPlace() int32 {
	switch s {
	case SymbolBitcoin, SymbolBitcoinCash, SymbolDogecoin:
		return 8
	case SymbolEthereum, SymbolUniswap:
		return 18
	default:
		return 8
	}
}

func (s Symbol) MinimumFee() decimal.Decimal {
	switch s {
	case SymbolBitcoin, SymbolBitcoinCash, SymbolDogecoin:
		return decimal.NewFromFloat(0.00000001)
	case SymbolEthereum, SymbolUniswap:
		return decimal.NewFromFloat(0.000000000000000001)
	default:
		return decimal.NewFromInt(0)
	}
}

// ValidateAddress checks that an address for a given symbol is a valid sending address
// supported addresses on Ethereum include EIP55
func (s Symbol) ValidateAddress(a string) bool {

	switch s {
	case SymbolBitcoin, SymbolBitcoinCash, SymbolDogecoin:
		// A Bitcoin address is between 25 and 34 characters long;
		if len(a) < 25 || len(a) > 34 {
			return false
		}

		// the address always starts with a 1;
		if string(a[0]) != "1" {
			return false
		}

		// an address can contain all alphanumeric characters, with the exceptions of 0, O, I, and l.
		if exceptionLetters.MatchString(a) {
			return false
		}
	case SymbolEthereum, SymbolUniswap:
		return validateEIP55(a)
	}

	return false
}

var (
	lenCheck         = regexp.MustCompile(`^(?P<Pref>0x)(?P<Addr>[0-9a-fA-F]{40})$`)
	capsCheck1       = regexp.MustCompile(`^(?P<Pref>0x)(?P<Addr>[0-9a-f]{40})$`)
	capsCheck2       = regexp.MustCompile(`^(?P<Pref>0x)(?P<Addr>[0-9A-F]{40})$`)
	exceptionLetters = regexp.MustCompile(`[0OIl]`)
)

// Checks if the given string is an address
func validateEIP55(s string) bool {

	if !lenCheck.MatchString(s) {
		// check if it has the basic requirements of an address
		return false
	}

	if capsCheck1.MatchString(s) || capsCheck2.MatchString(s) {
		// If it's all small caps or all caps, return true
		return true
	}

	/*
		// Checks if the given string is a checksummed address
		s = strings.ReplaceAll(s, `0x`, "")

		hash := sha3.NewLegacyKeccak256()
		_, err := io.WriteString(hash, strings.ToLower(s))
		if err != nil {
			return false
		}

		sm := hash.Sum(nil)
		fmt.Println(sm)

		if len(sm) < 40 {
			return false
		}

		for i := 0; i < 40; i++ {
			// the nth letter should be uppercase if the nth digit of casemap is 1
			nt, err := strconv.ParseInt(string(sm[i]), 16, 64)
			if err != nil {
				return false
			}

			str := string(s[i])
			upper := strings.ToUpper(str)
			lower := strings.ToLower(str)

			if (nt > 7 && upper != str) || (nt <= 7 && lower != str) {
				return false
			}
		}
	*/

	return true
}

// MarshalBinary ...
func (s Symbol) MarshalBinary() ([]byte, error) {
	return []byte(fmt.Sprintf("%d", int(s))), nil
}

// UnmarshalBinary ...
func (s *Symbol) UnmarshalBinary(b []byte) error {
	val, err := strconv.ParseInt(string(b), 10, 32)
	if err != nil {
		return err
	}

	reflect.ValueOf(s).Elem().Set(reflect.ValueOf(int(val)))
	return nil
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

	sym, err := FromString(str)
	if err != nil {
		return err
	}

	*s = sym

	return nil
}

func FromString(str string) (Symbol, error) {
	switch str {
	case symbolBitcoinName:
		return SymbolBitcoin, nil
	case symbolEthereumName:
		return SymbolEthereum, nil
	case symbolBitcoinCashName:
		return SymbolBitcoinCash, nil
	case symbolDogecoinName:
		return SymbolDogecoin, nil
	case symbolUniswapName:
		return SymbolUniswap, nil
	default:
		return 0, ErrSymbolUnrecognized
	}
}
