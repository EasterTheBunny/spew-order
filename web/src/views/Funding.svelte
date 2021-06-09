<script type="ts">
  import type { Readable } from "svelte/store"
  import Tab, { Label } from '@smui/tab'
  import TabBar from '@smui/tab-bar'
  import Paper, { Title, Subtitle, Content } from '@smui/paper';
  import Select, { Option } from '@smui/select';
  import Textfield from '@smui/textfield';
  import Icon from '@smui/textfield/icon';
  import LayoutGrid, { Cell } from '@smui/layout-grid';
  import { Currency } from '../constants';
  import CopyClipBoard from '../components/CopyClipBoard.svelte'
  import { getOidc } from "../oidc"
  import { getDataCtx } from "../exchange";

  let active = "Deposit"
  const { loggedIn } = getOidc()
  const {
    account,
  }: {
    account: Readable<IfcAccountResource>
  } = getDataCtx()

  const currencies = [{
    value: Currency.Ethereum,
    name: 'Ethereum',
  }, {
    value: Currency.Bitcoin,
    name: 'Bitcoin',
  }]

  const getHash: (c: Currency) => string = (c) => {
    for (let x = 0; x < $account.balances.length; x++) {
      if ($account.balances[x].symbol === c) {
        return $account.balances[x].funding
      }
    }
    return ""
  }

  let selectedCurrency: Currency = null
  $: hash = !!selectedCurrency ? getHash(selectedCurrency) : ""

  const copyHash = () => {
    const app = new CopyClipBoard({
			target: document.getElementById('clipboard'),
			props: { name: hash },
		});
		app.$destroy();
  }
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

        <LayoutGrid>
          <Cell span={6}>
            <div>
              <Select bind:value={selectedCurrency} label="Select Currency">
                {#each currencies as c}
                  <Option value={c.value}>{c.name}</Option>
                {/each}
              </Select>
            </div>
          
            {#if selectedCurrency != null}
            <div style="padding-top: 25px;">
              <Textfield bind:value={hash} label="Deposit Address" on:click={copyHash} >
                <Icon class="material-icons" slot="trailingIcon">content_copy</Icon>
              </Textfield>
            </div>
            <p>
              Copy the deposit address above or scan the code to the right with your wallet app to transfer
              funds.
            </p>
            {/if}
          
          </Cell>
          <Cell span={6}>
            {#if selectedCurrency != null}
            <img src="https://chart.googleapis.com/chart?chs=300x300&cht=qr&chl={hash}&choe=UTF-8" alt="{selectedCurrency} deposit address qrcode" />
            {/if}
          </Cell>
        </LayoutGrid>


        {:else if active === 'Withdraw'}
        <div>
          Withdrawing funds is currently unavailable
        </div>
        {/if}


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