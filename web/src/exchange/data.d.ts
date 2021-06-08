interface IfcAPIResponse<T = any> {
  data?: T
}

interface IfcAccountResource {
  id: string
  balances: IfcBalanceResource[]
}

interface IfcOrderResource {
  id: string
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
}

interface IfcMarketOrderRequest {
  name: OrderType
  base: Currency
  quantity: string
}

interface IfcBookProductSpread {
  ask: string
  bid: string
}