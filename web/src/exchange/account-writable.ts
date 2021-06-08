import type { Writable, Readable } from "svelte/store"
import type { User } from "oidc-client"

const AccountWritable = (loader: (accountID: string) => Promise<IfcAccountResource>, subscribedUser: Readable<User>): Writable<IfcAccountResource | IfcBalanceResource> => {

  let store: IfcAccountCache = {
    data: null,
    loading: false,
    lastUpdate: 0,
  }
  let subs = []                     // subscriber's handlers

  // for any change in user
  subscribedUser.subscribe((user: User) => {
    if (user === null) {
      store.data = null
    } else {
      store.loading = true
      loader(null).then((value) => {
        store.loading = false
        store.data = value
        subs.forEach(sub => sub(store.data)) 
      })
    }

    subs.forEach(sub => sub(store.data))
  })

  const subscribe = (handler: (v: IfcAccountResource) => void) => {

    // if the list of subscribers is empty and the store is null, set the store to initial value
    // and read the value from the data store

    subs = [...subs, handler]                                 // add handler to the array of subscribers
    handler(store.data)                                            // call handler with current value
    return () => subs = subs.filter(sub => sub !== handler)   // return unsubscribe function
  }

  const set = (new_value: IfcAccountResource) => {
    store.data = new_value                       // update value
    subs.forEach(sub => sub(store.data))         // update subscribers
  }

  const update = (fn: (r: IfcBalanceResource) => IfcBalanceResource) => {
    for (let x = 0; x < store.data.balances.length; x++) {
      store.data.balances[x] = fn(store.data.balances[x])
    }

    store.loading = true
    set(store.data)   // update function

    // the value is allowed to be updated from other functions on the page
    // allowing for a more dynamic interface. but to ensure data consistency
    // the balances need to be pulled from the data source.
    loader(store.data.id).then((value: IfcAccountResource) => {
      store.loading = false
      store.data = value
      set(store.data)
    })
  }

  return { subscribe, set, update }       // store contract
}

export default AccountWritable