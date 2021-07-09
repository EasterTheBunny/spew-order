<script type="ts">
  import type { Writable } from "svelte/store"
  import Button, { Label } from '@smui/button/styled';
  import AmountInputField from "./AmountInputField.svelte"
  import { getDataCtx } from "../exchange"
  import { OrderType, Currency, ActionType } from "../constants"
  
  export let action: ActionType = ActionType.Buy
  export let base: Currency = Currency.Bitcoin
  export let target: Currency = Currency.Ethereum
  export let currentPrice = "0.0"
  export let balances: IfcBalanceResource[] = []

  let total = "0.0000"
  let valid = false
  let order: IfcLimitOrder = {
    name: OrderType.Limit,
    base: base,
    price: currentPrice,
    quantity: "0.000000",
  }

  const {
    orders,
  }: {
    orders: Writable<IfcOrderResource[] | IfcOrderResource>
  } = getDataCtx()
  
  $: validOrder = validate(order, total, action, balanceMap(balances))

  const balanceMap: (b: IfcBalanceResource[]) => object = (b) => {
    const mp = {}
    for (var i = 0; i < b.length; i++) {
      mp[b[i].symbol] = parseFloat(b[i].quantity)
    }
    return mp
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
    order.price = currentPrice
    total = "0.000000"
  }

  const onTotalChange = () => {
    const t = parseFloat(total)
    const p = parseFloat(order.price)
    const q = t / p

    order.quantity = q.toFixed(8)
  }

  const onAmountChange = () => {
    const p = parseFloat(order.price)
    const q = parseFloat(order.quantity)
    const t = q * p

    total = t.toFixed(8)
  }

  const validate: (s: IfcLimitOrder, t: string, a: ActionType, bm: object) => boolean = (s, t, a, bm) => {
    let b = s.base

    switch (a) {
      case ActionType.Buy:
        b = base
        break;
      case ActionType.Sell:
        b = target
        break;
    }

    const max = bm[b]
    const tot = parseFloat(t)

    if (max >= tot && tot > 0) {
      return true
    }

    return false
  }

</script>

<div class="form-section">
  <AmountInputField
    bind:value={order.quantity}
    bind:symbol={target}
    label="amount"
    on:valid={(e) => valid = e.detail}
    on:keyup={onAmountChange} />
</div>

<div class="form-section">
  <AmountInputField
    bind:value={order.price}
    bind:symbol={order.base}
    label="price"
    on:valid={(e) => valid = e.detail}
    on:keyup={onAmountChange} />
</div>

<div class="form-section">
  <AmountInputField
    bind:value={total}
    bind:symbol={order.base}
    label="total"
    on:valid={(e) => valid = e.detail}
    on:keyup={onTotalChange} />
</div>

<div class="form-section">
  <Button
    on:click={onSubmitForm}
    variant="unelevated"
    class="button-shaped-round"
    style="width:100%"
    disabled={!validOrder}
  >
    <Label>Submit Order</Label>
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