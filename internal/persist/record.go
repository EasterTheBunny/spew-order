package persist

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/easterthebunny/spew-order/pkg/types"
	uuid "github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
)

var (
	ErrCannotSaveNilValue = errors.New("cannot save nil value")
)

type Key interface {
	String() string
}

type EncodingType int

const (
	JSON EncodingType = iota
	GOB
)

const (
	JSONEncodingTypeName = "application/json"
	GOBEncodingTypeName  = "application/gob"
)

type AccountRepository interface {
	Find(Key) (*Account, error)
	Save(*Account) error
	Balances(*Account, types.Symbol) BalanceRepository
}

// Account represents the entity object persisted to storage
type Account struct {
	ID string `json:"id"`
}

// Encode marshals Account to bytes based on selected encoding type; defaults to JSON
func (a Account) Encode(enc EncodingType) ([]byte, error) {
	return encode(enc, a)
}

// Decode unmarshals JSON encoded bytes
func (a *Account) Decode(b []byte, enc EncodingType) error {
	return decode(b, enc, a)
}

type BalanceRepository interface {
	GetBalance() (decimal.Decimal, error)
	UpdateBalance(decimal.Decimal) error
	FindHolds() ([]*BalanceItem, error)
	CreateHold(*BalanceItem) error
	DeleteHold(*BalanceItem) error
	FindPosts() ([]*BalanceItem, error)
	CreatePost(*BalanceItem) error
	DeletePost(*BalanceItem) error
}

type BalanceItem struct {
	ID        string          `json:"id"`
	Timestamp NanoTime        `json:"timestamp"`
	Amount    decimal.Decimal `json:"amount"`
}

func NewBalanceItem(amt decimal.Decimal) *BalanceItem {
	return &BalanceItem{
		ID:        uuid.NewV4().String(),
		Timestamp: NanoTime(time.Now()),
		Amount:    amt}
}

// Encode marshals to JSON encoded bytes
// this was shown to be faster than gob encoding by the included benchmark test
func (bi BalanceItem) Encode(enc EncodingType) ([]byte, error) {
	return encode(enc, bi)
}

// Decode unmarshals JSON encoded bytes
func (bi *BalanceItem) Decode(b []byte, enc EncodingType) error {
	return decode(b, enc, bi)
}

// AuthorizationRepository ...
type AuthorizationRepository interface {
	GetAuthorization(Key) (*Authorization, error)
	SetAuthorization(*Authorization) error
	DeleteAuthorization(*Authorization) error
}

// Authorization ...
type Authorization struct {
	ID       string   `json:"id"`
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Name     string   `json:"name"`
	Avatar   string   `json:"avatar"`
	Accounts []string `json:"accounts"`
}

func (a Authorization) Encode(enc EncodingType) ([]byte, error) {
	return encode(enc, a)
}

func (a *Authorization) Decode(b []byte, enc EncodingType) error {
	return decode(b, enc, a)
}

type BookRepository interface {
	SetBookItem(*BookItem) error
	GetHeadBatch(*BookItem, int) ([]*BookItem, error)
	DeleteBookItem(*BookItem) error
}

// BookItem is a struct for holding an order in storage
type BookItem struct {
	Timestamp  NanoTime         `json:"timestamp"`
	Order      types.Order      `json:"order"`
	ActionType types.ActionType `json:"action_type"`
}

// NewBookItem returns a new BookItem where the meta data for range queries
// includes the order Quantity and Timestamp
func NewBookItem(order types.Order) BookItem {
	// the action type will be used to search through the opposite sorted list
	var tp types.ActionType
	if order.Action == types.ActionTypeBuy {
		tp = types.ActionTypeSell
	} else {
		tp = types.ActionTypeBuy
	}

	return BookItem{
		Timestamp:  NanoTime(order.Timestamp),
		Order:      order,
		ActionType: tp}
}

func (bi BookItem) Encode(enc EncodingType) ([]byte, error) {
	return encode(enc, bi)
}

func (bi *BookItem) Decode(b []byte, enc EncodingType) error {
	return decode(b, enc, bi)
}

type NanoTime time.Time

func (t NanoTime) MarshalBinary() ([]byte, error) {
	return t.MarshalJSON()
}

func (t *NanoTime) UnmarshalBinary(b []byte) error {
	return t.UnmarshalJSON(b)
}

func (t NanoTime) MarshalJSON() ([]byte, error) {
	stamp := fmt.Sprintf("%d", time.Time(t).UnixNano())
	return []byte(stamp), nil
}

func (t *NanoTime) UnmarshalJSON(b []byte) error {
	val, err := strconv.ParseInt(string(b), 10, 64)
	if err != nil {
		return err
	}

	jt := NanoTime(time.Unix(0, val))
	reflect.ValueOf(t).Elem().Set(reflect.ValueOf(jt))
	return nil
}

func encode(enc EncodingType, val interface{}) ([]byte, error) {
	switch enc {
	case GOB:
		var buf bytes.Buffer
		err := gob.NewEncoder(&buf).Encode(val)
		return buf.Bytes(), err
	case JSON:
		return json.Marshal(val)
	default:
		return json.Marshal(val)
	}
}

func decode(b []byte, enc EncodingType, val interface{}) error {
	switch enc {
	case GOB:
		return gob.NewDecoder(bytes.NewBuffer(b)).Decode(val)
	case JSON:
		return json.Unmarshal(b, val)
	default:
		return json.Unmarshal(b, val)
	}
}
