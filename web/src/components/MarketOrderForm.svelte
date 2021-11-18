<script type="ts">
  import type { Writable } from "svelte/store"
  import FormField from '@smui/form-field/styled';
  import Radio from '@smui/radio/styled';
  import Button, { Label } from '@smui/button/styled';
  import AmountInputField from "./AmountInputField.svelte"
  import { getDataCtx } from "../exchange"
  import { OrderType, Currency, ActionType } from "../constants"
  import { calcTotal } from "../util"
  import { getLocalization } from '../i18n'

  export let action: ActionType = ActionType.Buy
  export let base: Currency = Currency.Bitcoin
  export let target: Currency = Currency.Ethereum
  export let currentPrice = "0.0"
  export let balances: IfcBalanceResource[] = []

  let amountInputGreaterThan0 = false
  let order: IfcMarketOrder = {
    name: OrderType.Market,
    base: base,
    quantity: "0.000000",
  }

  const {
    orders,
  }: {
    orders: Writable<IfcOrderResource[] | IfcOrderResource>
  } = getDataCtx()
  const {t} = getLocalization()

  const balanceMap: (b: IfcBalanceResource[]) => object = (b) => {
    const mp = {}
    for (var i = 0; i < b.length; i++) {
      mp[b[i].symbol] = parseFloat(b[i].quantity)
    }
    return mp
  }

  $: if(action === ActionType.Sell) {
    order.base = target
  } else {
    order.base = base
  }
  $: amountLabel = order.base === base ? "Total" : "Amount"
  $: symbolList = action === ActionType.Sell ? [target] : [base]
  $: amountHelp = ((order: IfcMarketOrder, action: ActionType, price: string) => {
    let helpTxt = ""

    if (action == ActionType.Buy) {
      if (order.base === base) {
        let amt = parseFloat(order.quantity)
        let priceFloat = parseFloat(currentPrice)

        amt = (amt * priceFloat)
        helpTxt = "you will receive approx. "+amt.toFixed(6)+" "+target
      }
    }

    return helpTxt
  })(order, action, currentPrice)

  $: validOrder = validate(order, balanceMap(balances)) && amountInputGreaterThan0

  const validate: (s: IfcMarketOrder, bm: object) => boolean = (s, bm) => {
    // if the action is a buy, the currency is the base currency, and the amount is greater than the maximum: invalid
    const total = calcTotal(s, action, currentPrice, base, target)
    let b = s.base

    switch (action) {
      case ActionType.Buy:
        if (s.base === target) {
          b = base
        }
        break;
      case ActionType.Sell:
        if (s.base === base) {
          b = target
        }
        break;
    }

    const max = bm[b]

    if (max >= total) {
      return true
    }

    return false
  }

  const onSubmitForm = () => {
    const o: IfcOrderResource = {
      guid: "",
      status: null,
      order: {
        base,
        target,
        action,
        type: order,
      },
    }
    orders.set(o)
    order.quantity = "0.000000"
  }
</script>

<div class="form-section" style="display:none;">
  {#each symbolList as symbol}
    <FormField>
      <Radio bind:group={order.base} value={symbol} touch />
      <span slot="label">{symbol}</span>
    </FormField>
  {/each}
</div>

<div class="form-section">
  <AmountInputField
    bind:value={order.quantity}
    bind:symbol={order.base}
    bind:label={amountLabel}
    bind:subtext={amountHelp}
    on:valid={(e) => amountInputGreaterThan0 = e.detail} />
</div>

<div class="form-section">
  <Button
    on:click={onSubmitForm}
    variant="unelevated"
    class="button-shaped-round"
    style="width:100%"
    disabled={!validOrder}
  >
    <Label>{$t('SubmitOrder')}</Label>
  </Button>
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