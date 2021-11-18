<script type="ts">
  import AssetCircle from "../components/AssetCircle.svelte"
  import AccessWall from "../components/AccessWall.svelte"
  import { getOidc } from "../oidc"
  import { getDataCtx } from "../exchange"
  import { getLocalization } from '../i18n'

  const { loggedIn } = getOidc()
  const {
    account,
  }: {
    account: Readable<IfcAccountResource>
  } = getDataCtx()
  const {t} = getLocalization()

  const nominals = {
    "ETH": 14,
    "BCH": 105,
    "UNI": 43000,
    "DOGE": 270000,
    "CMTN": 3000000,
    "BTC": 1,
  }

  const getBals: (acc: IfcBalanceResource[], n: any) => AssetItem[] = (acc, n) => {
    console.log(acc)
    return acc.map((b) => {
      const qty = parseFloat(b.quantity)
      return {
        "name": b.symbol,
        "nominal": qty/n[b.symbol],
        "amount": qty,
      }
    }).filter((a) => { return a.amount > 0 })
  }

</script>

<main>
{#if $loggedIn && $account}
  <div class="lead">
    <h3>{$t('AccountAssets')}</h3>
    <h4><i>{$t('NominalTo')} BTC</i></h4>
  </div>
  <AssetCircle chartData={getBals($account.balances, nominals)} />
{:else}
<AccessWall />
{/if}
</main>

<style>
  main {
    padding: 1em;
    max-width: 240px;
    margin: 0 auto;
  }

  @media (min-width: 640px) {
    main {
      max-width: none;
    }
  }

  .lead {
    text-align: center;
  }
</style>