package persist

import (
	"bytes"
	"context"
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
	ErrCannotParseValue   = errors.New("datastore collection parse error")
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
	Find(context.Context, Key) (*Account, error)
	FindByAddress(context.Context, string, types.Symbol) (*Account, error)
	Save(context.Context, *Account) error
	Balances(*Account, types.Symbol) BalanceRepository
	Transactions(*Account) TransactionRepository
	Orders(*Account) OrderRepository
}

// Account represents the entity object persisted to storage
type Account struct {
	ID        string           `json:"id"`
	Addresses []FundingAddress `json:"addresses"`
}

type FundingAddress struct {
	Symbol  types.Symbol `json:"symbol"`
	Address string       `json:"address"`
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
	DeleteHold(Key) error
	UpdateHold(Key, decimal.Decimal) error
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
	GetAuthorizations() ([]*Authorization, error)
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

// NewAuthorization returns a new auth with values set to defaults and a new
// id generated.
func NewAuthorization(accts ...Account) *Authorization {
	var ids []string
	for _, a := range accts {
		ids = append(ids, a.ID)
	}

	return &Authorization{
		ID:       uuid.NewV4().String(),
		Accounts: ids}
}

func (a Authorization) Encode(enc EncodingType) ([]byte, error) {
	return encode(enc, a)
}

func (a *Authorization) Decode(b []byte, enc EncodingType) error {
	return decode(b, enc, a)
}

type BookRepository interface {
	SetBookItem(*BookItem) error
	BookItemExists(*BookItem) (bool, error)
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

type OrderRepository interface {
	GetOrder(Key) (*Order, error)
	SetOrder(*Order) error
	GetOrdersByStatus(...FillStatus) ([]*Order, error)
	UpdateOrderStatus(Key, FillStatus, []string) error
}

type Order struct {
	Status       FillStatus  `json:"status"`
	Transactions [][]string  `json:"transactions"`
	Base         types.Order `json:"base"`
}

func (o Order) Encode(enc EncodingType) ([]byte, error) {
	return encode(enc, o)
}

func (o *Order) Decode(b []byte, enc EncodingType) error {
	return decode(b, enc, o)
}

type Transaction struct {
	Type            TransactionType
	AddressHash     string
	TransactionHash string
	OrderID         string
	Symbol          string
	Quantity        string
	Fee             string
	Timestamp       NanoTime
}

type TransactionType string

const (
	OrderTransactionType    = "order"
	DepositTransactionType  = "deposit"
	TransferTransactionType = "transfer"
)

type TransactionRepository interface {
	SetTransaction(*Transaction) error
	GetTransactions() ([]*Transaction, error)
}

func (t Transaction) Encode(enc EncodingType) ([]byte, error) {
	return encode(enc, t)
}

func (t *Transaction) Decode(b []byte, enc EncodingType) error {
	return decode(b, enc, t)
}

type AccountType int

const (
	Liability AccountType = iota
	Asset
)

type LedgerAccount int

const (
	Cash LedgerAccount = iota
	Sales
	TransfersPayable
	Transfers
)

type EntryType int

const (
	Credit EntryType = iota
	Debit
)

type LedgerEntry struct {
	Account   LedgerAccount   `json:"account"`
	Entry     EntryType       `json:"entry"`
	Symbol    types.Symbol    `json:"symbol"`
	Amount    decimal.Decimal `json:"amount"`
	Timestamp NanoTime        `json:"timestamp"`
}

func (e LedgerEntry) Encode(enc EncodingType) ([]byte, error) {
	return encode(enc, e)
}

func (e *LedgerEntry) Decode(b []byte, enc EncodingType) error {
	return decode(b, enc, e)
}

type LedgerRepository interface {
	// RecordDeposit saves a transfer to the exchange in the main ledger
	RecordDeposit(types.Symbol, decimal.Decimal) error
	// RecordTransfer saves a transfer from the exchange in the main ledger
	RecordTransfer(types.Symbol, decimal.Decimal) error
	// GetLiabilityBalance ...
	GetLiabilityBalance(a LedgerAccount) (balances map[types.Symbol]decimal.Decimal, err error)
	// GetAssetBalance ...
	GetAssetBalance(a LedgerAccount) (balances map[types.Symbol]decimal.Decimal, err error)
	// RecordFee saves a fee paid from a completed order in the main ledger
	RecordFee(types.Symbol, decimal.Decimal) error
}

type FillStatus int

const (
	StatusOpen FillStatus = iota
	StatusPartial
	StatusFilled
	StatusCanceled
)

func (s FillStatus) MarshalBinary() ([]byte, error) {
	return s.MarshalJSON()
}

func (s *FillStatus) UnmarshalBinary(b []byte) error {
	return s.UnmarshalJSON(b)
}

func (s FillStatus) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%d", s)), nil
}

func (s *FillStatus) UnmarshalJSON(b []byte) error {
	val, err := strconv.ParseInt(string(b), 10, 64)
	if err != nil {
		return err
	}

	switch int(val) {
	case 0:
		*s = StatusOpen
	case 1:
		*s = StatusPartial
	case 2:
		*s = StatusFilled
	case 3:
		*s = StatusCanceled
	}

	return nil
}

type NanoTime time.Time

func (t NanoTime) Value() int64 {
	return time.Time(t).UnixNano()
}

func (t NanoTime) String() string {
	return strconv.FormatInt(time.Time(t).UnixNano(), 10)
}

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
