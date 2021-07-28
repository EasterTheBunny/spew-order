<script type="ts">
  import type { Readable } from "svelte/store"
  import { getMarketCtx } from "../market";
  import { getLocalization } from '../i18n'

  const {
    price,
  }: {
    price: Readable<IfcBookProductSpread>
  } = getMarketCtx()
  const {t} = getLocalization()
</script>

<div class="snapshot">
  <h1>BTC-ETH</h1>
  <div>
    <h4>{$t('LastPrice')}</h4>
    <small>{$price.bid}</small>
  </div>
  <div>
    <h4>{$t('DailyChange')}</h4>
    <small>{$price.change24hr}</small>
  </div>
  <div>
    <h4>{$t('DailyRange')}</h4>
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