<script lang="ts" context="module">
  import type { Load } from '@sveltejs/kit'
  import { isServerSide } from '$lib/guards'
  import { getIdentity } from '$lib/api'
  import { auth } from '$lib/state';

  let loggedIn = false;
  auth.subscribe(authState => loggedIn = authState.loggedIn);

  export const load: Load = async ({ stuff }) => {
    let id: Identity = null

    if (!isServerSide && loggedIn) {
      try {
        id = await getIdentity()
      } catch(e) {
        console.log(e)
      }
    }

    return { props: { identity: id }, stuff: { identity: id }}
  }
</script>

<script lang="ts">
  import { setContext } from 'svelte';
  import type { MenuComponentDev } from '@smui/menu';
  import Menu from '@smui/menu';
  import { Anchor } from '@smui/menu-surface';
  import List, {
    Item,
    Separator,
    Text,
  } from '@smui/list';

  import TopAppBar, {
    Row,
    Section,
    Title,
    AutoAdjust,
    TopAppBarComponentDev,
  } from '@smui/top-app-bar';
  import Button, { Label } from '@smui/button';
  import Dialog, { Title, Content, Actions } from '@smui/dialog';
  import LayoutGrid, { Cell } from '@smui/layout-grid';
  import ConnectDialog, { ConnectButton } from '$lib/web3'
  import IconButton, { Icon } from '@smui/icon-button';
  import { Svg } from '@smui/common/elements';
  import { mdiDiscord, mdiTwitter } from '@mdi/js';

  let topAppBar: TopAppBarComponentDev;
  let open = false;
  let navMenuOpen = false;

  let menu: MenuComponentDev;
  let anchor: HTMLDivElement;
  let anchorClasses: { [k: string]: boolean } = { 'mdc-menu-surface--anchor': true };

  export let identity: Identity = null;

  setContext('SMUI:list:item:nav', true)
</script>

<svelte:head>
  <meta name="robots" content="index, follow" />
</svelte:head>  

<TopAppBar bind:this={topAppBar} variant="short">
  <Row>
    <Section>
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
      >
        <IconButton class="material-icons" on:click={() => {navMenuOpen = !navMenuOpen; menu.setOpen(navMenuOpen)}} size="mini">menu</IconButton>
        <Menu
          bind:this={menu}
          anchor={false}
          bind:anchorElement={anchor}
          anchorCorner="BOTTOM_LEFT"
          on:SMUIMenuSurface:closed={() => navMenuOpen = false}
        >

          <List>
            <Item href={`/chupagoat`}>
              <Text>Chupagoat <i>ARG</i></Text>
            </Item>
            <Item href={`/p`}>
              <Text>Projects</Text>
            </Item>
            <Item href={`/whitepaper.pdf`} target="_blank">
              <Text>Whitepaper</Text>
            </Item>
            <Item on:click={() => (open = true)}>
              <Text>Mint</Text>
            </Item>
          </List>

        </Menu>
      </div>
      <Title>
        <Button href="/">
          <img src="/images/header_logo.png" style="height: 35px;" class="h-20px" />
        </Button>
      </Title>

      <!--
      <Button href="#staking" variant="outlined" class="button-left">
        <Label>Stake</Label>
      </Button>
      -->
      <Button href="/whitepaper.pdf" target="_blank" variant="outlined" class="button-left nav-bar-extra-link">
        <Label>Whitepaper</Label>
      </Button>
      <!--
      <Button on:click={() => clicked++} variant="outlined" class="button-left">
        <Label>Roadmap</Label>
      </Button>
      -->
    </Section>
    <Section align="end" toolbar>
      <Button on:click={() => (open = true)} variant="outlined" class="button-right nav-bar-extra-link">
        <Label>Mint</Label>
      </Button>

      <ConnectButton color="secondary" variant="outlined" text="Connect" {identity} />
    </Section>
  </Row>
</TopAppBar>

<AutoAdjust {topAppBar}>
  <slot></slot>

  <LayoutGrid class="main-section">
    <Cell span={12}>
      <div style="text-align: center">
        <IconButton mini href="https://discord.gg/6gfNxC9Hj5" target="_blank" ripple={false}>
          <Icon component={Svg} viewBox="0 0 24 24">
            <path fill="currentColor" d={mdiDiscord} />
          </Icon>
        </IconButton>
        <IconButton mini href="https://twitter.com/CipherMountain" target="_blank" ripple={false}>
          <Icon component={Svg} viewBox="0 0 24 24">
            <path fill="currentColor" d={mdiTwitter} />
          </Icon>
        </IconButton>
      </div>
      <p style="text-align: center;font-size: 1.0rem;">
        <small>(c) 2022 Cipher Mountain LLC</small>
      </p>
    </Cell>
  </LayoutGrid>
</AutoAdjust>

<ConnectDialog />
<Dialog
  bind:open
  aria-labelledby="simple-title"
  aria-describedby="simple-content"
>
  <!-- Title cannot contain leading whitespace due to mdc-typography-baseline-top() -->
  <Title id="simple-title">Minting Notice</Title>
  <Content id="simple-content">Pre-sale mint begins on May 3rd 2022. Public sale begins May 4th 2022.</Content>
  <Actions>
    <Button on:click={() => (open = false)}>
      <Label>Ok</Label>
    </Button>
  </Actions>
</Dialog>
