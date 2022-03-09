<script type="ts">
  import { onMount } from 'svelte'
  import Dialog, { Title, Content, Actions } from '@smui/dialog';
  import Button, { Label } from '@smui/button';
  import { auth } from '$lib/state';
  import state from './state'
  import {
    loginWithEthereum,
    promptForAddress,
    checkConnectionStatus,
    WALLET_TITLE,
    WALLET_TEXT,
    AUTH_TITLE,
    AUTH_TEXT,
  } from './util'

  $: connected = $auth.address != '';
  $: title = !connected ? WALLET_TITLE : AUTH_TITLE
  $: text = !connected ? WALLET_TEXT : AUTH_TEXT

  let open = false

  const { dialogOpen, close } = state

  const okBtn = async () => {
    try {
      if (connected) {
        // need to do the authentication sequence
        await loginWithEthereum()
      } else {
        // need to do the connect sequence
        await promptForAddress()
      }
    } catch(e) {
      // TODO: send errors to ui or bubble up
    }
  }

  onMount(() => {
    close()
    checkConnectionStatus()
    return dialogOpen.subscribe((newState) => {

      if (open == newState) {
        open = !open
      }

      open = newState
    })
  })
</script>

<Dialog
  bind:open
  on:SMUIDialog:closed={close}
  aria-labelledby="notice-title"
  aria-describedby="notice-content"
  style="z-index:1000;"
>
  <!-- Title cannot contain leading whitespace due to mdc-typography-baseline-top() -->
  <Title id="notice-title">{title}</Title>
  <Content id="notice-content">{text}</Content>
  <Actions>
    <Button on:click={okBtn} color="secondary">
      <Label>Ok</Label>
    </Button>
    <Button on:click={close} color="secondary">
      <Label>Cancel</Label>
    </Button>
  </Actions>
</Dialog>
