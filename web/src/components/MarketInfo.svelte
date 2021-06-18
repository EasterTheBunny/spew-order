<div class="market-wrapper">
  <header class="market-header">
    <h2 class="mdc-typography--headline6" style="margin: 0;">
      {$price.bid}
    </h2>
    <h3 class="mdc-typography--subtitle2" style="margin: 0; color: #888;">
      BTC-ETH 
    </h3>
  </header>
  <article class="market-main">
    <PriceHistory bind:height={chartHeight} />
  </article>
  <aside class="market-book">
    <OrderBook src="asks" color="green" yAxis={false} bind:height={bookHeight} />
    <OrderBook src="bids" color="red" yAxis={false} bind:height={bookHeight} />
  </aside>
</div>

<script type="ts">
  import type { Readable } from "svelte/store"
  import OrderBook from "../components/OrderBook.svelte"
  import PriceHistory from "../components/PriceHistory.svelte"
  import { getDataCtx } from "../exchange";

  let bookHeight = 250
  let chartHeight = 500
  
  const {
    price,
  }: {
    price: Readable<IfcBookProductSpread>
  } = getDataCtx()
</script>

<style>
</style>