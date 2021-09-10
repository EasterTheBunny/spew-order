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
    user,
  }: {
    oidc: OidcService
    loggedIn: Readable<boolean>
    user: Readable<User>
  } = getOidc()

  const redirector = derived([loggedIn, user], ([$l, $u]) => {
    if ($l) {

      pendo.initialize({
        visitor: {
          id: $u.profile.sid,
          email: $u.profile.email,
          full_name: $u.profile.name,
        },
        account: {
          id: $u.profile.sid,
          name: $u.profile.name,
          is_paying: true,
          // monthly_value:// Recommended if using Pendo Feedback
        }
      });

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