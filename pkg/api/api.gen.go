// Package api provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen DO NOT EDIT.
package api

// Defines values for ActionType.
const (
	ActionTypeBUY ActionType = "BUY"

	ActionTypeSELL ActionType = "SELL"
)

// Defines values for OrderTypeName.
const (
	OrderTypeNameLIMIT OrderTypeName = "LIMIT"

	OrderTypeNameMARKET OrderTypeName = "MARKET"
)

// Defines values for SymbolType.
const (
	SymbolTypeBTC SymbolType = "BTC"

	SymbolTypeETH SymbolType = "ETH"
)

// Balances account
type Account struct {
	Id string `json:"id"`
}

// Action type: * `BUY` - use base currency to buy target currency * `SELL` - sell target currency for base currency
type ActionType string

// BookOrder defines model for BookOrder.
type BookOrder struct {
	Guid string `json:"guid"`
}

// CurrencyValue defines model for CurrencyValue.
type CurrencyValue string

// LimitOrderRequest defines model for LimitOrderRequest.
type LimitOrderRequest struct {
	// Embedded struct due to allOf(#/components/schemas/OrderType)
	OrderType `yaml:",inline"`
	// Embedded fields due to inline allOf schema

	// Symbol Type: * `BTC` - bitcoin currency identifier * `ETH` - ethereum currency identifier
	Base     SymbolType    `json:"base"`
	Price    CurrencyValue `json:"price"`
	Quantity CurrencyValue `json:"quantity"`
}

// MarketOrderRequest defines model for MarketOrderRequest.
type MarketOrderRequest struct {
	// Embedded struct due to allOf(#/components/schemas/OrderType)
	OrderType `yaml:",inline"`
	// Embedded fields due to inline allOf schema

	// Symbol Type: * `BTC` - bitcoin currency identifier * `ETH` - ethereum currency identifier
	Base     SymbolType    `json:"base"`
	Quantity CurrencyValue `json:"quantity"`
}

// Request to create a new order on the order book
type OrderRequest struct {

	// Action type: * `BUY` - use base currency to buy target currency * `SELL` - sell target currency for base currency
	Action ActionType `json:"action"`

	// Symbol Type: * `BTC` - bitcoin currency identifier * `ETH` - ethereum currency identifier
	Base SymbolType `json:"base"`

	// Symbol Type: * `BTC` - bitcoin currency identifier * `ETH` - ethereum currency identifier
	Target SymbolType       `json:"target"`
	Type   OrderRequestType `json:"type"`
}

// OrderRequestType defines model for OrderRequestType.
type OrderRequestType interface{}

// OrderType defines model for OrderType.
type OrderType struct {

	// Order type: * `MARKET` - order type used to buy or sell at market value * `LIMIT` - used to set buy or sell limit
	Name OrderTypeName `json:"name"`
}

// Order type: * `MARKET` - order type used to buy or sell at market value * `LIMIT` - used to set buy or sell limit
type OrderTypeName string

// ResponseError defines model for ResponseError.
type ResponseError struct {
	Detail string `json:"detail"`
}

// Symbol Type: * `BTC` - bitcoin currency identifier * `ETH` - ethereum currency identifier
type SymbolType string

// AccountPathParam defines model for AccountPathParam.
type AccountPathParam string

// PostAccountAccountIDOrderJSONBody defines parameters for PostAccountAccountIDOrder.
type PostAccountAccountIDOrderJSONBody OrderRequest

// PostAccountAccountIDOrderJSONRequestBody defines body for PostAccountAccountIDOrder for application/json ContentType.
type PostAccountAccountIDOrderJSONRequestBody PostAccountAccountIDOrderJSONBody
