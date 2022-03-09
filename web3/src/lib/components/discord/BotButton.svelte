<script type="ts">
  import Button, { Label } from '@smui/button';
  import { Button } from '@smui/common/elements';
  import { installDiscordBot } from '$lib/api'
  import { auth } from '$lib/state';
  import { state, util } from '$lib/web3';

  export let extended = true

  const { open } = state
  const { loginWithEthereum, promptForAddress } = util

  $: connected = $auth.address != ''
  $: loggedIn = $auth.loggedIn

  const handleClick = async (e) => {
    // ensure wallet is connected before moving on
    if (!loggedIn) {
      if (!connected) {
        await promptForAddress()
      }

      await loginWithEthereum()
    }

    if (extended) {
      window.location.href = await installDiscordBot(extended);
    } else {
      window.open(await installDiscordBot(extended), '_blank');
    }

    return
  }
</script>

<Button on:click={handleClick} {...$$restProps}>
  <Label>Install Chupagoat</Label>
</Button>
