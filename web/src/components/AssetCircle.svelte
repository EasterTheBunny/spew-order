<script type="ts">
  import { onMount } from 'svelte'
  import AssetChartFactory from '../charts/assets'

  let el
  let chart: AssetChart
  export let chartData: AssetItem[] = []

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

    if (!!chart) {
      chart.remove()
    } else {
      chart = AssetChartFactory(el)
    }

    chart.draw(height)
    chart.update(chartData)

    return () => {
      chart.remove()
    }
  })
</script>

<div bind:this={el} class="asset-chart"></div>

<style>

  .asset-chart {
    height: 500px;
    margin-top: 50px;
  }

  .asset-chart > svg {
    width: 100%;
    height: 100%;
  }
</style>