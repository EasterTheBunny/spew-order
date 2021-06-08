import type { Writable } from "svelte/store"
import { writable } from "svelte/store"

const PriceWritable = (): Writable<IfcBookProductSpread> => {
  const WebSocketStateEnum = {CONNECTING: 0, OPEN: 1, CLOSING: 2, CLOSED: 3};
  const { subscribe, set, update } = writable<IfcBookProductSpread>(null)
  let wsChannel = new WebSocket('wss://ws-feed.pro.coinbase.com')
  let subs = []                     // subscriber's handlers

  let product: string = "ETH-BTC"
  let bidLookup = {}
  let askLookup = {}

  const setLookups: (bids: object, asks: object) => void = (bids, asks) => {
    bidLookup = Object.keys(bids).sort((a, b) => {
      if (a[0] < b[0]) {
        return 1;
      }
      if (a[0] > b[0]) {
        return -1;
      }
      return 0;
    }).reduce(
      (obj, key) => { 
        obj[key] = bids[key]; 
        return obj;
      }, 
      {}
    )

    askLookup = Object.keys(asks).sort((a, b) => {
      if (a[0] > b[0]) {
        return 1;
      }
      if (a[0] < b[0]) {
        return -1;
      }
      return 0;
    }).reduce(
      (obj, key) => { 
        obj[key] = asks[key]; 
        return obj;
      }, 
      {}
    )
  }

  wsChannel.onopen = function() {
    const subscribe = {
      type: "subscribe",
      product_ids: [product],
      channels: ["ticker"],
    }

    const msg = JSON.stringify(subscribe)
    wsChannel.send(msg)
  }

  wsChannel.onmessage = function(evt) {
    const data = JSON.parse(evt.data)

    switch (data.type) {
      case "snapshot":
        // this is a full order book list
        const { bids, asks } = data
        const lbids = {}
        for (let x = 0; x < bids.length; x++) {
          lbids[bids[x][0]] = bids[x][1]
        }

        const lasks = {}
        for (let x = 0; x < asks.length; x++) {
          lasks[asks[x][0]] = asks[x][1]
        }

        setLookups(lbids, lasks)
        set({
          ask: Object.keys(askLookup)[0],
          bid: Object.keys(bidLookup)[0],
        })
        break
      case "l2update":
        // this is only a list of updates
        for (let x = 0; x < data.changes.length; x++) {
          if (data.changes[x][0] === "buy") {
            bidLookup[data.changes[x][1]] = bidLookup[data.changes[x][2]]
          } else {
            askLookup[data.changes[x][1]] = askLookup[data.changes[x][2]]
          }
        }
        setLookups(bidLookup, askLookup)
        set({
          ask: Object.keys(askLookup)[0],
          bid: Object.keys(bidLookup)[0],
        })
        break
      case "ticker":
        set({
          ask: data.price,
          bid: data.price,
        })
        break
    }
  }

  wsChannel.onclose = function(evt) {
    wsChannel = null;
  }

  wsChannel.onerror = function(evt) {
    if (wsChannel.readyState == WebSocketStateEnum.OPEN) {
      wsChannel.close();
    } else {
      wsChannel = null;
    }
  }

  return { subscribe, set, update }
}

export default PriceWritable