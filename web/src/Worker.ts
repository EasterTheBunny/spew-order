import { OrderBookDataManager } from './OrderBookDataManager'
import { WorkerMessageType } from './constants'

let dataManager: OrderBookDataManager | null = null
let lastTick: CBCoinbaseTicker | null = null
let timerHandle
const updateRate = 1000

interface DedicatedWorkerGlobalScope {
  postMessage: (msg: any) => void
  addEventListener: (msg: string, fnc: (event: MessageEvent) => void) => void
}

const _self: DedicatedWorkerGlobalScope = self as any;

_self.addEventListener('message', function(e) {
  const message = e.data || e;

  switch (message.type) {
    case 'init':
      dataManager = new OrderBookDataManager()
      connect(dataManager)
      timerHandle = setInterval(() => {
        if (!!lastTick) {
          const msg: IfcTickerMessage = {
            type: WorkerMessageType.Ticker,
            price: lastTick.price,
          }
          _self.postMessage(msg)
        }

        const asks = dataManager.topAsks()
        const bids = dataManager.topBids()

        const book: IfcBookMessage = {
          type: WorkerMessageType.Book,
          maxDepth: asks.concat(bids).reduce((a, v) => parseFloat(v[1]) > a ? parseFloat(v[1]) : a, 0),
          asks: dataManager.topAsks(),
          bids: dataManager.topBids(),
        }

        _self.postMessage(book)
      }, updateRate)
      break;
    case 'destroy':
      clearInterval(timerHandle)
      break;
    default:
      break;
  }
});

const processTicker: (data: CBCoinbaseTicker) => void = (data) => {
  lastTick = data
}

const connect: (m: OrderBookDataManager) => void = (m) => {
  let channel = new WebSocket('wss://ws-feed.pro.coinbase.com')

  channel.onopen = () => {
    const subscribe = {
      type: "subscribe",
      product_ids: ["ETH-BTC"],
      channels: ["level2", "ticker"],
    }

    channel.send(JSON.stringify(subscribe))
  }

  channel.onmessage = (e) => {
    const data = JSON.parse(e.data)

    if (isSnapshot(data)) {
      if (!!dataManager) {
        dataManager.processSnapshot(data)
      }
    } else if (isUpdate(data)) {
      if (!!dataManager) {
        dataManager.processUpdate(data)
      }
    } else if (isTicker(data)) {
      // process ticker
      processTicker(data)
    }
  }

  channel.onclose = (e) => {
    setTimeout(function() {
      connect(dataManager);
    }, 1000);
    channel = null
  }

  channel.onerror = (e: any) => {
    console.error('Socket encountered error: ', e.message, 'Closing socket');
    channel.close();
  }
}

function isSnapshot(item: CBMessage): item is CBCoinbaseSnapshot {
  return (item as CBCoinbaseSnapshot).type === "snapshot"
}

function isUpdate(item: CBMessage): item is CBCoinbaseL2Update {
  return (item as CBCoinbaseL2Update).type === "l2update"
}

function isTicker(item: CBMessage): item is CBCoinbaseTicker {
  return (item as CBCoinbaseTicker).type === "ticker"
}