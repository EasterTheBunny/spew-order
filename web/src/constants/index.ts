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
  BitcoinCash = 'BCH',
  Dogecoin = "DOGE",
  Uniswap = "UNI",
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

export enum WorkerMessageType {
  Ticker = "TICKER",
  Book = "BOOK",
}

export const markets: IfcMarket[] = [
  {
    base: Currency.Bitcoin,
    target: Currency.Ethereum,
  },
  {
    base: Currency.Bitcoin,
    target: Currency.BitcoinCash,
  },
  {
    base: Currency.Bitcoin,
    target: Currency.Dogecoin,
  },
  {
    base: Currency.Bitcoin,
    target: Currency.Uniswap,
  },
]

// left is this exchange; right is coinbase
export const CoinbaseMarketMap = {
  "BTC-ETH": "ETH-BTC",
  "BTC-BCH": "BCH-BTC",
  "BTC-DOGE": "DOGE-BTC",
  "BTC-UNI": "UNI-BTC",
}

export const validMarket: (market: IfcMarket) => boolean = (market) => {
  if (market === null) {
    return false
  }

  for (let x = 0; x < markets.length; x++) {
    if (markets[x].base === market.base && markets[x].target === market.target) {
      return true
    }
  }

  return false
}

export const marketFromString: (market: string) => IfcMarket | null = (market) => {
  const matcher = new RegExp('^([a-z]+)\-([a-z]+)$', 'i')
  const matches = market.trim().match(matcher)
  const currencies = [
    Currency.Bitcoin,
    Currency.Ethereum,
    Currency.BitcoinCash,
    Currency.Dogecoin,
    Currency.Uniswap,
  ]

  if (matches.length == 3) {
    const parsed: Currency[] = []
    const captures = matches.slice(1)
   
    while (captures.length > 0) {
      const capture = captures.shift().toUpperCase()

      for (let x = 0; x < currencies.length; x++) {
        if (currencies[x] == capture) {
          parsed.push(currencies[x])
          currencies.splice(x, 1)
        }
      }
    }

    if (parsed.length == 2) {
      return {
        base: parsed[0],
        target: parsed[1],
      }
    } else {
      return null
    }
  } else {
    return null
  }
}