import type { Writable, Readable } from "svelte/store"

const reloadWait = 5000

const TransactionWritable = (
  loader: (accountID: string, data: IfcTransactionRequest) => Promise<IfcTransactionResource[] | IfcTransactionResource>,
  account: Writable<IfcAccountResource | IfcBalanceResource>
): Writable<IfcTransactionResource[] | IfcTransactionRequest | IfcTransactionResource> => {

  let store: IfcTransactionCache = {
    data: [],
    loading: false,
    lastUpdate: 0,
  }
  let subs = []                     // subscriber's handlers
  let accountID = ""
  let tick = null

  const loadRecords = () => {
    loader(accountID, null).then((value: IfcTransactionResource[]) => {
      store.loading = false
      const prev = store.data.length
      if (value !== null) {
        value.sort((a, b) => {
          if (a.timestamp < b.timestamp) {
            return 1
          }
          if (a.timestamp > b.timestamp) {
            return -1
          }

          return 0
        })
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

  const set = (new_value: IfcTransactionResource[] | IfcTransactionRequest) => {
    // new value could be a transaction request
    if (isRequest(new_value)) {
      loader(accountID, new_value).then((tr: IfcTransactionResource) => {
        store.data.push(tr)
        subs.forEach(sub => sub(store.data))         // update subscribers
      })
    } else {
      if (store.data === new_value) return         // same value, exit
      store.data = new_value                       // update value
    }

    subs.forEach(sub => sub(store.data))         // update subscribers
  }

  const update = (fn: (r: IfcTransactionResource) => IfcTransactionResource) => () => {
    for (let i = 0; i < store.data.length; i++) {
      store.data[i] = fn(store.data[i])
    }

    set(store.data)   // update function
  }

  return { subscribe, set, update }
}

function isRequest(item: IfcTransactionResource[] | IfcTransactionRequest): item is IfcTransactionRequest {
  return (item as IfcTransactionRequest).address !== undefined
}

export default TransactionWritable