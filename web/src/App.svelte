<script lang="ts">
	import Root from "./views/Root.svelte"
	import Router from "./views/Router.svelte"
  import oidc from "./oidc"
  import initExchangeAPI from "./exchange"
  import { initLocalizationContext } from './i18n';

  export let client_id: string = process.env.AUTH0_CLIENTID;
  export let authority: string = process.env.AUTH0_DOMAIN;

  console.log(client_id)

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