<script type="ts">
  import type { Readable } from "svelte/store"
  import OrderSelectAction from "./OrderSelectAction.svelte"
  import OrderSelectType from "./OrderSelectType.svelte"
  import { OrderType, ActionType } from "../constants"
  import { getDataCtx } from "../exchange";
  import MarketOrderForm from "./MarketOrderForm.svelte";
  import LimitOrderForm from "./LimitOrderForm.svelte"
 
  let currentPrice = "0.00"  

  let selectedAction: ActionType = ActionType.Buy
  let selectedType: OrderType = OrderType.Market

  let selectedSymbol = "ETH";

  const {
    account,
    price,
  }: {
    account: Readable<IfcAccountResource>
    price: Readable<IfcBookProductSpread>
  } = getDataCtx()

  price.subscribe((s: IfcBookProductSpread) => {
    if (!!s) {
      currentPrice = s.ask
    }
  })

</script>

{#if $account}
  <div class="form-section">
    <OrderSelectAction bind:selected={selectedAction} />
  </div>

  <div class="form-section">
    <div class="mdc-typography--subtitle2"><small>Available to trade:</small></div>
    {#each $account.balances as balance}
    <div class="mdc-typography--caption"><small><b>{balance.symbol}</b>: {balance.quantity} {balance.symbol}</small></div>
    {/each}
  </div>

  <div class="form-section">
    <OrderSelectType bind:value={selectedType} />
  </div>

  {#if selectedType === OrderType.Market}
    <MarketOrderForm bind:action={selectedAction} bind:currentPrice={currentPrice} bind:balances={$account.balances} />
  {:else if selectedType === OrderType.Limit}
    <LimitOrderForm bind:action={selectedAction} bind:currentPrice={currentPrice} bind:balances={$account.balances} />
  {/if}
{/if}

<style>
  .form-section {
    margin-bottom: 20px;
  }
</style>