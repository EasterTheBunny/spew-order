import type { Writable } from "svelte/store"
import { writable } from "svelte/store"
import WebWorker from 'web-worker:../Worker.ts'
import { WorkerMessageType } from '../constants'

const PriceWritable = (): Writable<IfcBookProductSpread> => {
  const { subscribe, set, update } = writable<IfcBookProductSpread>({ maxDepth: 0, ask: "0.000", bid: "0.000", change24hr: "", range24hr: "", asks: [], bids: []})
  const worker = new WebWorker()

  worker.onmessage = (evt) => {
    if (isTicker(evt.data)) {
      const msg: IfcTickerMessage = evt.data

      const open = parseFloat(msg.open_24h)
      const price = parseFloat(msg.price)

      update((v) => {
        v.ask = msg.price
        v.bid = msg.price
        v.change24hr = (price - open).toFixed(8) + " (" + (((price / open) - 1) * 100).toFixed(2) + "%)"
        v.range24hr = msg.low_24h + " - " + msg.high_24h

        return v
      })
    } else if (isBook(evt.data)) {
      const msg: IfcBookMessage = evt.data

      update((v) => {
        v.maxDepth = msg.maxDepth
        v.asks = msg.asks
        v.bids = msg.bids

        return v
      })
    }
  }
  
  worker.postMessage({type: 'init'})

  return { subscribe, set, update }
}

function isTicker(item: IfcWorkerMessage): item is IfcTickerMessage {
  return (item as IfcTickerMessage).type === WorkerMessageType.Ticker
}

function isBook(item: IfcWorkerMessage): item is IfcBookMessage {
  return (item as IfcBookMessage).type === WorkerMessageType.Book
}

export default PriceWritable