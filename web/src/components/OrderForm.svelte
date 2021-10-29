<script type="ts">
  import type { Readable } from "svelte/store"
  import { onMount } from 'svelte'
  import OrderSelectAction from "./OrderSelectAction.svelte"
  import OrderSelectType from "./OrderSelectType.svelte"
  import { OrderType, ActionType } from "../constants"
  import { getDataCtx } from "../exchange";
  import { getMarketCtx } from "../market";
  import MarketOrderForm from "./MarketOrderForm.svelte";
  import LimitOrderForm from "./LimitOrderForm.svelte"
 
  let currentPrice = "0.00"  

  let selectedAction: ActionType = ActionType.Buy
  let selectedType: OrderType = OrderType.Market

  const {
    account,
  }: {
    account: Readable<IfcAccountResource>
  } = getDataCtx()

  const {
    price,
    market,
  }: {
    price: Readable<IfcBookProductSpread>,
    market: Readable<IfcMarket>,
  } = getMarketCtx()

  onMount(() => {
    return price.subscribe((s: IfcBookProductSpread) => {
      if (!!s) {
        currentPrice = s.ask
      }
    })
  })

</script>

{#if $account}
  <div class="form-section">
    <OrderSelectAction bind:active={selectedAction} />
  </div>

  <div class="form-section">
    <OrderSelectType bind:value={selectedType} />
  </div>

  {#if selectedType === OrderType.Market}
    <MarketOrderForm bind:action={selectedAction} bind:currentPrice={currentPrice} bind:balances={$account.balances} bind:base={$market.base} bind:target={$market.target} />
  {:else if selectedType === OrderType.Limit}
    <LimitOrderForm bind:action={selectedAction} bind:currentPrice={currentPrice} bind:balances={$account.balances} bind:base={$market.base} bind:target={$market.target} />
  {/if}
{/if}

<style>
  .form-section {
    margin-bottom: 20px;
  }
</style>