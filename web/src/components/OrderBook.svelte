<script type="ts">
  import type { Readable } from "svelte/store"
  import { onMount } from 'svelte'
  import PriceDepthChartFactory from '../charts/pricedepth'
  import { getMarketCtx } from "../market";

  const {
    price,
  }: {
    price: Readable<IfcBookProductSpread>
  } = getMarketCtx()

  let el
  let chart: PriceDepthChart

  export let src = "asks"
  export let name = "asks"
  export let yAxis = true
  export let height = 200

  const resizeDone = () => {
    chart.draw(el.offsetWidth, height, name, yAxis)
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
    
  window.onresize = reportWindowSize;

	onMount(() => {
    chart = PriceDepthChartFactory(el)
    chart.draw(el.offsetWidth, height, name, yAxis)

    price.subscribe((b) => {
      let x: PriceDepthItem[]
      if (src === "asks") {
        x = b.asks.map((item) => {
          return {
            price: item[0],
            depth: Math.ceil(parseFloat(item[1])),
          }
        }).reverse()
      } else {
        x = b.bids.map((item) => {
          return {
            price: item[0],
            depth: Math.ceil(parseFloat(item[1])),
          }
        })
      }

      chart.update(x, b.maxDepth)
    })
	});
</script>

<div bind:this={el} class="chart"></div>