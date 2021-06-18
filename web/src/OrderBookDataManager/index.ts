import SortedSet from 'js-sorted-set'

export class OrderBookDataManager {
  private bids: OrderedValues
  private asks: OrderedValues

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

      switch (action) {
        case "buy":
          if (size === "0.00000000") {
            this.bids.remove(price)
          } else {
            this.bids.insert(price, size)
          }

          break;
        case "sell":
          if (size === "0.00000000") {
            this.asks.remove(price)
          } else {
            this.asks.insert(price, size)
          }

          break;
      }
    }
  }

  public topBids(): string[][] {
    if (!this.bids) {
      return []
    }
    return this.bids.top(20)
  }

  public topAsks(): string[][] {
    if (!this.asks) {
      return []
    }
    return this.asks.top(20)
  }
}

class OrderedValues {
  private values: object = {}
  private set: SortedSet | null = null

  public constructor(t: string, init: string[][]) {
    this.set = new SortedSet({
      onInsertConflict: SortedSet.OnInsertConflictIgnore,
      comparator: t == "bids" ? this.bidsComparator : this.asksComparator,
    })
    
    for (let x = 0; x < init.length; x++) {
      this.set.insert(init[x][0])
      this.values[init[x][0]] = init[x][1]
    }
  }

  public insert(key: string, value: string): void {
    this.set.insert(key)
    this.values[key] = value
  }

  public remove(key: string): void {
    this.set.remove(key)
    delete this.values[key]
  }

  public top(count: number): string[][] {
    if (this.set.length == 0) {
      return []
    }

    let bins = {}
    const exponent = Math.pow(10, 3)

    let cnt = 0
    let iterator = this.set.beginIterator()
    while (iterator.value() != null && cnt <= count) {
      const key = iterator.value()
      const binkey = (Math.round(parseFloat(key) * exponent) / exponent).toFixed(3)

      if (!!bins[binkey]) {
        bins[binkey] = bins[binkey]+parseFloat(this.values[key])
      } else {
        if (cnt != count){
          bins[binkey] = parseFloat(this.values[key])
        }
        cnt++
      }

      iterator = iterator.next()
    }

    return Object.keys(bins).map(k => {
      return [k, bins[k].toFixed(3)]
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