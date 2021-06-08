import { getContext, setContext } from "svelte"
import type { Readable } from "svelte/store"
import type { User } from "oidc-client"
import AccountWritable from "./account-writable"
import ExchangeAPI from "./api-service"
import OrderWritable from "./order-writable"
import PriceWritable from "./price-writable"

const CONTEXT_KEY = {}

const initDataContext: (subscribedUser: Readable<User>) => void = (subscribedUser) => {

  const api = new ExchangeAPI("http://localhost:8080/api")

  const price = PriceWritable()
  const account = AccountWritable(api.getActiveAccountFunc(), subscribedUser)
  const orders = OrderWritable(api.getOrderFunc(), account, price)

  setDataCtx({
    api,
    account,
    orders,
    price,
  })

}

export const setDataCtx: (context: IfcDataContext) => void = (context) => {
  return setContext<IfcDataContext>(CONTEXT_KEY, context)
}

export const getDataCtx: () => IfcDataContext = () => {
  return getContext<IfcDataContext>(CONTEXT_KEY)
}

export default initDataContext
