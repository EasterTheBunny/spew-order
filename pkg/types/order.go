package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
)

// OrderRequest represents an incoming order request.
type OrderRequest struct {
	Base     Symbol          `json:"base"`
	Target   Symbol          `json:"target"`
	Action   ActionType      `json:"action"`
	Type     OrderType       `json:"type"`
	Price    decimal.Decimal `json:"price"`
	Quantity decimal.Decimal `json:"quantity"`
}

// Order is the complete order representation. Built by composition of the Request.
type Order struct {
	OrderRequest
	ID        uuid.UUID `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Owner     uuid.UUID `json:"owner"`
}

// NewOrderFromRequest ...
func NewOrderFromRequest(r OrderRequest) Order {
	return Order{
		OrderRequest: r,
		ID:           uuid.NewV4(),
		Timestamp:    time.Now(),
		Owner:        uuid.NewV4()}
}

// OrderType ...
type OrderType uint

const (
	// OrderTypeMarket ...
	OrderTypeMarket OrderType = iota
	// OrderTypeLimit ...
	OrderTypeLimit
)

const (
	orderTypeMarketName = "MARKET"
	orderTypeLimitName  = "LIMIT"
)

var (
	// ErrOrderTypeUnrecognized ...
	ErrOrderTypeUnrecognized = errors.New("unrecognized order type")
)

func (o OrderType) String() string {
	names := [...]string{
		orderTypeMarketName,
		orderTypeLimitName}

	if !o.typeInRange() {
		return ""
	}
	return names[o]
}

func (o OrderType) typeInRange() bool {
	return o >= OrderTypeMarket && o <= OrderTypeLimit
}

// MarshalJSON ...
func (o OrderType) MarshalJSON() ([]byte, error) {
	if !o.typeInRange() {
		return []byte(`""`), ErrOrderTypeUnrecognized
	}

	return []byte(fmt.Sprintf(`"%s"`, o.String())), nil
}

// UnmarshalJSON ...
func (o *OrderType) UnmarshalJSON(b []byte) error {
	var str string
	err := json.Unmarshal(b, &str)
	if err != nil {
		return err
	}

	switch str {
	case orderTypeMarketName:
		*o = OrderTypeMarket
	case orderTypeLimitName:
		*o = OrderTypeLimit
	default:
		return ErrOrderTypeUnrecognized
	}

	return nil
}
