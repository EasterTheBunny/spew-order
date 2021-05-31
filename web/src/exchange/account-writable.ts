import type { Readable } from "svelte/store"
import type { User } from "oidc-client"

const AccountWritable = (loader: (path: string) => Promise<IfcAccountResource>, subscribedUser: Readable<User>) => {

  const path = "/account"

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
      loader("/account").then((value) => {
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
    if (store.data === new_value) return         // same value, exit
    store.data = new_value                       // update value
    subs.forEach(sub => sub(store.data))         // update subscribers
  }

  const update = (fn: (r: IfcAccountResource) => IfcAccountResource) => set(fn(store.data))   // update function

  return { subscribe, set, update }       // store contract
}

export default AccountWritable