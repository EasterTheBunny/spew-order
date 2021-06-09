// Package api provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.7.1 DO NOT EDIT.
package api

// Defines values for ActionType.
const (
	ActionTypeBUY ActionType = "BUY"

	ActionTypeSELL ActionType = "SELL"
)

// Defines values for OrderStatus.
const (
	OrderStatusCANCELLED OrderStatus = "CANCELLED"

	OrderStatusFILLED OrderStatus = "FILLED"

	OrderStatusOPEN OrderStatus = "OPEN"

	OrderStatusPARTIAL OrderStatus = "PARTIAL"
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

// Defines values for TransactionType.
const (
	TransactionTypeDEPOSIT TransactionType = "DEPOSIT"

	TransactionTypeORDER TransactionType = "ORDER"

	TransactionTypeTRANSFER TransactionType = "TRANSFER"
)

// Balances account
type Account struct {
	Balances *BalanceList `json:"balances,omitempty"`
	Id       string       `json:"id"`
}

// Action type: * `BUY` - use base currency to buy target currency * `SELL` - sell target currency for base currency
type ActionType string

// BalanceItem defines model for BalanceItem.
type BalanceItem struct {

	// Address hash for funding this balance
	Funding  string        `json:"funding"`
	Quantity CurrencyValue `json:"quantity"`

	// Symbol Type: * `BTC` - bitcoin currency identifier * `ETH` - ethereum currency identifier
	Symbol SymbolType `json:"symbol"`
}

// BalanceList defines model for BalanceList.
type BalanceList []BalanceItem

// BookOrder defines model for BookOrder.
type BookOrder struct {
	Guid string `json:"guid"`

	// Request to create a new order on the order book
	Order OrderRequest `json:"order"`

	// Symbol Type: * `OPEN` - incomplete order * `PARTIAL` - partial order * `FILLED` - filled order * `CANCELLED` - cancelled order
	Status OrderStatus `json:"status"`
}

// BookOrderList defines model for BookOrderList.
type BookOrderList []BookOrder

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

// Symbol Type: * `OPEN` - incomplete order * `PARTIAL` - partial order * `FILLED` - filled order * `CANCELLED` - cancelled order
type OrderStatus string

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

// Account balance change
type Transaction struct {
	Fee      CurrencyValue `json:"fee"`
	Orderid  string        `json:"orderid"`
	Quantity CurrencyValue `json:"quantity"`

	// Symbol Type: * `BTC` - bitcoin currency identifier * `ETH` - ethereum currency identifier
	Symbol    SymbolType `json:"symbol"`
	Timestamp string     `json:"timestamp"`

	// Transaction Type: * `ORDER` - transaction resulting from a match on the order book * `DEPOSIT` - transaction resulting from a funding deposit * `TRANSFER` - transaction resulting from a funding withdrawal
	Type TransactionType `json:"type"`
}

// TransactionList defines model for TransactionList.
type TransactionList []Transaction

// Transaction Type: * `ORDER` - transaction resulting from a match on the order book * `DEPOSIT` - transaction resulting from a funding deposit * `TRANSFER` - transaction resulting from a funding withdrawal
type TransactionType string

// AccountPathParam defines model for AccountPathParam.
type AccountPathParam string

// OrderPathParam defines model for OrderPathParam.
type OrderPathParam string

// PostApiAccountsAccountIDOrdersJSONBody defines parameters for PostApiAccountsAccountIDOrders.
type PostApiAccountsAccountIDOrdersJSONBody OrderRequest

// PostApiAccountsAccountIDOrdersJSONRequestBody defines body for PostApiAccountsAccountIDOrders for application/json ContentType.
type PostApiAccountsAccountIDOrdersJSONRequestBody PostApiAccountsAccountIDOrdersJSONBody
