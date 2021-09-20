<script type="ts">
  import LayoutGrid, { Cell } from '@smui/layout-grid/styled'
  import CurrencySelect from '../components/CurrencySelect.svelte'
  import Textfield from '@smui/textfield/styled'
  import Icon from '@smui/textfield/icon/styled'
  import CopyClipBoard from '../components/CopyClipBoard.svelte'
  import type { Currency } from '../constants'
  import { getDataCtx } from "../exchange"
  import { getLocalization } from '../i18n'

  export let accountid: string = ""

  let selected: Currency
  const {
    api,
  }: {
    api: ExchangeAPI
  } = getDataCtx()

  async function getHash(c: Currency, a: string) {
    let func = api.getAddressFunc()
    const addr = await func(a, c)
    return addr.address
  }

  const {t} = getLocalization()
  const copyHash = () => {
    const app = new CopyClipBoard({
			target: document.getElementById('clipboard'),
			props: { name: hash },
		});
		app.$destroy();
  }

  $: hash = !!selected ? getHash(selected, accountid) : ""
</script>

<LayoutGrid>
  <Cell span={6}>
    <div>
      <CurrencySelect bind:selected />
    </div>
  
    {#if selected != null}
    {#await hash then value}
    <div style="padding-top: 25px;">
      <Textfield value={value} label="Deposit Address" on:click={copyHash} >
        <Icon class="material-icons" slot="trailingIcon">content_copy</Icon>
      </Textfield>
    </div>
    <p>{$t('CopyAddressInstruction')}</p>
    {/await}
    {/if}
  
  </Cell>
  <Cell span={6}>
    {#if selected != null}
    {#await hash then value}
    <img src="https://chart.googleapis.com/chart?chs=300x300&cht=qr&chl={value}&choe=UTF-8" alt="{selected} deposit address qrcode" />
    {/await}
    {/if}
  </Cell>
</LayoutGrid>