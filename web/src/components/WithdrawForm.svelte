<script type="ts">
  import type { Writable } from "svelte/store"
  import LayoutGrid, { Cell } from '@smui/layout-grid/styled';
  import Button, { Label } from '@smui/button/styled';
  import AddressInputField from "../components/AddressInputField.svelte"
  import AmountInputField from '../components/AmountInputField.svelte'
  import CurrencySelect from '../components/CurrencySelect.svelte'
  import { balanceMap } from '../util'
  import { getDataCtx } from "../exchange"
  
  export let balances: IfcBalanceResource[] = []

  let req: IfcTransactionRequest = {
    symbol: null,
    address: "",
    quantity: "",
  }

  let validAddress = false
  let validQuantity = false
  const {
    transactions,
  }: {
    transactions: Writable<IfcTransactionRequest>
  } = getDataCtx()

  $: formValid = validAddress && validQuantity && validate(req, balanceMap(balances))

  const validate: (r: IfcTransactionRequest, bm: object) => boolean = (r, bm) => {
    const max = bm[req.symbol]
    const tot = parseFloat(r.quantity)

    if (max >= tot && tot > 0) {
      return true
    }

    return false
  }

  const onSubmitForm = () => {
    if (formValid) {
      transactions.set(req)
      req.address = ""
      req.quantity = ""
    }
  }
</script>

<LayoutGrid>
  <Cell span={12}>
    <div>
      <CurrencySelect bind:selected={req.symbol} />
    </div>
  
    {#if req.symbol != null}
    <div style="padding-top: 25px;">
      <AddressInputField
        bind:value={req.address}
        bind:currency={req.symbol}
        on:valid={(e) => validAddress = e.detail} />
    </div>

    <div style="padding-top: 25px;">
      <AmountInputField
        bind:value={req.quantity}
        bind:symbol={req.symbol}
        label="total"
        on:valid={(e) => validQuantity = e.detail} />
    </div>

    <div style="padding-top: 25px;">
      <Button
        on:click={onSubmitForm}
        variant="unelevated"
        class="button-shaped-round"
        style="width:100%"
        disabled={!formValid}
      >
        <Label>Submit Withdrawal</Label>
      </Button>
    </div>

    {/if}
  </Cell>
</LayoutGrid>