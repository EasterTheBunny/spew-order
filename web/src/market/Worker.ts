import type { OrderBookDataManager } from '../OrderBookDataManager'
import { WorkerMessageType } from '../constants'
import websocket from './socket'

let dataManager: OrderBookDataManager | null = null
let timerHandle
const updateRate = 3000
const sock = websocket()

interface DedicatedWorkerGlobalScope {
  postMessage: (msg: any) => void
  addEventListener: (msg: string, fnc: (event: MessageEvent) => void) => void
}

const _self: DedicatedWorkerGlobalScope = self as any;

_self.addEventListener('message', function(e) {
  const message = e.data || e;

  switch (message.type) {
    case 'subscribe':
      dataManager = sock.subscribe(message.market)
      timerHandle = setInterval(() => {
        let tick = dataManager.lastTick()
        let precision = 3

        if (!!tick) {
          const msg: IfcTickerMessage = {
            type: WorkerMessageType.Ticker,
            price: tick.price,
            high_24h: tick.high_24h,
            low_24h: tick.low_24h,
            open_24h: tick.open_24h,
          }
          _self.postMessage(msg)

          // calculate precision from last price
          let price = tick.price;
          let ind = price.indexOf(".")
          let calcPrecision = 0
          for (let i = ind; i < price.length; i++) {
            calcPrecision++
            if (price[i] != "0" && price[i] != ".") {
              break
            }
          }

          if (calcPrecision > precision) {
            precision = calcPrecision
          }
        }

        const asks = dataManager.topAsks(precision)
        const bids = dataManager.topBids(precision)

        const book: IfcBookMessage = {
          type: WorkerMessageType.Book,
          maxDepth: asks.concat(bids).reduce((a, v) => parseFloat(v[1]) > a ? parseFloat(v[1]) : a, 0),
          asks: asks,
          bids: bids,
        }

        _self.postMessage(book)
      }, updateRate)
      break
    case 'unsubscribe':
      clearInterval(timerHandle)
      sock.unsubscribe()
      break
    default:
      break;
  }
});