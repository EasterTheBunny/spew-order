import { getContext, setContext } from "svelte"
import { WorkerMessageType } from '../constants'
import PriceWritable from "./price-writable"
import MarketWritable from "./market-writable"
import WebWorker from 'web-worker:./Worker.ts'

const CONTEXT_KEY = {}

const initMarketContext: () => void = () => {
  const worker = new WebWorker()
  const market = MarketWritable(worker)
  const price = PriceWritable()

  worker.onmessage = (evt) => {
    if (isTicker(evt.data)) {
      const msg: IfcTickerMessage = evt.data

      const open = parseFloat(msg.open_24h)
      const p = parseFloat(msg.price)

      price.update((v) => {
        v.ask = msg.price
        v.bid = msg.price
        v.change24hr = (p - open).toFixed(8) + " (" + (((p / open) - 1) * 100).toFixed(2) + "%)"
        v.range24hr = msg.low_24h + " - " + msg.high_24h

        return v
      })
    } else if (isBook(evt.data)) {
      const msg: IfcBookMessage = evt.data

      price.update((v) => {
        v.maxDepth = msg.maxDepth
        v.asks = msg.asks
        v.bids = msg.bids

        return v
      })
    }
  }
  
  worker.postMessage({type: 'init'})

  setMarketCtx({
    price,
    market,
  })
}

export const setMarketCtx: (context: IfcMarketContext) => void = (context) => {
  return setContext<IfcMarketContext>(CONTEXT_KEY, context)
}

export const getMarketCtx: () => IfcMarketContext = () => {
  return getContext<IfcMarketContext>(CONTEXT_KEY)
}

function isTicker(item: IfcWorkerMessage): item is IfcTickerMessage {
  return (item as IfcTickerMessage).type === WorkerMessageType.Ticker
}

function isBook(item: IfcWorkerMessage): item is IfcBookMessage {
  return (item as IfcBookMessage).type === WorkerMessageType.Book
}

export default initMarketContext