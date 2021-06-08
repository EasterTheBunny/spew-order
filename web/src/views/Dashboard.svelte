<script type="ts">
  import type { Readable } from "svelte/store"
  import Paper, { Title, Content } from "@smui/paper";
  import LayoutGrid, { Cell } from "@smui/layout-grid";
  import List, { Item, Text, PrimaryText, SecondaryText } from "@smui/list";
  import OrderForm from "../components/OrderForm.svelte"
  import OrderList from "../components/OrderList.svelte"
  import { getOidc } from "../oidc"
  import { getDataCtx } from "../exchange";

  let elevation = 1;
  let color = 'default';
  let clicked = '';


  const { loggedIn } = getOidc()
  const {
    account,
    orders,
  }: {
    account: Readable<IfcAccountResource>
    orders: Readable<IfcOrderResource[]>
  } = getDataCtx()

  const fullName = (symbol: string): string => {
    switch(symbol) {
      case "BTC":
        return "Bitcoin"
      case "ETH":
        return "Ethereum"
      default:
        return ""
    }
  }
</script>

{#if $loggedIn && $account}
<div class="paper-container">
  <LayoutGrid>

  <Cell span={3}>
    <Paper transition {elevation} {color} class="paper-demo">
      <Title>Assets</Title>
      <Content>
        <List class="demo-list">
          {#each $account.balances as balance}
          <Item on:SMUI:action={() => (clicked = balance.symbol)}>
            <Text>
              <PrimaryText>{fullName(balance.symbol)}</PrimaryText>
              <SecondaryText>{balance.quantity} {balance.symbol}</SecondaryText>
            </Text>
          </Item>
          {/each}
        </List>
      </Content>
    </Paper>
  </Cell>
  
  <Cell span={6}>
    <Paper transition {elevation} {color} class="paper-demo">
      <Title>{clicked} Order List</Title>
      <Content>
        <OrderList orders={$orders} />
      </Content>
    </Paper>
  </Cell>
  
  <Cell span={3}>
    <Paper transition {elevation} {color} class="paper-demo">
      <Title>Create New Order</Title>
      <Content>
        <OrderForm />
      </Content>
    </Paper>
  </Cell>

  </LayoutGrid>
</div>
{/if}

<style>
  .paper-container {
    background-color: var(--mdc-theme-background, #f8f8f8);
    border: 1px solid
      var(--mdc-theme-text-hint-on-background, rgba(0, 0, 0, 0.1));
  }
  * :global(.paper-demo) {
    margin: 0 auto;
    max-width: 600px;
  }
  * :global(.demo-list) {
    max-width: 600px;
    border: 1px solid
      var(--mdc-theme-text-hint-on-background, rgba(0, 0, 0, 0.1));
  }
</style>