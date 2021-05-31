<script type="ts">
  import type { Readable } from "svelte/store"
  import OrderSelectAction from "./OrderSelectAction.svelte"
  import OrderSelectType from "./OrderSelectType.svelte"
  import { OrderType, ActionType } from "../constants"
  import { getDataCtx } from "../exchange";
  import MarketOrderForm from "./MarketOrderForm.svelte";
  import LimitOrderForm from "./LimitOrderForm.svelte"
 
  export let currentPrice = "14.009"  

  let selectedAction: ActionType = ActionType.Buy
  let selectedType: OrderType = OrderType.Market

  let selectedSymbol = "ETH";

  const {
    account,
  }: {
    account: Readable<IfcAccountResource>
  } = getDataCtx()

</script>

{#if $account}
  <div class="form-section">
    <OrderSelectAction bind:selected={selectedAction} />
  </div>

  <div class="form-section">
    <OrderSelectType bind:value={selectedType} />
  </div>

  {#if selectedType === OrderType.Market}
    <MarketOrderForm bind:action={selectedAction} bind:currentPrice={currentPrice} bind:balances={$account.balances} />
  {:else if selectedType === OrderType.Limit}
    <LimitOrderForm />
  {/if}
{/if}

<style>
  .form-section {
    margin-bottom: 20px;
  }
</style>