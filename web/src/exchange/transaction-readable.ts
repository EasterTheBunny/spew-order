import type { Writable, Readable } from "svelte/store"

const reloadWait = 5000

const TransactionReadable = (
  loader: (accountID: string) => Promise<IfcTransactionResource[]>,
  account: Writable<IfcAccountResource | IfcBalanceResource>
): Readable<IfcTransactionResource[]> => {

  let store: IfcTransactionCache = {
    data: [],
    loading: false,
    lastUpdate: 0,
  }
  let subs = []                     // subscriber's handlers
  let accountID = ""
  let tick = null

  const loadRecords = () => {
    loader(accountID).then((value: IfcTransactionResource[]) => {
      store.loading = false
      const prev = store.data.length
      if (value !== null) {
        store.data = value
      } else {
        store.data = []
      }

      tick = setTimeout(loadRecords, reloadWait)
      // only notify if new records are loaded
      if (prev !== store.data.length) {
        // force the account balance to reload
        account.update((a: IfcBalanceResource) => a)
        subs.forEach(sub => sub(store.data))
      }
    })
  }

  // for any change in account
  account.subscribe((acc: IfcAccountResource) => {
    if (acc !== null && acc.id !== accountID) {
      accountID = acc.id
      store.loading = true
      store.data = []
      loadRecords()
    } else if (acc === null) {
      store.data = []
      accountID = ""
      if (tick !== null) {
        clearTimeout(tick)
      }
    }

    subs.forEach(sub => sub(store.data))
  })

  const subscribe = (handler: (v: IfcTransactionResource[]) => void) => {
    // if the list of subscribers is empty and the store is null, set the store to initial value
    // and read the value from the data store

    subs = [...subs, handler]                                 // add handler to the array of subscribers
    handler(store.data)                                            // call handler with current value
    return () => subs = subs.filter(sub => sub !== handler)   // return unsubscribe function
  }

  return { subscribe }
}

export default TransactionReadable