import type ExchangeAPI from "../../oidc"
import { Currency } from "../../constants"

export enum TransakEnvironment {
  STAGING = "STAGING",
  PRODUCTION = "PRODUCTION",
}

async function getHash(c: Currency, a: string, api: ExchangeAPI) {
  let func = api.getAddressFunc()
  const addr = await func(a, c)
  return addr.address
}

export const getTransakConfig: (
  user: User,
  account: IfcAccountResource,
  api: ExchangeAPI,
  list: Currency[],
  key: string,
  environment: TransakEnvironment,
  redirect: string
) => any = async (
  user,
  account,
  api,
  list,
  key, 
  environment,
  redirect
) => {
  if (user == null) {
    return null
  }

  if (account == null) {
    return null
  }

  const ethHash = await getHash(Currency.Ethereum, account.id, api)
  const btcHash = await getHash(Currency.Bitcoin, account.id, api)
  const dgeHash = await getHash(Currency.Dogecoin, account.id, api)

  return {
    apiKey: key,
    environment: environment,
    themeColor: '000000',
    email: user.profile.name, // Your customer's email address
    redirectURL: redirect,
    hostURL: "https://app.ciphermtn.com",
    widgetHeight: '550px',
    widgetWidth: '450px',
    walletAddressesData: {
      networks : {
        'ethereum' : {address : ethHash},
      },
      coins : {
        'BTC': {address : btcHash},
        'DOGE': {address : dgeHash},
      }
    },
    cryptoCurrencyList: list.join(","),
  }
}
