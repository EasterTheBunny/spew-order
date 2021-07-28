interface IfcMarketContext {
  price: PriceWritable
  market: MarketWritable
}

interface IfcDataSocket {
  subscribe: (market: IfcMarket) => OrderBookDataManager
  unsubscribe: () => void
}