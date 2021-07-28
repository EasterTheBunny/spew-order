<script type="ts">
  import type { Readable } from 'svelte/store'
  import type { AxiosResponse } from "axios"
  import { onMount } from 'svelte'
  import axios from "axios"
  import PriceHistoryChartFactory from '../charts/pricehistory'
  import { getMarketCtx } from '../market'

  let el
  let chart: PriceHistoryChart
  let chartData: CandleItem[]

  const {
    market,
  }: {
    market: Readable<IfcMarket>
  } = getMarketCtx()
  
  export let height = 400

  const resizeDone = () => {
    chart.draw(el.offsetWidth, height)
    chart.update(chartData)
  }

  let timeout
  const reportWindowSize = () => {
    if (!!timeout) {
      clearTimeout(timeout)
    }
    chart.remove()

    // delay the draw until after the resize event is done
    timeout = setTimeout(resizeDone, 1000)
  }
	
  onMount(() => {
    window.onresize = reportWindowSize;

    const unsubscribe = market.subscribe(mkt => {
      if (mkt === null) {
        return
      }

      if (!!chart) {
        chart.remove()
      } else {
        chart = PriceHistoryChartFactory(el)
      }

      chart.draw(el.offsetWidth, height)

      let api = axios.create({
        timeout: 1000,
        headers: {
          "Content-Type": "application/json"
        },
        baseURL: "https://api.pro.coinbase.com",
      })

      const m = mkt.base + "-" + mkt.target
      const marketLookup = {
        "BTC-ETH": "ETH-BTC",
        "BTC-BCH": "BCH-BTC",
      }
      console.log(m)
      console.log(marketLookup[m])

      api.get("/products/"+marketLookup[m]+"/candles?granularity=3600").then((r: AxiosResponse) => {

        let data = r.data.slice(0, 100)
    
        data.sort((a, b) => {
          return a[0] - b[0]
        })

        chartData = data.map((d: number[]) => {
          return {
            time: new Date(d[0]*1000),
            low: d[1],
            high: d[2],
            open: d[3],
            close: d[4],
            volume: d[5],
          }
        })

        chart.update(chartData)
      })
    })

    return () => {
      unsubscribe()
      chart.remove()
    }
  })
</script>

<div bind:this={el} class="chart"></div>