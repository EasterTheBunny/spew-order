package types

import (
	"encoding/json"
	"fmt"
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

type testFill struct {
	Name                     string
	Description              string
	BookOrderType            OrderType
	RequestOrder             Order
	ExpectedHold             decimal.Decimal
	ExpectedHoldSymbol       Symbol
	ExpectedOrderType        OrderType
	ExpectedTransaction      Transaction
	ExpectedOrderTransaction *Transaction
}

var flatFee int64 = 100

var tests = []testFill{
	{
		Name:        "SellMarketOrderA_BuyLimitOrderB_ALessThanB_QuantityLimit",
		Description: "a sell market order with a quantity limit on the order book is paired with a buy limit order request where the quantity of the market order is less than the quantity of the limit order",
		BookOrderType: &MarketOrderType{
			Base:     SymbolEthereum,
			Quantity: decimal.NewFromFloat(0.00042),
		},
		RequestOrder: newTestRequest(accountIDB, ActionTypeBuy, &LimitOrderType{
			Base:     SymbolBitcoin,
			Price:    decimal.NewFromFloat(0.075),
			Quantity: decimal.NewFromFloat(0.00045),
		}),
		ExpectedHold:       decimal.NewFromFloat(0.00003375),
		ExpectedHoldSymbol: SymbolBitcoin,
		ExpectedOrderType: &LimitOrderType{
			Base:     SymbolBitcoin,
			Price:    decimal.NewFromFloat(0.075),
			Quantity: decimal.NewFromFloat(0.00003),
		},
		ExpectedTransaction: Transaction{
			A: BalanceEntry{
				AccountID:   accountIDA,
				AddSymbol:   SymbolBitcoin,
				AddQuantity: decimal.NewFromFloat(0.0000315), // maker fee calculated on this amount
				SubSymbol:   SymbolEthereum,
				SubQuantity: decimal.NewFromFloat(0.00042),
			},
			B: BalanceEntry{
				AccountID:   accountIDB,
				AddSymbol:   SymbolEthereum,
				AddQuantity: decimal.NewFromFloat(0.00042), // taker fee calculated on this amount
				SubSymbol:   SymbolBitcoin,
				SubQuantity: decimal.NewFromFloat(0.0000315),
			},
			Filled: []Order{},
		},
		ExpectedOrderTransaction: &Transaction{
			A: BalanceEntry{
				AccountID:   accountIDA,
				AddSymbol:   SymbolBitcoin,
				AddQuantity: decimal.NewFromFloat(0.0000315),
				FeeQuantity: decimal.NewFromInt(flatFee),
				SubSymbol:   SymbolEthereum,
				SubQuantity: decimal.NewFromFloat(0.00042),
			},
			B: BalanceEntry{
				AccountID:   accountIDB,
				AddSymbol:   SymbolEthereum,
				AddQuantity: decimal.NewFromFloat(0.00042), // taker fee calculated on this amount
				FeeQuantity: decimal.NewFromInt(flatFee),
				SubSymbol:   SymbolBitcoin,
				SubQuantity: decimal.NewFromFloat(0.0000315),
			},
			Filled: []Order{
				newTestRequest(accountIDA, ActionTypeSell, &MarketOrderType{
					Base:     SymbolEthereum,
					Quantity: decimal.NewFromFloat(0.00042),
				}),
			},
		},
	},
	{
		Name:        "SellMarketOrderA_BuyLimitOrderB_AGreaterThanB_QuantityLimit",
		Description: "a sell market order with a quantity limit on the order book is paired with a buy limit order request where the quantity of the market order is greater than the quantity of the limit order",
		BookOrderType: &MarketOrderType{
			Base:     SymbolEthereum,
			Quantity: decimal.NewFromFloat(0.00048),
		},
		RequestOrder: newTestRequest(accountIDB, ActionTypeBuy, &LimitOrderType{
			Base:     SymbolBitcoin,
			Price:    decimal.NewFromFloat(0.075),
			Quantity: decimal.NewFromFloat(0.00045),
		}),
		ExpectedHold:       decimal.NewFromFloat(0.00003375),
		ExpectedHoldSymbol: SymbolBitcoin,
		ExpectedOrderType: &MarketOrderType{
			Base:     SymbolEthereum,
			Quantity: decimal.NewFromFloat(0.00003),
		},
		ExpectedTransaction: Transaction{
			A: BalanceEntry{
				AccountID:   accountIDA,
				AddSymbol:   SymbolBitcoin,
				AddQuantity: decimal.NewFromFloat(0.00003375),
				SubSymbol:   SymbolEthereum,
				SubQuantity: decimal.NewFromFloat(0.00045),
			},
			B: BalanceEntry{
				AccountID:   accountIDB,
				AddSymbol:   SymbolEthereum,
				AddQuantity: decimal.NewFromFloat(0.00045),
				SubSymbol:   SymbolBitcoin,
				SubQuantity: decimal.NewFromFloat(0.00003375),
			},
			Filled: []Order{
				newTestRequest(accountIDB, ActionTypeBuy, &LimitOrderType{
					Base:     SymbolBitcoin,
					Price:    decimal.NewFromFloat(0.075),
					Quantity: decimal.NewFromFloat(0.00045),
				}),
			},
		},
		ExpectedOrderTransaction: &Transaction{
			A: BalanceEntry{
				AccountID:   accountIDA,
				AddSymbol:   SymbolBitcoin,
				AddQuantity: decimal.NewFromFloat(0.00003375),
				FeeQuantity: decimal.NewFromInt(flatFee),
				SubSymbol:   SymbolEthereum,
				SubQuantity: decimal.NewFromFloat(0.00045),
			},
			B: BalanceEntry{
				AccountID:   accountIDB,
				AddSymbol:   SymbolEthereum,
				AddQuantity: decimal.NewFromFloat(0.00045),
				FeeQuantity: decimal.NewFromInt(flatFee),
				SubSymbol:   SymbolBitcoin,
				SubQuantity: decimal.NewFromFloat(0.00003375),
			},
			Filled: []Order{
				newTestRequest(accountIDB, ActionTypeBuy, &LimitOrderType{
					Base:     SymbolBitcoin,
					Price:    decimal.NewFromFloat(0.075),
					Quantity: decimal.NewFromFloat(0.00045),
				}),
			},
		},
	},
	{
		Name:        "BuyMarketOrderA_SellLimitOrderB_ALessThanB_QuantityLimit",
		Description: "a buy market order with a quantity limit on the order book is paired with a sell limit order request where the quantity of the market order is less than the quantity of the limit order",
		BookOrderType: &MarketOrderType{
			Base:     SymbolEthereum,
			Quantity: decimal.NewFromFloat(0.00042),
		},
		RequestOrder: newTestRequest(accountIDB, ActionTypeSell, &LimitOrderType{
			Base:     SymbolBitcoin,
			Price:    decimal.NewFromFloat(0.075),
			Quantity: decimal.NewFromFloat(0.00045),
		}),
		ExpectedHold:       decimal.NewFromFloat(0.00045),
		ExpectedHoldSymbol: SymbolEthereum,
		ExpectedOrderType: &LimitOrderType{
			Base:     SymbolBitcoin,
			Price:    decimal.NewFromFloat(0.075),
			Quantity: decimal.NewFromFloat(0.00003),
		},
		ExpectedTransaction: Transaction{
			A: BalanceEntry{
				AccountID:   accountIDA,
				AddSymbol:   SymbolEthereum,
				AddQuantity: decimal.NewFromFloat(0.00042),
				SubSymbol:   SymbolBitcoin,
				SubQuantity: decimal.NewFromFloat(0.0000315),
			},
			B: BalanceEntry{
				AccountID:   accountIDB,
				AddSymbol:   SymbolBitcoin,
				AddQuantity: decimal.NewFromFloat(0.0000315),
				SubSymbol:   SymbolEthereum,
				SubQuantity: decimal.NewFromFloat(0.00042),
			},
			Filled: []Order{},
		},
		ExpectedOrderTransaction: &Transaction{
			A: BalanceEntry{
				AccountID:   accountIDA,
				AddSymbol:   SymbolEthereum,
				AddQuantity: decimal.NewFromFloat(0.00042),
				FeeQuantity: decimal.NewFromInt(flatFee),
				SubSymbol:   SymbolBitcoin,
				SubQuantity: decimal.NewFromFloat(0.0000315),
			},
			B: BalanceEntry{
				AccountID:   accountIDB,
				AddSymbol:   SymbolBitcoin,
				AddQuantity: decimal.NewFromFloat(0.0000315),
				FeeQuantity: decimal.NewFromInt(flatFee),
				SubSymbol:   SymbolEthereum,
				SubQuantity: decimal.NewFromFloat(0.00042),
			},
			Filled: []Order{
				newTestRequest(accountIDA, ActionTypeBuy, &MarketOrderType{
					Base:     SymbolEthereum,
					Quantity: decimal.NewFromFloat(0.00042),
				}),
			},
		},
	},
	{
		Name:        "BuyMarketOrderA_SellLimitOrderB_AGreaterThanB_QuantityLimit",
		Description: "a buy market order with a quantity limit on the order book is paired with a sell limit order request where the quantity of the market order is greater than the quantity of the limit order",
		BookOrderType: &MarketOrderType{
			Base:     SymbolEthereum,
			Quantity: decimal.NewFromFloat(0.00048),
		},
		RequestOrder: newTestRequest(accountIDB, ActionTypeSell, &LimitOrderType{
			Base:     SymbolBitcoin,
			Price:    decimal.NewFromFloat(0.075),
			Quantity: decimal.NewFromFloat(0.00045),
		}),
		ExpectedHold:       decimal.NewFromFloat(0.00045),
		ExpectedHoldSymbol: SymbolEthereum,
		ExpectedOrderType: &MarketOrderType{
			Base:     SymbolEthereum,
			Quantity: decimal.NewFromFloat(0.00003),
		},
		ExpectedTransaction: Transaction{
			A: BalanceEntry{
				AccountID:   accountIDA,
				AddSymbol:   SymbolEthereum,
				AddQuantity: decimal.NewFromFloat(0.00045),
				SubSymbol:   SymbolBitcoin,
				SubQuantity: decimal.NewFromFloat(0.00003375),
			},
			B: BalanceEntry{
				AccountID:   accountIDB,
				AddSymbol:   SymbolBitcoin,
				AddQuantity: decimal.NewFromFloat(0.00003375),
				SubSymbol:   SymbolEthereum,
				SubQuantity: decimal.NewFromFloat(0.00045),
			},
			Filled: []Order{
				newTestRequest(accountIDB, ActionTypeBuy, &LimitOrderType{
					Base:     SymbolBitcoin,
					Price:    decimal.NewFromFloat(0.075),
					Quantity: decimal.NewFromFloat(0.00045),
				}),
			},
		},
		ExpectedOrderTransaction: &Transaction{
			A: BalanceEntry{
				AccountID:   accountIDA,
				AddSymbol:   SymbolEthereum,
				AddQuantity: decimal.NewFromFloat(0.00045),
				FeeQuantity: decimal.NewFromInt(flatFee),
				SubSymbol:   SymbolBitcoin,
				SubQuantity: decimal.NewFromFloat(0.00003375),
			},
			B: BalanceEntry{
				AccountID:   accountIDB,
				AddSymbol:   SymbolBitcoin,
				AddQuantity: decimal.NewFromFloat(0.00003375),
				FeeQuantity: decimal.NewFromInt(flatFee),
				SubSymbol:   SymbolEthereum,
				SubQuantity: decimal.NewFromFloat(0.00045),
			},
			Filled: []Order{
				newTestRequest(accountIDB, ActionTypeBuy, &LimitOrderType{
					Base:     SymbolBitcoin,
					Price:    decimal.NewFromFloat(0.075),
					Quantity: decimal.NewFromFloat(0.00045),
				}),
			},
		},
	},
	{
		Name:        "BuyMarketOrderA_SellLimitOrderB_ALessThanB_SpendingLimit",
		Description: "a buy market order with a spending limit on the order book is paired with a sell limit order request where the quantity of the market order is less than the quantity of the limit order",
		BookOrderType: &MarketOrderType{
			Base:     SymbolBitcoin,
			Quantity: decimal.NewFromFloat(0.0000315),
		},
		RequestOrder: newTestRequest(accountIDB, ActionTypeSell, &LimitOrderType{
			Base:     SymbolBitcoin,
			Price:    decimal.NewFromFloat(0.075),
			Quantity: decimal.NewFromFloat(0.00045),
		}),
		ExpectedHold:       decimal.NewFromFloat(0.00045),
		ExpectedHoldSymbol: SymbolEthereum,
		ExpectedOrderType: &LimitOrderType{
			Base:     SymbolBitcoin,
			Price:    decimal.NewFromFloat(0.075),
			Quantity: decimal.NewFromFloat(0.00003),
		},
		ExpectedTransaction: Transaction{
			A: BalanceEntry{
				AccountID:   accountIDA,
				AddSymbol:   SymbolEthereum,
				AddQuantity: decimal.NewFromFloat(0.00042),
				SubSymbol:   SymbolBitcoin,
				SubQuantity: decimal.NewFromFloat(0.0000315),
			},
			B: BalanceEntry{
				AccountID:   accountIDB,
				AddSymbol:   SymbolBitcoin,
				AddQuantity: decimal.NewFromFloat(0.0000315),
				SubSymbol:   SymbolEthereum,
				SubQuantity: decimal.NewFromFloat(0.00042),
			},
			Filled: []Order{},
		},
		ExpectedOrderTransaction: &Transaction{
			A: BalanceEntry{
				AccountID:   accountIDA,
				AddSymbol:   SymbolEthereum,
				AddQuantity: decimal.NewFromFloat(0.00042),
				FeeQuantity: decimal.NewFromInt(flatFee),
				SubSymbol:   SymbolBitcoin,
				SubQuantity: decimal.NewFromFloat(0.0000315),
			},
			B: BalanceEntry{
				AccountID:   accountIDB,
				AddSymbol:   SymbolBitcoin,
				AddQuantity: decimal.NewFromFloat(0.0000315),
				FeeQuantity: decimal.NewFromInt(flatFee),
				SubSymbol:   SymbolEthereum,
				SubQuantity: decimal.NewFromFloat(0.00042),
			},
			Filled: []Order{
				newTestRequest(accountIDA, ActionTypeBuy, &MarketOrderType{
					Base:     SymbolBitcoin,
					Quantity: decimal.NewFromFloat(0.0000315),
				}),
			},
		},
	},
	{
		Name:        "BuyMarketOrderA_SellLimitOrderB_AGreaterThanB_SpendingLimit",
		Description: "a buy market order with a spending limit on the order book is paired with a sell limit order request where the quantity of the market order is greater than the quantity of the limit order",
		BookOrderType: &MarketOrderType{
			Base:     SymbolBitcoin,
			Quantity: decimal.NewFromFloat(0.00004375),
		},
		RequestOrder: newTestRequest(accountIDB, ActionTypeSell, &LimitOrderType{
			Base:     SymbolBitcoin,
			Price:    decimal.NewFromFloat(0.075),
			Quantity: decimal.NewFromFloat(0.00045),
		}),
		ExpectedHold:       decimal.NewFromFloat(0.00045),
		ExpectedHoldSymbol: SymbolEthereum,
		ExpectedOrderType: &MarketOrderType{
			Base:     SymbolBitcoin,
			Quantity: decimal.NewFromFloat(0.00001),
		},
		ExpectedTransaction: Transaction{
			A: BalanceEntry{
				AccountID:   accountIDA,
				AddSymbol:   SymbolEthereum,
				AddQuantity: decimal.NewFromFloat(0.00045),
				SubSymbol:   SymbolBitcoin,
				SubQuantity: decimal.NewFromFloat(0.00003375),
			},
			B: BalanceEntry{
				AccountID:   accountIDB,
				AddSymbol:   SymbolBitcoin,
				AddQuantity: decimal.NewFromFloat(0.00003375),
				SubSymbol:   SymbolEthereum,
				SubQuantity: decimal.NewFromFloat(0.00045),
			},
			Filled: []Order{
				newTestRequest(accountIDB, ActionTypeBuy, &LimitOrderType{
					Base:     SymbolBitcoin,
					Price:    decimal.NewFromFloat(0.075),
					Quantity: decimal.NewFromFloat(0.00045),
				}),
			},
		},
		ExpectedOrderTransaction: &Transaction{
			A: BalanceEntry{
				AccountID:   accountIDA,
				AddSymbol:   SymbolEthereum,
				AddQuantity: decimal.NewFromFloat(0.00045),
				FeeQuantity: decimal.NewFromInt(flatFee),
				SubSymbol:   SymbolBitcoin,
				SubQuantity: decimal.NewFromFloat(0.00003375),
			},
			B: BalanceEntry{
				AccountID:   accountIDB,
				AddSymbol:   SymbolBitcoin,
				AddQuantity: decimal.NewFromFloat(0.00003375),
				FeeQuantity: decimal.NewFromInt(flatFee),
				SubSymbol:   SymbolEthereum,
				SubQuantity: decimal.NewFromFloat(0.00045),
			},
			Filled: []Order{
				newTestRequest(accountIDB, ActionTypeBuy, &LimitOrderType{
					Base:     SymbolBitcoin,
					Price:    decimal.NewFromFloat(0.075),
					Quantity: decimal.NewFromFloat(0.00045),
				}),
			},
		},
	},
	{
		Name:        "BuyLimitOrderA_SellMarketOrderB_AGreaterThanB_QuantityLimit",
		Description: "a buy limit order on the order book is paired with a sell market order request with a quantity limit where the quantity of the limit order is greater than the quantity of the market order",
		BookOrderType: &LimitOrderType{ // on the book as a buy/bid
			Base:     SymbolBitcoin,                 // base is BTC, target is ETH
			Price:    decimal.NewFromFloat(0.075),   // bid price in ETH
			Quantity: decimal.NewFromFloat(0.00045), // quantity of ETH
		},
		RequestOrder: newTestRequest(accountIDB, ActionTypeSell, &MarketOrderType{
			Base:     SymbolEthereum,                // ask for ETH, no price since market order
			Quantity: decimal.NewFromFloat(0.00042), // selling ETH
		}),
		ExpectedHold:       decimal.NewFromFloat(0.00042),
		ExpectedHoldSymbol: SymbolEthereum,
		ExpectedOrderType: &LimitOrderType{
			Base:     SymbolBitcoin,
			Price:    decimal.NewFromFloat(0.075),
			Quantity: decimal.NewFromFloat(0.00003),
		},
		ExpectedTransaction: Transaction{
			A: BalanceEntry{
				AccountID:   accountIDA,
				AddSymbol:   SymbolEthereum,                  // add the target type
				AddQuantity: decimal.NewFromFloat(0.00042),   // add sell amount
				SubSymbol:   SymbolBitcoin,                   // remove base type
				SubQuantity: decimal.NewFromFloat(0.0000315), // rem price * sell amount
			},
			B: BalanceEntry{
				AccountID:   accountIDB,
				AddSymbol:   SymbolBitcoin,
				AddQuantity: decimal.NewFromFloat(0.0000315),
				SubSymbol:   SymbolEthereum,
				SubQuantity: decimal.NewFromFloat(0.00042),
			},
			Filled: []Order{
				newTestRequest(accountIDB, ActionTypeSell, &MarketOrderType{
					Base:     SymbolEthereum,
					Quantity: decimal.NewFromFloat(0.00042),
				}),
			},
		},
		ExpectedOrderTransaction: &Transaction{
			A: BalanceEntry{
				AccountID:   accountIDA,
				AddSymbol:   SymbolEthereum,
				AddQuantity: decimal.NewFromFloat(0.00042),
				FeeQuantity: decimal.NewFromInt(flatFee),
				SubSymbol:   SymbolBitcoin,
				SubQuantity: decimal.NewFromFloat(0.0000315),
			},
			B: BalanceEntry{
				AccountID:   accountIDB,
				AddSymbol:   SymbolBitcoin,
				AddQuantity: decimal.NewFromFloat(0.0000315),
				FeeQuantity: decimal.NewFromInt(flatFee),
				SubSymbol:   SymbolEthereum,
				SubQuantity: decimal.NewFromFloat(0.00042),
			},
			Filled: []Order{
				newTestRequest(accountIDB, ActionTypeSell, &MarketOrderType{
					Base:     SymbolEthereum,
					Quantity: decimal.NewFromFloat(0.00042),
				}),
			},
		},
	},
	{
		Name:        "BuyLimitOrderA_SellMarketOrderB_ALessThanB_QuantityLimit",
		Description: "a buy limit order on the order book is paired with a sell market order request with a quantity limit where the quantity of the limit order is less than the quantity of the market order",
		BookOrderType: &LimitOrderType{
			Base:     SymbolBitcoin,
			Price:    decimal.NewFromFloat(0.075),
			Quantity: decimal.NewFromFloat(0.00045),
		},
		RequestOrder: newTestRequest(accountIDB, ActionTypeSell, &MarketOrderType{
			Base:     SymbolEthereum,
			Quantity: decimal.NewFromFloat(0.00048),
		}),
		ExpectedHold:       decimal.NewFromFloat(0.00048),
		ExpectedHoldSymbol: SymbolEthereum,
		ExpectedOrderType: &MarketOrderType{
			Base:     SymbolEthereum,
			Quantity: decimal.NewFromFloat(0.00003),
		},
		ExpectedTransaction: Transaction{
			A: BalanceEntry{
				AccountID:   accountIDA,
				AddSymbol:   SymbolEthereum,
				AddQuantity: decimal.NewFromFloat(0.00045),
				SubSymbol:   SymbolBitcoin,
				SubQuantity: decimal.NewFromFloat(0.00003375),
			},
			B: BalanceEntry{
				AccountID:   accountIDB,
				AddSymbol:   SymbolBitcoin,
				AddQuantity: decimal.NewFromFloat(0.00003375),
				SubSymbol:   SymbolEthereum,
				SubQuantity: decimal.NewFromFloat(0.00045),
			},
			Filled: []Order{},
		},
		ExpectedOrderTransaction: &Transaction{
			A: BalanceEntry{
				AccountID:   accountIDA,
				AddSymbol:   SymbolEthereum,
				AddQuantity: decimal.NewFromFloat(0.00045),
				FeeQuantity: decimal.NewFromInt(flatFee),
				SubSymbol:   SymbolBitcoin,
				SubQuantity: decimal.NewFromFloat(0.00003375),
			},
			B: BalanceEntry{
				AccountID:   accountIDB,
				AddSymbol:   SymbolBitcoin,
				AddQuantity: decimal.NewFromFloat(0.00003375),
				FeeQuantity: decimal.NewFromInt(flatFee),
				SubSymbol:   SymbolEthereum,
				SubQuantity: decimal.NewFromFloat(0.00045),
			},
			Filled: []Order{
				newTestRequest(accountIDA, ActionTypeBuy, &LimitOrderType{
					Base:     SymbolBitcoin,
					Price:    decimal.NewFromFloat(0.075),
					Quantity: decimal.NewFromFloat(0.00045),
				}),
			},
		},
	},
	{
		Name:        "SellLimitOrderA_BuyMarketOrderB_ALessThanB_SpendingLimit",
		Description: "a sell limit order on the order book is paired with a buy market order request with a spending limit where the quantity of the limit order is less than the quantity of the market order",
		BookOrderType: &LimitOrderType{
			Base:     SymbolBitcoin,
			Price:    decimal.NewFromFloat(0.075),
			Quantity: decimal.NewFromFloat(0.00045),
		},
		RequestOrder: newTestRequest(accountIDB, ActionTypeBuy, &MarketOrderType{
			Base:     SymbolBitcoin,
			Quantity: decimal.NewFromFloat(0.00004375),
		}),
		ExpectedHold:       decimal.NewFromFloat(0.00004375),
		ExpectedHoldSymbol: SymbolBitcoin,
		ExpectedOrderType: &MarketOrderType{
			Base:     SymbolBitcoin,
			Quantity: decimal.NewFromFloat(0.00001),
		},
		ExpectedTransaction: Transaction{
			A: BalanceEntry{
				AccountID:   accountIDA,
				AddSymbol:   SymbolBitcoin,
				AddQuantity: decimal.NewFromFloat(0.00003375),
				SubSymbol:   SymbolEthereum,
				SubQuantity: decimal.NewFromFloat(0.00045),
			},
			B: BalanceEntry{
				AccountID:   accountIDB,
				AddSymbol:   SymbolEthereum,
				AddQuantity: decimal.NewFromFloat(0.00045),
				SubSymbol:   SymbolBitcoin,
				SubQuantity: decimal.NewFromFloat(0.00003375),
			},
			Filled: []Order{},
		},
		ExpectedOrderTransaction: &Transaction{
			A: BalanceEntry{
				AccountID:   accountIDA,
				AddSymbol:   SymbolBitcoin,
				AddQuantity: decimal.NewFromFloat(0.00003375),
				FeeQuantity: decimal.NewFromInt(flatFee),
				SubSymbol:   SymbolEthereum,
				SubQuantity: decimal.NewFromFloat(0.00045),
			},
			B: BalanceEntry{
				AccountID:   accountIDB,
				AddSymbol:   SymbolEthereum,
				AddQuantity: decimal.NewFromFloat(0.00045),
				FeeQuantity: decimal.NewFromInt(flatFee),
				SubSymbol:   SymbolBitcoin,
				SubQuantity: decimal.NewFromFloat(0.00003375),
			},
			Filled: []Order{
				newTestRequest(accountIDA, ActionTypeSell, &LimitOrderType{
					Base:     SymbolBitcoin,
					Price:    decimal.NewFromFloat(0.075),
					Quantity: decimal.NewFromFloat(0.00045),
				}),
			},
		},
	},
	{
		Name:        "SellLimitOrderA_BuyMarketOrderB_AGreaterThanB_SpendingLimit",
		Description: "a sell limit order on the order book is paired with a buy market order request with a spending limit where the quantity of the limit order is greater than the quantity of the market order",
		BookOrderType: &LimitOrderType{
			Base:     SymbolBitcoin,
			Price:    decimal.NewFromFloat(0.075),
			Quantity: decimal.NewFromFloat(0.00045),
		},
		RequestOrder: newTestRequest(accountIDB, ActionTypeBuy, &MarketOrderType{
			Base:     SymbolBitcoin,
			Quantity: decimal.NewFromFloat(0.0000315),
		}),
		ExpectedHold:       decimal.NewFromFloat(0.0000315),
		ExpectedHoldSymbol: SymbolBitcoin,
		ExpectedOrderType: &LimitOrderType{
			Base:     SymbolBitcoin,
			Price:    decimal.NewFromFloat(0.075),
			Quantity: decimal.NewFromFloat(0.00003),
		},
		ExpectedTransaction: Transaction{
			A: BalanceEntry{
				AccountID:   accountIDA,
				AddSymbol:   SymbolBitcoin,
				AddQuantity: decimal.NewFromFloat(0.0000315),
				SubSymbol:   SymbolEthereum,
				SubQuantity: decimal.NewFromFloat(0.00042),
			},
			B: BalanceEntry{
				AccountID:   accountIDB,
				AddSymbol:   SymbolEthereum,
				AddQuantity: decimal.NewFromFloat(0.00042),
				SubSymbol:   SymbolBitcoin,
				SubQuantity: decimal.NewFromFloat(0.0000315),
			},
			Filled: []Order{
				newTestRequest(accountIDB, ActionTypeBuy, &MarketOrderType{
					Base:     SymbolBitcoin,
					Quantity: decimal.NewFromFloat(0.0000315),
				}),
			},
		},
		ExpectedOrderTransaction: &Transaction{
			A: BalanceEntry{
				AccountID:   accountIDA,
				AddSymbol:   SymbolBitcoin,
				AddQuantity: decimal.NewFromFloat(0.0000315),
				FeeQuantity: decimal.NewFromInt(flatFee),
				SubSymbol:   SymbolEthereum,
				SubQuantity: decimal.NewFromFloat(0.00042),
			},
			B: BalanceEntry{
				AccountID:   accountIDB,
				AddSymbol:   SymbolEthereum,
				AddQuantity: decimal.NewFromFloat(0.00042),
				FeeQuantity: decimal.NewFromInt(flatFee),
				SubSymbol:   SymbolBitcoin,
				SubQuantity: decimal.NewFromFloat(0.0000315),
			},
			Filled: []Order{
				newTestRequest(accountIDB, ActionTypeBuy, &MarketOrderType{
					Base:     SymbolBitcoin,
					Quantity: decimal.NewFromFloat(0.0000315),
				}),
			},
		},
	},
	{
		Name:        "SellLimitOrderA_BuyMarketOrderB_AGreaterThanB_QuantityLimit",
		Description: "a sell limit order on the order book is paired with a buy market order request with a quantity limit where the quantity of the limit order is greater than the quantity of the market order",
		BookOrderType: &LimitOrderType{
			Base:     SymbolBitcoin,
			Price:    decimal.NewFromFloat(0.075),
			Quantity: decimal.NewFromFloat(0.00045),
		},
		RequestOrder: newTestRequest(accountIDB, ActionTypeBuy, &MarketOrderType{
			Base:     SymbolEthereum,
			Quantity: decimal.NewFromFloat(0.0004),
		}),
		ExpectedHold:       decimal.NewFromInt(0),
		ExpectedHoldSymbol: SymbolBitcoin,
		ExpectedOrderType: &LimitOrderType{
			Base:     SymbolBitcoin,
			Price:    decimal.NewFromFloat(0.075),
			Quantity: decimal.NewFromFloat(0.00005),
		},
		ExpectedTransaction: Transaction{
			A: BalanceEntry{
				AccountID:   accountIDA,
				AddSymbol:   SymbolBitcoin,
				AddQuantity: decimal.NewFromFloat(0.00003),
				SubSymbol:   SymbolEthereum,
				SubQuantity: decimal.NewFromFloat(0.0004),
			},
			B: BalanceEntry{
				AccountID:   accountIDB,
				AddSymbol:   SymbolEthereum,
				AddQuantity: decimal.NewFromFloat(0.0004),
				SubSymbol:   SymbolBitcoin,
				SubQuantity: decimal.NewFromFloat(0.00003),
			},
			Filled: []Order{
				newTestRequest(accountIDB, ActionTypeBuy, &MarketOrderType{
					Base:     SymbolEthereum,
					Quantity: decimal.NewFromFloat(0.0004),
				}),
			},
		},
		ExpectedOrderTransaction: &Transaction{
			A: BalanceEntry{
				AccountID:   accountIDA,
				AddSymbol:   SymbolBitcoin,
				AddQuantity: decimal.NewFromFloat(0.00003),
				FeeQuantity: decimal.NewFromInt(flatFee),
				SubSymbol:   SymbolEthereum,
				SubQuantity: decimal.NewFromFloat(0.0004),
			},
			B: BalanceEntry{
				AccountID:   accountIDB,
				AddSymbol:   SymbolEthereum,
				AddQuantity: decimal.NewFromFloat(0.0004),
				FeeQuantity: decimal.NewFromInt(flatFee),
				SubSymbol:   SymbolBitcoin,
				SubQuantity: decimal.NewFromFloat(0.00003),
			},
			Filled: []Order{
				newTestRequest(accountIDB, ActionTypeBuy, &MarketOrderType{
					Base:     SymbolEthereum,
					Quantity: decimal.NewFromFloat(0.0004),
				}),
			},
		},
	},
	{
		Name:        "SellLimitOrderA_BuyLimitOrderB_QuantityAGreaterThanB_PriceBGreaterThanA",
		Description: "a sell limit order on the order book is paired with a buy limit order request where the quantity of A is greater than the quantity of B",
		BookOrderType: &LimitOrderType{
			Base:     SymbolBitcoin,
			Price:    decimal.NewFromFloat(0.075),
			Quantity: decimal.NewFromFloat(0.00045),
		},
		RequestOrder: newTestRequest(accountIDB, ActionTypeBuy, &LimitOrderType{
			Base:     SymbolBitcoin,
			Price:    decimal.NewFromFloat(0.075),
			Quantity: decimal.NewFromFloat(0.00042),
		}),
		ExpectedHold:       decimal.NewFromFloat(0.0000315),
		ExpectedHoldSymbol: SymbolBitcoin,
		ExpectedOrderType: &LimitOrderType{
			Base:     SymbolBitcoin,
			Price:    decimal.NewFromFloat(0.075),
			Quantity: decimal.NewFromFloat(0.00003),
		},
		ExpectedTransaction: Transaction{
			A: BalanceEntry{
				AccountID:   accountIDA,
				AddSymbol:   SymbolBitcoin,
				AddQuantity: decimal.NewFromFloat(0.0000315), // maker fee calculated on this amount
				SubSymbol:   SymbolEthereum,
				SubQuantity: decimal.NewFromFloat(0.00042),
			},
			B: BalanceEntry{
				AccountID:   accountIDB,
				AddSymbol:   SymbolEthereum,
				AddQuantity: decimal.NewFromFloat(0.00042), // taker fee calculated on this amount
				SubSymbol:   SymbolBitcoin,
				SubQuantity: decimal.NewFromFloat(0.0000315),
			},
			Filled: []Order{
				newTestRequest(accountIDB, ActionTypeBuy, &LimitOrderType{
					Base:     SymbolBitcoin,
					Price:    decimal.NewFromFloat(0.075),
					Quantity: decimal.NewFromFloat(0.00042),
				}),
			},
		},
		ExpectedOrderTransaction: &Transaction{
			A: BalanceEntry{
				AccountID:   accountIDA,
				AddSymbol:   SymbolBitcoin,
				AddQuantity: decimal.NewFromFloat(0.0000315), // maker fee calculated on this amount
				FeeQuantity: decimal.NewFromInt(flatFee),
				SubSymbol:   SymbolEthereum,
				SubQuantity: decimal.NewFromFloat(0.00042),
			},
			B: BalanceEntry{
				AccountID:   accountIDB,
				AddSymbol:   SymbolEthereum,
				AddQuantity: decimal.NewFromFloat(0.00042), // taker fee calculated on this amount
				FeeQuantity: decimal.NewFromInt(flatFee),
				SubSymbol:   SymbolBitcoin,
				SubQuantity: decimal.NewFromFloat(0.0000315),
			},
			Filled: []Order{
				newTestRequest(accountIDB, ActionTypeBuy, &LimitOrderType{
					Base:     SymbolBitcoin,
					Price:    decimal.NewFromFloat(0.075),
					Quantity: decimal.NewFromFloat(0.00042),
				}),
			},
		},
	},
	{
		Name:        "SellLimitOrderA_BuyLimitOrderB_QuantityALessThanB_PriceBGreaterThanA",
		Description: "a sell limit order on the order book is paired with a buy limit order request where the quantity of A is less than the quantity of B",
		BookOrderType: &LimitOrderType{
			Base:     SymbolBitcoin,
			Price:    decimal.NewFromFloat(0.075),
			Quantity: decimal.NewFromFloat(0.00042),
		},
		RequestOrder: newTestRequest(accountIDB, ActionTypeBuy, &LimitOrderType{
			Base:     SymbolBitcoin,
			Price:    decimal.NewFromFloat(0.075),
			Quantity: decimal.NewFromFloat(0.00045),
		}),
		ExpectedHold:       decimal.NewFromFloat(0.00003375),
		ExpectedHoldSymbol: SymbolBitcoin,
		ExpectedOrderType: &LimitOrderType{
			Base:     SymbolBitcoin,
			Price:    decimal.NewFromFloat(0.075),
			Quantity: decimal.NewFromFloat(0.00003),
		},
		ExpectedTransaction: Transaction{
			A: BalanceEntry{
				AccountID:   accountIDA,
				AddSymbol:   SymbolBitcoin,
				AddQuantity: decimal.NewFromFloat(0.0000315),
				SubSymbol:   SymbolEthereum,
				SubQuantity: decimal.NewFromFloat(0.00042),
			},
			B: BalanceEntry{
				AccountID:   accountIDB,
				AddSymbol:   SymbolEthereum,
				AddQuantity: decimal.NewFromFloat(0.00042),
				SubSymbol:   SymbolBitcoin,
				SubQuantity: decimal.NewFromFloat(0.0000315),
			},
			Filled: []Order{},
		},
		ExpectedOrderTransaction: &Transaction{
			A: BalanceEntry{
				AccountID:   accountIDA,
				AddSymbol:   SymbolBitcoin,
				AddQuantity: decimal.NewFromFloat(0.0000315),
				FeeQuantity: decimal.NewFromInt(flatFee),
				SubSymbol:   SymbolEthereum,
				SubQuantity: decimal.NewFromFloat(0.00042),
			},
			B: BalanceEntry{
				AccountID:   accountIDB,
				AddSymbol:   SymbolEthereum,
				AddQuantity: decimal.NewFromFloat(0.00042), // taker fee calculated on this amount
				FeeQuantity: decimal.NewFromInt(flatFee),
				SubSymbol:   SymbolBitcoin,
				SubQuantity: decimal.NewFromFloat(0.0000315),
			},
			Filled: []Order{
				newTestRequest(accountIDA, ActionTypeSell, &LimitOrderType{
					Base:     SymbolBitcoin,
					Price:    decimal.NewFromFloat(0.075),
					Quantity: decimal.NewFromFloat(0.00042),
				}),
			},
		},
	},
	{
		Name:        "SellLimitOrderA_BuyLimitOrderB_QuantityAEqualToB_PriceBGreaterThanA",
		Description: "a sell limit order on the order book is paired with a buy limit order request where the quantity of A is equal to the quantity of B",
		BookOrderType: &LimitOrderType{
			Base:     SymbolBitcoin,
			Price:    decimal.NewFromFloat(0.075),
			Quantity: decimal.NewFromFloat(0.00042),
		},
		RequestOrder: newTestRequest(accountIDB, ActionTypeBuy, &LimitOrderType{
			Base:     SymbolBitcoin,
			Price:    decimal.NewFromFloat(0.075),
			Quantity: decimal.NewFromFloat(0.00042),
		}),
		ExpectedHold:       decimal.NewFromFloat(0.0000315),
		ExpectedHoldSymbol: SymbolBitcoin,
		ExpectedOrderType:  nil,
		ExpectedTransaction: Transaction{
			A: BalanceEntry{
				AccountID:   accountIDA,
				AddSymbol:   SymbolBitcoin,
				AddQuantity: decimal.NewFromFloat(0.0000315),
				SubSymbol:   SymbolEthereum,
				SubQuantity: decimal.NewFromFloat(0.00042),
			},
			B: BalanceEntry{
				AccountID:   accountIDB,
				AddSymbol:   SymbolEthereum,
				AddQuantity: decimal.NewFromFloat(0.00042),
				SubSymbol:   SymbolBitcoin,
				SubQuantity: decimal.NewFromFloat(0.0000315),
			},
			Filled: []Order{
				newTestRequest(accountIDB, ActionTypeBuy, &LimitOrderType{
					Base:     SymbolBitcoin,
					Price:    decimal.NewFromFloat(0.075),
					Quantity: decimal.NewFromFloat(0.00042),
				}),
			},
		},
		ExpectedOrderTransaction: &Transaction{
			A: BalanceEntry{
				AccountID:   accountIDA,
				AddSymbol:   SymbolBitcoin,
				AddQuantity: decimal.NewFromFloat(0.0000315), // maker fee calculated on this amount
				FeeQuantity: decimal.NewFromInt(flatFee),
				SubSymbol:   SymbolEthereum,
				SubQuantity: decimal.NewFromFloat(0.00042),
			},
			B: BalanceEntry{
				AccountID:   accountIDB,
				AddSymbol:   SymbolEthereum,
				AddQuantity: decimal.NewFromFloat(0.00042), // taker fee calculated on this amount
				FeeQuantity: decimal.NewFromInt(flatFee),
				SubSymbol:   SymbolBitcoin,
				SubQuantity: decimal.NewFromFloat(0.0000315),
			},
			Filled: []Order{
				newTestRequest(accountIDB, ActionTypeBuy, &LimitOrderType{
					Base:     SymbolBitcoin,
					Price:    decimal.NewFromFloat(0.075),
					Quantity: decimal.NewFromFloat(0.00042),
				}),
				newTestRequest(accountIDA, ActionTypeSell, &LimitOrderType{
					Base:     SymbolBitcoin,
					Price:    decimal.NewFromFloat(0.075),
					Quantity: decimal.NewFromFloat(0.00042),
				}),
			},
		},
	},
	{
		Name:        "BuyLimitOrderA_SellLimitOrderB_QuantityAGreaterThanB_PriceAGreaterThanB",
		Description: "a buy limit order on the order book is paired with a sell limit order request where the quantity of A is greater than the quantity of B",
		BookOrderType: &LimitOrderType{
			Base:     SymbolBitcoin,
			Price:    decimal.NewFromFloat(0.075),
			Quantity: decimal.NewFromFloat(0.00045),
		},
		RequestOrder: newTestRequest(accountIDB, ActionTypeSell, &LimitOrderType{
			Base:     SymbolBitcoin,
			Price:    decimal.NewFromFloat(0.074),
			Quantity: decimal.NewFromFloat(0.00042),
		}),
		ExpectedHold:       decimal.NewFromFloat(0.00042),
		ExpectedHoldSymbol: SymbolEthereum,
		ExpectedOrderType: &LimitOrderType{
			Base:     SymbolBitcoin,
			Price:    decimal.NewFromFloat(0.075),
			Quantity: decimal.NewFromFloat(0.00003),
		},
		ExpectedTransaction: Transaction{
			A: BalanceEntry{
				AccountID:   accountIDA,
				AddSymbol:   SymbolEthereum,
				AddQuantity: decimal.NewFromFloat(0.00042),
				SubSymbol:   SymbolBitcoin,
				SubQuantity: decimal.NewFromFloat(0.0000315),
			},
			B: BalanceEntry{
				AccountID:   accountIDB,
				AddSymbol:   SymbolBitcoin,
				AddQuantity: decimal.NewFromFloat(0.0000315),
				SubSymbol:   SymbolEthereum,
				SubQuantity: decimal.NewFromFloat(0.00042),
			},
			Filled: []Order{
				newTestRequest(accountIDA, ActionTypeSell, &LimitOrderType{
					Base:     SymbolBitcoin,
					Price:    decimal.NewFromFloat(0.074),
					Quantity: decimal.NewFromFloat(0.00042),
				}),
			},
		},
		ExpectedOrderTransaction: &Transaction{
			A: BalanceEntry{
				AccountID:   accountIDA,
				AddSymbol:   SymbolEthereum,
				AddQuantity: decimal.NewFromFloat(0.00042),
				FeeQuantity: decimal.NewFromInt(flatFee),
				SubSymbol:   SymbolBitcoin,
				SubQuantity: decimal.NewFromFloat(0.0000315),
			},
			B: BalanceEntry{
				AccountID:   accountIDB,
				AddSymbol:   SymbolBitcoin,
				AddQuantity: decimal.NewFromFloat(0.0000315),
				FeeQuantity: decimal.NewFromInt(flatFee),
				SubSymbol:   SymbolEthereum,
				SubQuantity: decimal.NewFromFloat(0.00042),
			},
			Filled: []Order{
				newTestRequest(accountIDA, ActionTypeSell, &LimitOrderType{
					Base:     SymbolBitcoin,
					Price:    decimal.NewFromFloat(0.074),
					Quantity: decimal.NewFromFloat(0.00042),
				}),
			},
		},
	},
	{
		Name:        "BuyLimitOrderA_SellLimitOrderB_QuantityALessThanB_PriceAGreaterThanB",
		Description: "a buy limit order on the order book is paired with a sell limit order request where the quantity of A is less than the quantity of B",
		BookOrderType: &LimitOrderType{
			Base:     SymbolBitcoin,
			Price:    decimal.NewFromFloat(0.075),
			Quantity: decimal.NewFromFloat(0.00042),
		},
		RequestOrder: newTestRequest(accountIDB, ActionTypeSell, &LimitOrderType{
			Base:     SymbolBitcoin,
			Price:    decimal.NewFromFloat(0.074),
			Quantity: decimal.NewFromFloat(0.00045),
		}),
		ExpectedHold:       decimal.NewFromFloat(0.00045),
		ExpectedHoldSymbol: SymbolEthereum,
		ExpectedOrderType: &LimitOrderType{
			Base:     SymbolBitcoin,
			Price:    decimal.NewFromFloat(0.074),
			Quantity: decimal.NewFromFloat(0.00003),
		},
		ExpectedTransaction: Transaction{
			A: BalanceEntry{
				AccountID:   accountIDA,
				AddSymbol:   SymbolEthereum,
				AddQuantity: decimal.NewFromFloat(0.00042),
				SubSymbol:   SymbolBitcoin,
				SubQuantity: decimal.NewFromFloat(0.0000315),
			},
			B: BalanceEntry{
				AccountID:   accountIDB,
				AddSymbol:   SymbolBitcoin,
				AddQuantity: decimal.NewFromFloat(0.0000315),
				SubSymbol:   SymbolEthereum,
				SubQuantity: decimal.NewFromFloat(0.00042),
			},
			Filled: []Order{},
		},
		ExpectedOrderTransaction: &Transaction{
			A: BalanceEntry{
				AccountID:   accountIDA,
				AddSymbol:   SymbolEthereum,
				AddQuantity: decimal.NewFromFloat(0.00042),
				FeeQuantity: decimal.NewFromInt(flatFee),
				SubSymbol:   SymbolBitcoin,
				SubQuantity: decimal.NewFromFloat(0.0000315),
			},
			B: BalanceEntry{
				AccountID:   accountIDB,
				AddSymbol:   SymbolBitcoin,
				AddQuantity: decimal.NewFromFloat(0.0000315),
				FeeQuantity: decimal.NewFromInt(flatFee),
				SubSymbol:   SymbolEthereum,
				SubQuantity: decimal.NewFromFloat(0.00042),
			},
			Filled: []Order{
				newTestRequest(accountIDA, ActionTypeBuy, &LimitOrderType{
					Base:     SymbolBitcoin,
					Price:    decimal.NewFromFloat(0.075),
					Quantity: decimal.NewFromFloat(0.00042),
				}),
			},
		},
	},
}

func TestOrderTypeFillWith(t *testing.T) {
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			tr, ot := test.BookOrderType.FillWith(test.RequestOrder)

			if tr == nil {
				t.Fatalf("nil Transaction")
			}

			if ot != nil && test.ExpectedOrderType != nil {
				assertOrderType(t, test.ExpectedOrderType, ot)
			} else if test.ExpectedOrderType == nil && ot != nil {
				t.Fatalf("unexpected OrderType")
			}

			assertTransaction(t, test.ExpectedTransaction, *tr, false)
		})
	}
}

func TestOrderResolve(t *testing.T) {
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			baseAction := test.RequestOrder.Action
			if baseAction == ActionTypeBuy {
				baseAction = ActionTypeSell
			} else {
				baseAction = ActionTypeBuy
			}

			order := newTestRequest(accountIDA, baseAction, test.BookOrderType)
			o := &order

			tr, _ := o.Resolve(test.RequestOrder)

			if tr == nil {
				t.Fatalf("nil Transaction")
			}

			tran := test.ExpectedTransaction
			if test.ExpectedOrderTransaction != nil {
				tran = *test.ExpectedOrderTransaction
			}
			assertTransaction(t, tran, *tr, true)
		})
	}
}

func TestOrderHold(t *testing.T) {
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			order := test.RequestOrder
			symb, amt := order.OrderRequest.Type.HoldAmount(order.Action, order.Base, order.Target)
			expected := test.ExpectedHold

			if !amt.Equal(expected) {
				t.Errorf("unexpected hold amount: %s; expected %s", amt.StringFixedBank(12), expected.StringFixedBank(12))
			}

			if symb.String() != test.ExpectedHoldSymbol.String() {
				t.Errorf("unexpected hold symbol: %s; expected %s", symb.String(), test.ExpectedHoldSymbol.String())
			}
		})
	}
}

func TestMarshalOrder(t *testing.T) {
	order := NewOrder()
	order.OrderRequest = OrderRequest{
		Owner:     uuid.NewV4().String(),
		Account:   uuid.NewV4(),
		Base:      SymbolBitcoin,
		FeeHoldID: uuid.NewV4().String(),
		HoldID:    uuid.NewV4().String(),
		Target:    SymbolEthereum,
		Action:    ActionTypeBuy,
		Type: &MarketOrderType{
			Base:     SymbolBitcoin,
			Quantity: decimal.NewFromFloat(0.001),
		},
	}

	b, err := json.Marshal(order)
	assert.NoError(t, err)

	e := `{"account":"%s","action":"BUY","base":"BTC","feeHoldID":"%s","holdID":"%s","id":"%s","owner":"%s","target":"ETH","timestamp":%d,"type":{"base":"BTC","name":"MARKET","quantity":"0.001"}}`
	expected := fmt.Sprintf(e, order.Account, order.FeeHoldID, order.HoldID, order.ID, order.Owner, order.Timestamp.UnixNano())

	assert.Equal(t, expected, string(b))
}

func TestUnmarshalOrder(t *testing.T) {
	order := NewOrder()
	order.OrderRequest = OrderRequest{
		Base:   SymbolBitcoin,
		Target: SymbolEthereum,
		Action: ActionTypeBuy,
		Type: &MarketOrderType{
			Base:     SymbolBitcoin,
			Quantity: decimal.NewFromFloat(0.001),
		},
	}

	e := `{"account":"%s","action":"BUY","base":"BTC","id":"%s","owner":"%s","target":"ETH","timestamp":%d,"type":{"base":"BTC","name":"MARKET","quantity":"0.001"}}`
	j := fmt.Sprintf(e, order.Account, order.ID, order.Owner, order.Timestamp.Unix())

	var unmarshalled Order
	err := json.Unmarshal([]byte(j), &unmarshalled)

	assert.NoError(t, err)

	assert.Equal(t, order.ID, unmarshalled.ID)
}

func assertOrderType(t *testing.T, expected, actual OrderType) {

	assert.Equal(t, expected.Name(), actual.Name())

	switch v := expected.(type) {
	case *MarketOrderType:
		r, ok := actual.(*MarketOrderType)
		if !ok {
			assert.FailNow(t, "order type does not implement MarketOrderType")
		}

		assertDecimal(t, v.Quantity, r.Quantity, v.Base.RoundingPlace(), "quantity must match")
	case *LimitOrderType:
		r, ok := actual.(*LimitOrderType)
		if !ok {
			assert.FailNow(t, "order type does not implement MarketOrderType")
		}

		assertDecimal(t, v.Quantity, r.Quantity, v.Base.RoundingPlace(), "quantity must match")
	default:
		assert.FailNow(t, "unexpected order type")
	}
}

func assertDecimal(t *testing.T, expected, actual decimal.Decimal, places int32, msgAndArgs ...interface{}) {
	e := expected.StringFixed(places)
	a := actual.StringFixed(places)

	assert.Equal(t, e, a, msgAndArgs...)
}

func assertTransaction(t *testing.T, expected Transaction, actual Transaction, acctOK bool) {
	if acctOK {
		assert.Equal(t, expected.A.AccountID.String(), actual.A.AccountID.String(), "account id should match")
		assert.Equal(t, expected.B.AccountID.String(), actual.B.AccountID.String(), "account id should match")
	}

	assert.Equal(t, expected.A.AddSymbol.String(), actual.A.AddSymbol.String(), "transaction symbol entry A must match expected")
	assertDecimal(t, expected.A.AddQuantity, actual.A.AddQuantity, expected.A.AddSymbol.RoundingPlace(), "transaction add balance entry A must match expected")
	assertDecimal(t, expected.A.FeeQuantity, actual.A.FeeQuantity, expected.A.AddSymbol.RoundingPlace(), "transaction fee entry A must match expected")

	assert.Equal(t, expected.A.SubSymbol.String(), actual.A.SubSymbol.String(), "transaction symbol entry A must match expected")
	assertDecimal(t, expected.A.SubQuantity, actual.A.SubQuantity, expected.A.SubSymbol.RoundingPlace(), "transaction sub balance entry A must match expected")

	assert.Equal(t, expected.B.AddSymbol.String(), actual.B.AddSymbol.String(), "transaction symbol entry B must match expected")
	assertDecimal(t, expected.B.AddQuantity, actual.B.AddQuantity, expected.B.AddSymbol.RoundingPlace(), "transaction add balance entry B must match expected")
	assertDecimal(t, expected.B.FeeQuantity, actual.B.FeeQuantity, expected.B.AddSymbol.RoundingPlace(), "transaction fee entry B must match expected")

	assert.Equal(t, expected.B.SubSymbol.String(), actual.B.SubSymbol.String(), "transaction symbol entry B must match expected")
	assertDecimal(t, expected.B.SubQuantity, actual.B.SubQuantity, expected.B.SubSymbol.RoundingPlace(), "transaction sub balance entry B must match expected")

	assert.Len(t, actual.Filled, len(expected.Filled))

	if len(actual.Filled) == len(expected.Filled) && len(actual.Filled) > 0 {
		assert.Equal(t, expected.Filled[0].ID, actual.Filled[0].ID)
	}
}

var baseOrder = NewOrder()
var accountIDA = uuid.NewV4()
var accountIDB = uuid.NewV4()

func newTestRequest(acct uuid.UUID, a ActionType, tp OrderType) Order {
	base := baseOrder
	base.OrderRequest = OrderRequest{
		Owner:   uuid.NewV4().String(),
		Account: acct,
		Base:    SymbolBitcoin,  // base of trade pair
		Target:  SymbolEthereum, // target of trade pair
		Action:  a,              // [action] ethereum
		Type:    tp,
	}
	return base
}
