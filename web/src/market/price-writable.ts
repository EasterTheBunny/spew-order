import type { Writable } from "svelte/store"
import { writable } from "svelte/store"

const initialValue = {
  maxDepth: 0,
  ask: "0.000",
  bid: "0.000",
  change24hr: "",
  range24hr: "",
  asks: [],
  bids: [],
}

const PriceWritable = (): Writable<IfcBookProductSpread> => {
  return writable<IfcBookProductSpread>(initialValue)
}

export default PriceWritable