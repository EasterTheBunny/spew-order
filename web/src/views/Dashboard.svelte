<script type="ts">
  import type { Readable } from "svelte/store"
  import Paper, { Title, Content } from "@smui/paper/styled";
  import LayoutGrid, { Cell } from "@smui/layout-grid/styled";
  import OrderForm from "../components/OrderForm.svelte"
  import OrderList from "../components/OrderList.svelte"
  import MarketInfo from "../components/MarketInfo.svelte"
  import SnapshotInfo from "../components/SnapshotInfo.svelte"
  import AccountSummary from "../components/AccountSummary.svelte"
  import OrderBook from "../components/OrderBook.svelte"
  import AccessWall from "../components/AccessWall.svelte"
  import { getOidc } from "../oidc"
  import { getDataCtx } from "../exchange"
  import { getMarketCtx } from "../market"
  import { getLocalization } from '../i18n'
  import { onMount } from "svelte"
  import { validMarket } from "../constants"

  let elevation = 1
  let color = 'default'
  let bookHeight = 250
  export let market: IfcMarket = null

  const { loggedIn } = getOidc()
  const {
    account,
  }: {
    account: Readable<IfcAccountResource>
  } = getDataCtx()

  const mkt = getMarketCtx().market
  const {t} = getLocalization()

  $: {
    if (market !== null && validMarket(market)) {
      mkt.update(() => market)
    }
  }

  onMount(() => {
    if (market !== null && validMarket(market)) {
      mkt.update(() => market)
    }

    return () => mkt.update(() => null)
  })
</script>

{#if $loggedIn && $account}
<div class="paper-container">
  <LayoutGrid>
    <Cell span={3}>
      <Paper transition {elevation} {color} class="paper-demo">
        <Title>{$t('AccountSummary')}</Title>
        <Content>
          <AccountSummary />
        </Content>
      </Paper>
      
      <Paper transition {elevation} {color} class="paper-demo">
        <Content>
          <div class="book-header">
            <span style="width: 50%;">{$t('Price')}</span>
            <span>{$t('Quantity')}</span>
          </div>
          <OrderBook src="asks" name="asks" yAxis={false} bind:height={bookHeight} />
          <OrderBook src="bids" name="bids" yAxis={false} bind:height={bookHeight} />
        </Content>
      </Paper>
    </Cell>

    <Cell span={6}>
      <Paper transition {elevation} {color} class="paper-demo">
        <Content>
          <SnapshotInfo />
        </Content>
      </Paper>

      <Paper transition {elevation} {color} class="paper-demo">
        <Content>
          <MarketInfo />
        </Content>
      </Paper>
    </Cell>
    
    <Cell span={3}>
      <Paper transition {elevation} {color} class="paper-demo">
        <Title>{$t('CreateNewOrder')}</Title>
        <Content>
          <OrderForm />
        </Content>
      </Paper>

      <Paper transition {elevation} {color} class="paper-demo">
        <Title>{$t('AllPositions')}</Title>
        <Content>
          <OrderList />
        </Content>
      </Paper>
    </Cell>
  </LayoutGrid>
</div>
{:else}
<AccessWall />
{/if}

<style lang="scss">
  .paper-container {
    background-color: var(--mdc-theme-background, #f8f8f8);
    border: 1px solid
      var(--mdc-theme-text-hint-on-background, rgba(0, 0, 0, 0.1));
  }

  .book-header {
    display: flex;
    flex-flow: row nowrap;
    justify-content: flex-start;
  }

  * :global(.paper-demo) {
    margin: 0 auto;
    margin-bottom: 15px;
  }

  * :global(.market-book-text) {
    font-size: 11px;
  }

  * :global(.market-book-depth-asks) {
    fill: #672b7a;
  }

  * :global(.market-book-depth-bids) {
    opacity: 0.5;
    fill: #ffbb00;
  }
</style>