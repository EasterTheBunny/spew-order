<script type="ts">
  import type { Readable } from "svelte/store"
  import Paper, { Title, Content } from "@smui/paper";
  import LayoutGrid, { Cell } from "@smui/layout-grid";
  import OrderForm from "../components/OrderForm.svelte"
  import OrderList from "../components/OrderList.svelte"
  import { getOidc } from "../oidc"
  import { getDataCtx } from "../exchange";

  let elevation = 1;
  let color = 'default';

  const { loggedIn } = getOidc()
  const {
    account,
  }: {
    account: Readable<IfcAccountResource>
  } = getDataCtx()
</script>

{#if $loggedIn && $account}
<div class="paper-container">
  <LayoutGrid>
  
  <Cell span={9}>
    <Paper transition {elevation} {color} class="paper-demo">
      <Title>Order List</Title>
      <Content>
        <OrderList />
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
  }
</style>