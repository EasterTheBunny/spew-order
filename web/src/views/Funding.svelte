<script type="ts">
  import type { Readable, Writable } from "svelte/store"
  import Tab, { Label } from '@smui/tab/styled'
  import TabBar from '@smui/tab-bar/styled'
  import Paper, { Content } from '@smui/paper/styled';
  import TransactionList from '../components/TransactionList.svelte'
  import DepositForm from '../components/DepositForm.svelte'
  import WithdrawForm from "../components/WithdrawForm.svelte";
  import MarketOrderForm from "../components/MarketOrderForm.svelte";
  import CurrencySelect from "../components/CurrencySelect.svelte";
  import { ActionType, Currency } from "../constants"
  import { getOidc } from "../oidc"
  import { getDataCtx } from "../exchange";

  let active = "Buy Tokens"
  let selectedCurrency: Currency = Currency.Ethereum;

  const prices = [{
    currency: Currency.Bitcoin,
    price: "3000000.0",
  },{
    currency: Currency.Ethereum,
    price: "250000.0",
  }]

  const currencies = [{
    value: Currency.Ethereum,
    name: 'Ethereum',
  }, {
    value: Currency.Bitcoin,
    name: 'Bitcoin',
  }]
  
  const { loggedIn } = getOidc()
  const {
    account,
  }: {
    account: Readable<IfcAccountResource>
  } = getDataCtx()

  $: activePrice = prices.find((a) => a.currency == selectedCurrency)
  $: priceMsg = ((currency: Currency) => {
    const active = prices.find((a) => a.currency == currency)
    const name = currencies.find((a) => a.value == currency)

    let cmtn = 100.0
    let price = parseFloat(active.price)

    return cmtn + " CMTN for " + (cmtn/price).toFixed(6) + " " + name.name
  })(selectedCurrency)
</script>

{#if $loggedIn && $account}
<div>

  <div class="paper-container">
    <Paper class="paper-demo">
      <Content>

        <TabBar tabs={['Buy Tokens', 'Deposit', 'Withdraw']} let:tab bind:active>
          <!-- Note: the `tab` property is required! -->
          <Tab {tab}>
            <Label>{tab}</Label>
          </Tab>
        </TabBar>

        {#if active === 'Buy Tokens'}
        <h2>Cipher Mountain Token</h2>
        <p>CMTN is the native token used for flat fee trading. Buy with Bitcoin or Ethereum.</p>
        <p><i>{priceMsg}</i></p>
        <div class="form-section">
          <CurrencySelect bind:selected={selectedCurrency} currencies={currencies} />
        </div>

        <div class="form-section">
          <MarketOrderForm bind:base={selectedCurrency} bind:balances={$account.balances} target={Currency.CipherMtn} action={ActionType.Buy} bind:currentPrice={activePrice.price} />
        </div>
        {:else if active === 'Deposit'}
        <DepositForm bind:accountid={$account.id} />
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

  .form-section {
    margin-bottom: 20px;
  }

  * :global(.paper-demo) {
    margin: 0 auto;
    max-width: 800px;
  }

  #clipboard {
    display: none;
  }
</style>