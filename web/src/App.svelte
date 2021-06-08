<script lang="ts">
	import Root from "./views/Root.svelte"
	import Router from "./views/Router.svelte"
  import oidc from "./oidc"
  import initExchangeAPI from "./exchange"
  import { initLocalizationContext } from './i18n';

  export let client_id: string = "DcMwCcm9VNE3xMz6Sxtde8FqdXH8Berq"
  export let authority: string = "https://dev-xfscxtiv.us.auth0.com"

  initLocalizationContext()
  const user = oidc({
    client_id: client_id,
    authority: authority,
  })
  initExchangeAPI(user)

</script>

<Root>
	<Router url="" />
</Root>