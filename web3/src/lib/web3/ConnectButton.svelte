<script type="ts">
  import { setContext } from 'svelte';
  import Button, { Group, Label, Icon } from '@smui/button';
  import type { MenuComponentDev } from '@smui/menu';
  import Menu from '@smui/menu';
  import { Anchor } from '@smui/menu-surface';

  import List, {
    Item,
    Separator,
    Text,
  } from '@smui/list';

  import { auth } from '$lib/state';
  import state from './state'
  import {
    updateTokenInAuth,
    shortenAddress,
  } from './util'

  export let text = 'Connect';
  export let variant = "outlined";
  export let color = "secondary";
  export let nav = false;
  export let withLogin = false;
  export let identity: Identity = null;

  let originalText = text;
  let menu: MenuComponentDev;
  let anchor: HTMLDivElement;
  let anchorClasses: { [k: string]: boolean } = { 'mdc-menu-surface--anchor': true };
  const { open } = state

  $: connected = $auth.address != '';
  $: text = $auth.address != '' ? shortenAddress($auth.address) : originalText;
  $: loggedIn = $auth.loggedIn

  setContext('SMUI:list:item:nav', true)
</script>

<Group style="display: flex; justify-content: stretch;">

<div
  class={Object.keys(anchorClasses).join(' ')}
  use:Anchor={{
    addClass: (className) => {
      if (!anchorClasses[className]) {
        anchorClasses[className] = true;
      }
    },
    removeClass: (className) => {
      if (anchorClasses[className]) {
        delete anchorClasses[className];
        anchorClasses = anchorClasses;
      }
    },
  }}
  bind:this={anchor}
  style="width: 100%"
>
  <Button on:click={() => connected ? menu.setOpen(true) : open()} variant={variant} color={color} class={nav?'':'button-right'} style="width: 100%;">
    <Label>{text}</Label>
  </Button>
  <Menu
    bind:this={menu}
    anchor={false}
    bind:anchorElement={anchor}
    anchorCorner="BOTTOM_LEFT"
    style=" margin-top: 15px;"
  >
    <List>
      {#if loggedIn && !!identity}
      {#each identity.projects as project}
      <Item href={`/p/${project.prettyPath}`}>
        <Text>{project.name}</Text>
      </Item>
      {/each}
      <Separator />
      {/if}
      {#if loggedIn}
      <Item on:SMUI:action={() => {updateTokenInAuth('')}}>
        <Text>Sign Out</Text>
      </Item>
      {:else}
      <Item on:SMUI:action={open}>
        <Text>Sign In</Text>
      </Item>
      {/if}
    </List>
  </Menu>
</div>
</Group>
