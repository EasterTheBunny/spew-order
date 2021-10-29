import SortedSet from 'js-sorted-set'

export class OrderBookDataManager {
  public bids: OrderedValues
  public asks: OrderedValues
  private tick: CBCoinbaseTicker | null = null

  public constructor() {}

  public processSnapshot(snapshot: CBCoinbaseSnapshot): void {
    const { bids, asks } = snapshot

    this.bids = new OrderedValues("bids", bids)
    this.asks = new OrderedValues("asks", asks)
  }

  public processUpdate(update: CBCoinbaseL2Update): void {
    for (var i = 0; i < update.changes.length; i++) {
      const change = update.changes[i]
      const action = change[0]
      const price = change[1]
      const size = change[2]

      const zeroSize = parseFloat(size) === 0
      switch (action) {
        case "buy":
          if (zeroSize) {
            this.bids.remove(price)
          } else {
            this.bids.insert(price, size)
          }

          break;
        case "sell":
          if (zeroSize) {
            this.asks.remove(price)
          } else {
            this.asks.insert(price, size)
          }

          break;
      }
    }
  }

  public processTicker(tick: CBCoinbaseTicker): void {
    this.tick = tick
  }

  public topBids(precision: number): string[][] {
    if (!this.bids) {
      return []
    }
    return this.bids.top(10, precision)
  }

  public topAsks(precision: number): string[][] {
    if (!this.asks) {
      return []
    }
    return this.asks.top(10, precision)
  }

  public lastTick(): CBCoinbaseTicker {
    return this.tick
  }
}

class OrderedValues {
  private values: object = {}
  private set: SortedSet | null = null
  public length = 0
  private setType: string = ""

  public constructor(t: string, init: string[][]) {
    this.setType = t
    this.set = new SortedSet({
      strategy: SortedSet.ArrayStrategy,
      onInsertConflict: SortedSet.OnInsertConflictIgnore,
      comparator: this.setType == "bids" ? this.bidsComparator : this.asksComparator,
    })
    
    for (let x = 0; x < init.length; x++) {
      this.set.insert(init[x][0])
      this.length++
      this.values[init[x][0]] = init[x][1]
    }
  }

  public insert(key: string, value: string): void {
    this.set.insert(key)
    this.values[key] = value
    this.length++
  }

  public remove(key: string): void {
    this.set.remove(key)
    delete this.values[key]
    this.length--
  }

  public top(count: number, precision: number): string[][] {
    if (this.set.length == 0) {
      return []
    }

    let bins = {}
    const exponent = Math.pow(10, precision)

    let cnt = 0
    let iterator = this.set.beginIterator()
    let lastBinKey = 0
    while (iterator.value() != null && cnt <= count) {
      const key = iterator.value()
      const binKeyNum = Math.round(parseFloat(key) * exponent) / exponent
      lastBinKey = binKeyNum
      const binkey = binKeyNum.toFixed(precision)

      if (!!bins[binkey]) {
        bins[binkey] = bins[binkey]+parseFloat(this.values[key])
      } else {
        if (cnt != count){
          bins[binkey] = parseFloat(this.values[key])
          cnt++
        }
      }

      iterator = iterator.next()
    }

    for (let i = cnt+1; i <= count; i++) {
      let nextBin = 0
      if (this.setType === "bids") {
        nextBin = lastBinKey - (1/exponent);
      } else {
        nextBin = lastBinKey + (1/exponent);
      }
      bins[nextBin.toFixed(precision)] = 0
      lastBinKey = nextBin
    }

    return Object.keys(bins).map(k => {
      return [k, bins[k]]
    })
  }

  private bidsComparator(a: string, b: string): number {
    if (a < b) {
      return 1
    }
    if (a > b) {
      return -1
    }
    return 0
  }

  private asksComparator(a: string, b: string): number {
    if (a > b) {
      return 1
    }
    if (a < b) {
      return -1
    }
    return 0
  }
}