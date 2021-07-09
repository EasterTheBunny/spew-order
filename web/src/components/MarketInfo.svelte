<div class="market-wrapper">
  <article class="market-main">
    <PriceHistory bind:height={chartHeight} />
  </article>
</div>

<script type="ts">
  import PriceHistory from "../components/PriceHistory.svelte"

  let chartHeight = 500
</script>

<style lang="scss">
  .market-wrapper {
    display: inline-flex;
    flex-flow: row wrap;
    height: 500px;
    width: 100%;
    justify-content: right;
  }

  .market-main {
    height: 100%;
    background: var(--mdc-theme-background, #fff);
  }

  .market-main {
    flex: 1;
  }
</style>