<script type="ts">
  import type { Readable } from "svelte/store"
  import { getDataCtx } from "../exchange";

  const {
    price,
  }: {
    price: Readable<IfcBookProductSpread>
  } = getDataCtx()
</script>

<div class="snapshot">
  <h1>BTC-ETH</h1>
  <div>
    <h4>Last Price</h4>
    <small>{$price.bid}</small>
  </div>
  <div>
    <h4>24hr change</h4>
    <small>{$price.change24hr}</small>
  </div>
  <div>
    <h4>24hr range</h4>
    <small>{$price.range24hr}</small>
  </div>
</div>

<style lang="scss">
  .snapshot {
    display: flex;
    flex-flow: row nowrap;
    justify-content: space-around;
  }
</style>