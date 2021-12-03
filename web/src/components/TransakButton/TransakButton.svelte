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
        
        if (config != null && !ready) {
          transak = new transakSDK(config)
          ready = true

          // This will trigger when the user closed the widget
          transak.on(transak.EVENTS.TRANSAK_WIDGET_CLOSE, (orderData) => {
            transak.close();
          }); 

          let span = document.getElementsByClassName("transak_close")[0];
          span.onclick = () => {
            return transak.close();
          }; // When the user clicks anywhere outside of the modal, close it

          window.onclick = event => {
            if (event.target === document.getElementById("transak_modal-overlay")) transak.close();
          };

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
