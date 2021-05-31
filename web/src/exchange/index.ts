import { getContext, setContext } from "svelte"
import type { Readable } from "svelte/store"
import type { User } from "oidc-client"
import AccountWritable from "./account-writable"
import ExchangeAPI from "./api-service"

const CONTEXT_KEY = {}

const initDataContext: (subscribedUser: Readable<User>) => void = (subscribedUser) => {

  const api = new ExchangeAPI("http://localhost:8080/api")

  const account = AccountWritable(api.getActiveAccountFunc(), subscribedUser)

  setDataCtx({
    api,
    account,
  })

}

export const setDataCtx: (context: IfcDataContext) => void = (context) => {
  return setContext<IfcDataContext>(CONTEXT_KEY, context)
}

export const getDataCtx: () => IfcDataContext = () => {
  return getContext<IfcDataContext>(CONTEXT_KEY)
}

export default initDataContext
