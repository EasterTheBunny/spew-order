import { OrderBookDataManager } from '../OrderBookDataManager'

const socket: () => IfcDataSocket = () => {
  let WebSocketStateEnum = {CONNECTING: 0, OPEN: 1, CLOSING: 2, CLOSED: 3}
  let msgQueue = []
  let closedByHost = true
  let retryWait = 500
  const url = 'wss://ws-feed.pro.coinbase.com'
  let channel
  let timeout
  let subscriptions = []
  let dataManager: OrderBookDataManager | null = null

  const openChannel = () => {
    channel = new WebSocket(url);
    channel.onopen = function() {
      while (msgQueue.length > 0) {
        channel.send(JSON.stringify(msgQueue.shift()))
      }
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
        dataManager.processTicker(data)
      }
    }

    // a websocket can be closed by the host at any point
    // in that event, open a new connection to keep the messages coming in
    channel.onclose = (e) => {
      if (closedByHost && retryWait <=16000) {
        // reopen with exponential backoff
        timeout = setTimeout(function() {
          if (!channel || channel.readyState === WebSocketStateEnum.CLOSED) {
            openChannel()
          }
        }, retryWait);
        retryWait = retryWait * 2
      } else {
        // reset state for new connection
        closedByHost = true
        retryWait = 500
      }
      channel = null
    }

    channel.onerror = (e) => {
      if (channel.readyState == WebSocketStateEnum.OPEN) {
        console.error('Socket encountered error: ', e.message, 'Closing socket');
        channel.close();
      } else {
        channel = null;
      }
    }
  }

  const subscribe = (market: IfcMarket): OrderBookDataManager => {
    const m = market.base + "-" + market.target
    let msgs = []

    if (subscriptions.length > 0) {
      const msg = {
        type: "unsubscribe",
        product_ids: subscriptions,
        channels: ["level2", "ticker"],
      }
      msgs.push(msg)
      subscriptions = []
    }

    const marketLookup = {
      "BTC-ETH": "ETH-BTC",
      "BTC-BCH": "BCH-BTC",
    }

    const msg = {
      type: "subscribe",
      product_ids: [marketLookup[m]],
      channels: ["level2", "ticker"],
    }

    msgs.push(msg)
    dataManager = new OrderBookDataManager()

    for (let x = 0; x < msgs.length; x++) {
      const newMsg = msgs[x]
      if (!channel || channel.readyState != WebSocketStateEnum.OPEN) {
        msgQueue.push(newMsg);
      } else {
        channel.send(JSON.stringify(newMsg))
      }
    }
    
    if (!channel) {
      if (!!timeout) {
        clearTimeout(timeout)
      }
      openChannel()
    }
    subscriptions.push(marketLookup[m])

    return dataManager
  }

  const unsubscribe = () => {
    if (channel && channel.readyState == WebSocketStateEnum.OPEN) {
      const msg = {
        type: "unsubscribe",
        product_ids: subscriptions,
        channels: ["level2", "ticker"],
      }
      subscriptions = []
      channel.send(JSON.stringify(msg))
    }
  }

  return {
    subscribe,
    unsubscribe,
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

export default socket