import { OrderBookDataManager } from "./index";

const initialSnapshot: CBCoinbaseSnapshot = {
  type: "snapshot",
  product_id: "BTC-ETH",
  asks: [
    ["0.00042108", "500.00"],
    ["0.00042208", "500.00"],
    ["0.00042308", "500.00"],
  ],
  bids: [
    ["0.000415", "19530.3"],
    ["0.000416", "200.7"],
    ["0.00040", "5495.19999999"],
    ["0.00039", "5533.40000001"],
    ["0.00038", "100.5"],
  ],
}

const updateSnapshot: CBCoinbaseL2Update = {
  type: "update",
  product_id: "BTC-ETH",
  time: "0",
  changes: [
    ["buy", "0.00041600", "0.00000000"],
    ["sell", "0.00041600", "100.4"],
  ],
}

it('processes coinbase snapshot', () => {
  const ob = new OrderBookDataManager();
  
  ob.processSnapshot(initialSnapshot)

  expect(ob.asks.length).toBe(3)
  expect(ob.bids.length).toBe(5)
})

it('processes coinbase update', () => {
  const ob = new OrderBookDataManager();
  
  ob.processSnapshot(initialSnapshot)
  ob.processUpdate(updateSnapshot)

  expect(ob.asks.length).toBe(4)
  expect(ob.bids.length).toBe(4)
})

it('rolls up asks on precision 5', () => {
  const ob = new OrderBookDataManager();
  const precision = 5
  
  ob.processSnapshot(initialSnapshot)
  ob.processUpdate(updateSnapshot)

  const expected = [
    ["0.00042", 1600.4],
    ["0.00043", 0],
    ["0.00044", 0],
    ["0.00045", 0],
    ["0.00046", 0],
    ["0.00047", 0],
    ["0.00048", 0],
    ["0.00049", 0],
    ["0.00050", 0],
    ["0.00051", 0],
  ]
  const asks = ob.topAsks(precision)
  
  expect(asks.length).toBe(10)
  expect(asks).toStrictEqual(expected)
})

it('rolls up bids on precision 5', () => {
  const ob = new OrderBookDataManager();
  const precision = 5
  
  ob.processSnapshot(initialSnapshot)
  ob.processUpdate(updateSnapshot)

  const expected = [
    ["0.00042", 19530.3],
    ["0.00040", 5495.19999999],
    ["0.00039", 5533.40000001],
    ["0.00038", 100.5],
    ["0.00037", 0],
    ["0.00036", 0],
    ["0.00035", 0],
    ["0.00034", 0],
    ["0.00033", 0],
    ["0.00032", 0],
  ]
  const bids = ob.topBids(precision)
  
  expect(bids.length).toBe(10)
  expect(bids).toStrictEqual(expected)
})