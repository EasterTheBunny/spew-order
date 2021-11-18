<script type="ts">
  import type { Readable } from "svelte/store"
  import type { User } from "oidc-client"

  import { onMount } from "svelte"
  import transakSDK from "@transak/transak-sdk"
  import Button, { Label } from '@smui/button/styled'

  import { TransakEnvironment, getTransakConfig } from "./constants"
  import { Currency } from "../../constants"
  import { getDataCtx } from "../../exchange"

  export let label = "Buy with Transak"
  export let key: string = ""
  export let environment: TransakEnvironment = TransakEnvironment.STAGING
  export let user: Readable<User> = null
  export let redirect: string = "https://app.ciphermtn.com/dashboard"

  let ready = false
  let config = null
  let transak: transakSDK
  let accountid: string = ""

  const purchaseList: Currency[] = [ Currency.Ethereum, Currency.Bitcoin, Currency.Dogecoin ]
  const {
    api,
    account,
  }: {
    api: ExchangeAPI,
    account: Readable<IfcAccountResource>
  } = getDataCtx()

  const openTransak = () => {
    if (!!transak) {
      transak.init()
    }
  }

  onMount(() => {
    const unsubscribe = account.subscribe(async (acc) => {
      if (!! acc && acc.id != accountid) {
        accountid = acc.id
        config = await getTransakConfig(user, acc, api, purchaseList, key, environment, redirect)
        
        if (config != null) {
          transak = new transakSDK(config)
          //ready = true // TODO: turn on for production

          transak.on(transak.EVENTS.TRANSAK_ORDER_SUCCESSFUL, (orderData) => {
            transak.close();
          })
        }
      }

    })

    return () => unsubscribe()
  })

</script>

{#if ready}
<Button on:click={openTransak} color="secondary" variant="unelevated">
  <Label>{label}</Label>
</Button>
{/if}
