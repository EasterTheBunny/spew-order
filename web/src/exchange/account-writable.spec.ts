import { writable } from "svelte/store"
import AccountWritable from "./account-writable"
import type { User } from "oidc-client"

const delayedLoad = function<T>(delay: number, response: T, callback: () => void): (path: string) => Promise<T> {
  return (path: string) => {
    return new Promise((resolve, _) => {
      setTimeout(() => {
        callback()
        resolve(response)
      }, delay)
    })
  }
}

it('loads account data when user updates', () => {
  const w = writable<User>(null)
  const u: User = {
    id_token: "",
    access_token: "",
    token_type: "",
    scope: "",
    profile: null,
    expires_at: 0,
    state: "",
    expires_in: 0,
    expired: false,
    scopes: [],
    toStorageString: (): string => { return ""}
  }

  const r: IfcAccountResource = {
    id: "account-id-test-1",
    balances: [{
      symbol: "BTC",
      quantity: "1.234",
      funding: "funding-test-hash-btc",
    },{
      symbol: "ETH",
      quantity: "24.23498",
      funding: "funding-test-hash-eth",
    }],
  }
  let v = 0
  let delay = 100
  const l = delayedLoad<IfcAccountResource>(delay, r, () => v++)
  const a = AccountWritable(l, w)
  const { subscribe, set, update } = a

  let expectedAccount: IfcAccountResource = null
  let expectedLoadIncrement = 0

  // ensure initial state with empty account
  subscribe((acct: IfcAccountResource) => {
    expect(acct).toBe(expectedAccount)
    expect(v).toBe(expectedLoadIncrement)
  })

  w.set(u)

  expectedAccount = r
  expectedLoadIncrement++

})