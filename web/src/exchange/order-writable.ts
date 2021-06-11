import type { Writable, Readable } from "svelte/store"
import { ActionType, Currency, OrderType } from "../constants"
import { calcTotal, roundingPlace } from "../util"

const reloadWait = 5000

const OrderWritable = (
  loader: (accountID: string, data: IfcOrderResource) => Promise<IfcOrderResource[] | IfcOrderResource>,
  account: Writable<IfcAccountResource | IfcBalanceResource>,
  price: Readable<IfcBookProductSpread>
): Writable<IfcOrderResource[] | IfcOrderResource> => {

  let store: IfcOrderCache = {
    data: [],
    loading: false,
    lastUpdate: 0,
  }
  let subs = []                     // subscriber's handlers
  let accountID = ""
  let currentPrice = {
    ask: "0.000",
    bid: "0.000",
  }
  let tick = null

  price.subscribe((p: IfcBookProductSpread) => {
    currentPrice = p
  })

  const loadRecords = () => {
    loader(accountID, null).then((value: IfcOrderResource[]) => {
      store.loading = false
      if (value !== null) {
        store.data = value
      } else {
        store.data = []
      }

      tick = setTimeout(loadRecords, reloadWait)
      account.update((a: IfcBalanceResource) => a)
      subs.forEach(sub => sub(store.data))
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

  const subscribe = (handler: (v: IfcOrderResource[]) => void) => {

    // if the list of subscribers is empty and the store is null, set the store to initial value
    // and read the value from the data store

    subs = [...subs, handler]                                 // add handler to the array of subscribers
    handler(store.data)                                            // call handler with current value
    return () => subs = subs.filter(sub => sub !== handler)   // return unsubscribe function
  }

  const set = (new_value: IfcOrderResource[] | IfcOrderResource) => {
    
    // new value could be a new single order resource
    if (isSingle(new_value)) {
      loader(accountID, new_value).then((order: IfcOrderResource) => {
        store.data.unshift(order)
        account.update((a: IfcBalanceResource) => {
          let symbol: Currency = order.order.base
          let amt = 0
          // calculate the balance change and apply it
          if (order.order.action === ActionType.Buy) {
            if (isLimit(order.order.type)) {

            } else if (isMarket(order.order.type)) {
              if (order.order.type.base === order.order.target) {
                symbol = order.order.base
              } else {
                symbol = order.order.target
              }

              amt = calcTotal(order.order.type, order.order.action, currentPrice.ask, order.order.base, order.order.target)
            }
          } else {
            if (isLimit(order.order.type)) {

            } else if (isMarket(order.order.type)) {
              amt = calcTotal(order.order.type, order.order.action, currentPrice.bid, order.order.base, order.order.target)
            }
          }

          if (a.symbol === symbol) {
            const qnt = parseFloat(a.quantity)
            a.quantity = (qnt - amt).toFixed(roundingPlace(symbol))
          }

          return a
        })
        subs.forEach(sub => sub(store.data))         // update subscribers
      })
    } else {
      store.data = new_value                       // update value
    }

    subs.forEach(sub => sub(store.data))         // update subscribers
  }

  const update: (fn: (r: IfcOrderResource) => IfcOrderResource) => void = (fn) => {
    for (let i = 0; i < store.data.length; i++) {
      const oldStatus = store.data[i].status
      store.data[i] = fn(store.data[i])
    
      // run a patch command here for a cancelled order
      if (store.data[i].status != oldStatus) {
        loader(accountID, store.data[i])
      }
    }

    set(store.data)   // update function
  }

  return { subscribe, set, update }       // store contract
}

function isSingle(item: IfcOrderResource[] | IfcOrderResource): item is IfcOrderResource {
  return (item as IfcOrderResource).guid !== undefined
}

function isLimit(item: IfcMarketOrder | IfcLimitOrder): item is IfcLimitOrder {
  return (item as IfcLimitOrder).name === OrderType.Limit
}

function isMarket(item: IfcMarketOrder | IfcLimitOrder): item is IfcMarketOrder {
  return (item as IfcMarketOrder).name === OrderType.Market
}

export default OrderWritable