<script type="ts">
  import type { Readable } from "svelte/store"
  import { onMount } from "svelte"
  import { derived } from "svelte/store"
  import type { User } from "oidc-client"
  import { navigate } from "svelte-navigator"
  import { getOidc } from "../oidc"

  const {
    oidc,
    loggedIn,
  }: {
    oidc: OidcService
    loggedIn: Readable<boolean>
  } = getOidc()

  const redirector = derived([loggedIn], ([$l]) => {
    if ($l) {
      navigate("/", { replace: true })
    }
    return ""
  })

  onMount(() => {
    let u: Promise<User> = oidc.manager.signinRedirectCallback()
  })
</script>

{$redirector}
{#if $loggedIn}
You should be redirected. Click <a href="/">here</a> if not.
{/if}