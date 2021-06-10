<script type="ts">
  import type { Readable } from "svelte/store"
  import Tab, { Label } from '@smui/tab'
  import TabBar from '@smui/tab-bar'
  import Paper, { Content } from '@smui/paper';
  import TransactionList from '../components/TransactionList.svelte'
  import DepositForm from '../components/DepositForm.svelte'
  import WithdrawForm from "../components/WithdrawForm.svelte";
  import { getOidc } from "../oidc"
  import { getDataCtx } from "../exchange";

  let active = "Deposit"

  const { loggedIn } = getOidc()
  const {
    account,
  }: {
    account: Readable<IfcAccountResource>
  } = getDataCtx()

</script>

{#if $loggedIn && $account}
<div>

  <div class="paper-container">
    <Paper class="paper-demo">
      <Content>

        <TabBar tabs={['Deposit', 'Withdraw']} let:tab bind:active>
          <!-- Note: the `tab` property is required! -->
          <Tab {tab}>
            <Label>{tab}</Label>
          </Tab>
        </TabBar>

        {#if active === 'Deposit'}
        <DepositForm bind:balances={$account.balances} />
        {:else if active === 'Withdraw'}
        <WithdrawForm bind:balances={$account.balances} />
        {/if}

      </Content>
    </Paper>
  </div>

  <div class="paper-container">
    <Paper class="paper-demo">
      <Content>
        <TransactionList />
      </Content>
    </Paper>
  </div>
</div>
{/if}
<div id="clipboard"></div>

<style>
  .paper-container {
    padding: 36px 18px;
    background-color: var(--mdc-theme-background, #f8f8f8);
    border: 1px solid
      var(--mdc-theme-text-hint-on-background, rgba(0, 0, 0, 0.1));
  }

  * :global(.paper-demo) {
    margin: 0 auto;
    max-width: 800px;
  }

  #clipboard {
    display: none;
  }
</style>