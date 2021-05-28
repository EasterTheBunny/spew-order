<script lang="ts">
	import Root from "./views/Root.svelte"
	import Router from "./views/Router.svelte"
  import oidc from "./oidc"

  export let client_id: string = "0qx7yV52wFZyuVOcXHbSkXF3OckMApXx"
  export let authority: string = "https://dev-xfscxtiv.us.auth0.com"

  oidc({
    client_id: client_id,
    authority: authority,
  })

</script>

<Root>
	<Router url="" />
</Root>