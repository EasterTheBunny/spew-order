package types

import (
	"time"

	uuid "github.com/satori/go.uuid"
)

// OrderRequest represents an incoming order request.
type OrderRequest struct {
	Base     Symbol     `json:"base"`
	Target   Symbol     `json:"target"`
	Price    float64    `json:"price"`
	Quantity float64    `json:"quantity"`
	Action   ActionType `json:"action"`
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
