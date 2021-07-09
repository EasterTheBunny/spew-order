interface IfcAPIResponse<T = any> {
  data?: T
}

interface IfcAccountResource {
  id: string
  balances: IfcBalanceResource[]
}

interface IfcOrderResource {
  guid: string
  status: OrderStatus
  order: IfcOrderRequest
}

interface IfcOrderRequest {
  base: Currency
  target: Currency
  action: ActionType
  type: IfcMarketOrder | IfcLimitOrder
}

interface IfcMarketOrder {
  name: OrderType
  base: Currency
  quantity: string
}

interface IfcLimitOrder {
  name: OrderType
  base: Currency
  quantity: string
  price: string
}

interface IfcAccountCache {
  data: IfcAccountResource
  loading: boolean
  lastUpdate: number
}

interface IfcOrderCache {
  data: IfcOrderResource[]
  loading: boolean
  lastUpdate: number
}

interface IfcTransactionCache {
  data: IfcTransactionResource[]
  loading: boolean
  lastUpdate: number
}

interface IfcBalanceResource {
  symbol: Currency
  quantity: string
  funding: string
}

interface IfcDataContext {
  api: ExchangeAPI
  account: AccountWritable
  orders: OrderWritable
  price: PriceWritable
  transactions: TransactionReadable
}

interface IfcMarketOrderRequest {
  name: OrderType
  base: Currency
  quantity: string
}

interface IfcBookProductSpread {
  maxDepth: number
  ask: string
  bid: string
  change24hr: string
  range24hr: string
  asks: string[][]
  bids: string[][]
}

interface IfcTransactionResource {
  type: TransactionType
  symbol: Currency
  quantity: string
  fee: string
  orderid: string
  timestamp: string
  transactionHash: string
}

interface IfcTransactionRequest {
  symbol: Currency
  address: string
  quantity: string
}