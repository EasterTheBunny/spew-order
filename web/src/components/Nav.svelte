<script type="ts">
  import { link } from "svelte-navigator"
  import TopAppBar, {
    Row,
    Section,
    Title,
    AutoAdjust,
  } from "@smui/top-app-bar/styled"
  import Drawer, {
    AppContent,
    Content,
    Header,
    Subtitle,
    Scrim,
  } from '@smui/drawer/styled'
  import List, { Item, Text } from '@smui/list/styled'
  import IconButton from "@smui/icon-button/styled"
  import Button, { Label } from '@smui/button/styled'
  import UserMenu from './UserMenu.svelte'
  import { getOidc } from "../oidc"
  import { getLocalization } from '../i18n'
  import { markets } from '../constants'
  import { Currency } from '../constants'

  import TransakButton from './TransakButton'
  
  let topAppBar;
  let dense = true
  let prominent = false
  let open = false

  const { oidc, user, loggedIn } = getOidc()
  const {t} = getLocalization()
  const marketToName = (m: IfcMarket) => m.base+"-"+m.target
  
  let transak_api_key: string = process.env.TRANSAK_API_KEY;
  let transak_env: string = process.env.TRANSAK_ENV;
</script>

<TopAppBar bind:this={topAppBar} {dense} {prominent} >
  <Row>
    <Section>
      <IconButton class="material-icons" on:click={() => open = !open}>menu</IconButton>
      <Title>Cipher Mountain</Title>
    </Section>
    <Section align="end" toolbar>
      {#if $loggedIn}
      <TransakButton label={$t('BuyCrypto')} key={transak_api_key} environment={transak_env} user={$user} />
      <UserMenu />
      {:else}
      <Button on:click={() => oidc.signIn()} variant="unelevated">
        <Label>{$t('Login')}</Label>
      </Button>
      <Button on:click={() => oidc.signIn()} color="secondary" variant="unelevated">
        <Label>{$t('Signup')}</Label>
      </Button>
      {/if}
    </Section>
  </Row>
</TopAppBar>
<AutoAdjust {topAppBar}>
  <Drawer variant="modal" fixed={false} bind:open>
    <Header>
      <Title>{$t('Exchange')}</Title>
      <Subtitle>{$t('ChoosePair')}</Subtitle>
    </Header>
    <Content>
      <List>
        {#each markets as market}
        <Item
          href={"/dashboard/"+marketToName(market)}
          use={[link]}
          on:click={() => open = false}
        >
          <Text>{marketToName(market)}</Text>
        </Item>
        {/each}
      </List>
    </Content>
    <Header>
      <Title>{$t('AccountSummary')}</Title>
    </Header>
    <Content>
      <List>
        <Item
          href={"/dashboard"}
          use={[link]}
          on:click={() => open = false}
        >
          <Text>{$t('Dashboard')}</Text>
        </Item>
        <Item
          href={"/funding"}
          use={[link]}
          on:click={() => open = false}
        >
          <Text>{$t('AddFunds')}</Text>
        </Item>
      </List>
    </Content>
  </Drawer>
  <Scrim fixed={false} />
  <AppContent class="app-content">
    <slot></slot>
  </AppContent>
</AutoAdjust>