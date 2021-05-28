import { getContext, setContext } from "svelte"
import type { UserManagerSettings } from "oidc-client"
import OidcService from "./oidc-service"
import UserService from "./user-service"
import Oidc from "oidc-client"

Oidc.Log.logger = console

const CONTEXT_KEY = {}

const initOidcContext: (config: UserManagerSettings) => void = (config) => {
  // Initialize our services
  const oidc = new OidcService(config)
  const service = new UserService(oidc)

  // Setting the Svelte context
  setOidc({
    oidc: oidc,
    user: service.user,
    loggedIn: service.loggedIn,
  })
}

export const setOidc: (context: OidcContext) => void = (context) => {
  return setContext<OidcContext>(CONTEXT_KEY, context)
}

// To make retrieving the t function easier.
export const getOidc: () => OidcContext = () => {
  return getContext<OidcContext>(CONTEXT_KEY)
}

export default initOidcContext
