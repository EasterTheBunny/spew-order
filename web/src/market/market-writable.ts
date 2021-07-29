import type { Writable } from "svelte/store"
import { writable } from "svelte/store"

const MarketWritable = (worker: Worker): Writable<IfcMarket> => {
  const { subscribe, set, update } = writable<IfcMarket>(null)

  const setMarket = (newMarket: IfcMarket) => {
    worker.postMessage({type: 'subscribe', market: newMarket})
    set(newMarket)
  }

  const updateMarket = (updater: (newMarket: IfcMarket) => IfcMarket) => {
    update((currentValue) => {
      const nextValue = updater(currentValue)

      if (nextValue !== null) {
        worker.postMessage({type: 'subscribe', market: nextValue})
      } else {
        worker.postMessage({type: 'unsubscribe'})
      }
      return nextValue
    })
  }

  return {
    subscribe,
    set: setMarket,
    update: updateMarket,
  }
}

export default MarketWritable