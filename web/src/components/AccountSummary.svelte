<script type="ts">
  import type { Readable } from "svelte/store"
  import { Link } from "svelte-navigator"
  import { getDataCtx } from "../exchange"
  import { getLocalization } from '../i18n'

  const {
    account,
  }: {
    account: Readable<IfcAccountResource>
  } = getDataCtx()
  const {t} = getLocalization()

  const formatQuantity = (quantity: string): string => {
    const decimal = quantity.indexOf(".")
    const places = 4;

    if (decimal < 0) {
      for (let i = 0; i < places; i++) {
        quantity+=" ";
      }
    } else {
      let end = 5;
      if (quantity.length >= decimal+end) {
        end = decimal+end;
      } else {
        end = quantity.length;
      }
      quantity = quantity.slice(0, end);
    }

    return quantity;
  }

</script>

<dl class="balance-list">
  {#each $account.balances as balance}
  <dt>{balance.symbol}</dt>
  <dd>{formatQuantity(balance.quantity)}</dd>
  {/each}
</dl>
<Link to="/funding">{$t('AddFunds')}</Link>

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