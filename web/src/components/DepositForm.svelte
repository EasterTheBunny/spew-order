<script type="ts">
  import LayoutGrid, { Cell } from '@smui/layout-grid/styled'
  import CurrencySelect from '../components/CurrencySelect.svelte'
  import Textfield from '@smui/textfield/styled'
  import Icon from '@smui/textfield/icon/styled'
  import CopyClipBoard from '../components/CopyClipBoard.svelte'
  import type { Currency } from '../constants'
  import { getLocalization } from '../i18n'

  export let balances: IfcBalanceResource[] = []

  let selected: Currency

  const getHash: (c: Currency, b: IfcBalanceResource[]) => string = (c, b) => {
    for (let x = 0; x < b.length; x++) {
      if (b[x].symbol === c) {
        return b[x].funding
      }
    }
    return ""
  }

  const {t} = getLocalization()
  const copyHash = () => {
    const app = new CopyClipBoard({
			target: document.getElementById('clipboard'),
			props: { name: hash },
		});
		app.$destroy();
  }

  $: hash = !!selected ? getHash(selected, balances) : ""
</script>

<LayoutGrid>
  <Cell span={6}>
    <div>
      <CurrencySelect bind:selected />
    </div>
  
    {#if selected != null}
    <div style="padding-top: 25px;">
      <Textfield bind:value={hash} label="Deposit Address" on:click={copyHash} >
        <Icon class="material-icons" slot="trailingIcon">content_copy</Icon>
      </Textfield>
    </div>
    <p>{$t('CopyAddressInstruction')}</p>
    {/if}
  
  </Cell>
  <Cell span={6}>
    {#if selected != null}
    <img src="https://chart.googleapis.com/chart?chs=300x300&cht=qr&chl={hash}&choe=UTF-8" alt="{selected} deposit address qrcode" />
    {/if}
  </Cell>
</LayoutGrid>