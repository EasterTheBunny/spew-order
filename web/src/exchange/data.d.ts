interface IfcAPIResponse<T = any> {
  data?: T
}

interface IfcAccountResource {
  id: string
  balances: IfcBalanceResource[]
}

interface IfcAccountCache {
  data: IfcAccountResource
  loading: boolean
  lastUpdate: number
}

interface IfcBalanceResource {
  symbol: string
  quantity: string
  funding: string
}

interface IfcDataContext {
  api: ExchangeAPI
  account: AccountWritable
}

interface IfcMarketOrderRequest {
  name: OrderType
  base: Currency
  quantity: string
}