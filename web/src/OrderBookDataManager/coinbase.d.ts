interface CBMessage {
  type: string
}

interface CBCoinbaseSnapshot extends CBMessage {
  product_id: string
  bids: string[][]
  asks: string[][]
}

interface CBCoinbaseL2Update extends CBMessage {
  product_id: string
  time: string
  changes: string[][]
}

interface CBCoinbaseTicker extends CBMessage {
  product_id: string
  price: string
  last_size: string
  best_ask: string
  best_bid: string
  side: string
  volume_24h: string
  time: string
  trade_id: number
  sequence: number
  high_24h: string
  low_24h: string
  open_24h: string
  volume_30d: string
}
