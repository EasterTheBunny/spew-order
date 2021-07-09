<script type="ts">
  import type { Readable } from "svelte/store"
  import { Link } from "svelte-routing"
  import { getDataCtx } from "../exchange"

  const {
    account,
  }: {
    account: Readable<IfcAccountResource>
  } = getDataCtx()

</script>

<dl class="balance-list">
  {#each $account.balances as balance}
  <dt>{balance.symbol}</dt>
  <dd>{balance.quantity} <i>{balance.symbol}</i></dd>
  {/each}
</dl>
<Link to="funding">Add Funds</Link>

<style lang="scss">
  .balance-list {
    width: 100%;
    overflow: hidden;
    padding: 0;
    margin: 0
  }

  .balance-list > dt {
    float: left;
    width: 50%;
    padding: 0;
    margin: 0
  }

  .balance-list > dd {
    float: left;
    text-align: right;
    width: 50%;
    padding: 0;
    margin: 0
  }
</style>