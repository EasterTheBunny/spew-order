import { getContext, setContext } from "svelte"
import type { Readable } from "svelte/store"
import type { User } from "oidc-client"
import AccountWritable from "./account-writable"
import ExchangeAPI from "./api-service"
import OrderWritable from "./order-writable"
import TransactionWritable from "./transaction-writable"

const CONTEXT_KEY = {}

const initDataContext: (subscribedUser: Readable<User>) => void = (subscribedUser) => {

  const url: string = process.env.API_URL;
  const api = new ExchangeAPI(url)

  subscribedUser.subscribe((u: User) => {
    if (!!u) {
      api.setBearerToken(u.id_token)
    }
  })

  const account = AccountWritable(api.getActiveAccountFunc(), subscribedUser)
  const orders = OrderWritable(api.getOrderFunc(), account)
  const transactions = TransactionWritable(api.getTransactionFunc(), account)

  setDataCtx({
    api,
    account,
    orders,
    transactions,
  })
}

export const setDataCtx: (context: IfcDataContext) => void = (context) => {
  return setContext<IfcDataContext>(CONTEXT_KEY, context)
}

export const getDataCtx: () => IfcDataContext = () => {
  return getContext<IfcDataContext>(CONTEXT_KEY)
}

export default initDataContext
