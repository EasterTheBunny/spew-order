<script type="ts">
  import FormField from '@smui/form-field';
  import Radio from '@smui/radio';
  import AmountInputField from "./AmountInputField.svelte"
  import { OrderType, Currency, ActionType } from "../constants"

  export let action: ActionType = ActionType.Buy
  export let base: Currency = Currency.Bitcoin
  export let target: Currency = Currency.Ethereum
  export let currentPrice = "0.00000"
  export let balances: IfcBalanceResource[] = []

  let data: IfcMarketOrderRequest = {
    name: OrderType.Market,
    base: base,
    quantity: "0.000000",
  }

  const calcTotal: (b: Currency, q: string) => string = (b, q) => {

    if (b === base) {
      let a = parseFloat(q)
      let b = parseFloat(currentPrice)

      let c = a/b
      return c.toFixed(8)
    }

    return "0"
  }

  const getSymbolValue = (symbol: string): string => {
    let filtered = balances.filter((b) => b.symbol === symbol)
    if (filtered.length === 0) {
      return "0.0000"
    }

    return filtered[0].quantity
  }

  $: if(base||target) {
    data.base = base
  }
  $: amountLabel = data.base === base ? "Total" : "Amount"
  $: total = calcTotal(data.base, data.quantity)
  $: amountHelp = data.base === base ? "you'll receive approx. "+total+data.base : "you will spend approx. "+total+data.base
  $: symbolList = [target, base].filter((v) => {
    if (action === ActionType.Sell && v === base){
      return false
    }
    return true
  })
  $: maximum = getSymbolValue(data.base)

  const validate: (b: Currency, q: string) => boolean = (b, q) => {
    // if the action is a buy, the currency is the base currency, and the amount is greater than the maximum: invalid
    if (b === base) {

    }
    return false
  }
</script>

<div class="form-section">
  {#each symbolList as symbol}
    <FormField>
      <Radio bind:group={data.base} value={symbol} touch on:click={() => console.log("clicked")} />
      <span slot="label">{symbol}</span>
    </FormField>
  {/each}
</div>

<div class="form-section">
  <AmountInputField bind:value={data.quantity} bind:symbol={data.base} bind:label={amountLabel} bind:subtext={amountHelp} />
</div>

<style>
  * :global(.shaped-outlined
      .mdc-notched-outline
      .mdc-notched-outline__leading) {
    border-radius: 28px 0 0 28px;
    width: 28px;
  }
  * :global(.shaped-outlined
      .mdc-notched-outline
      .mdc-notched-outline__trailing) {
    border-radius: 0 28px 28px 0;
  }
  * :global(.shaped-outlined .mdc-notched-outline .mdc-notched-outline__notch) {
    max-width: calc(100% - 28px * 2);
  }
  * :global(.shaped-outlined.mdc-text-field--with-leading-icon:not(.mdc-text-field--label-floating)
      .mdc-floating-label) {
    left: 16px;
  }
  * :global(.shaped-outlined + .mdc-text-field-helper-line) {
    padding-left: 32px;
    padding-right: 28px;
  }
</style>