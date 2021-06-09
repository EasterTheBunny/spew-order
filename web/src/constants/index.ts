export enum ActionType {
  Buy = "BUY",
  Sell = "SELL",
}

export enum OrderType {
  Market = "MARKET",
  Limit = "LIMIT",
}

export enum Currency {
  Bitcoin = "BTC",
  Ethereum = "ETH",
}

export enum OrderStatus {
  Open = "OPEN",
  Partial = "PARTIAL",
  Filled = "FILLED",
  Cancelled = "CANCELLED",
}

export enum TransactionType {
  Order = "ORDER",
  Deposit = "DEPOSIT",
  Transfer = "TRANSFER",
}