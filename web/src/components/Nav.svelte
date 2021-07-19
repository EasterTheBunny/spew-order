<script type="ts">
  import { navigate } from "svelte-routing"
  import TopAppBar, {
    Row,
    Section,
    Title,
    AutoAdjust,
  } from "@smui/top-app-bar/styled";
  import IconButton from "@smui/icon-button/styled"
  import Button, { Label } from '@smui/button/styled'
  import UserMenu from './UserMenu.svelte'
  import { getOidc } from "../oidc";
  import { getLocalization } from '../i18n';
  
  let topAppBar;
  let dense = true;
  let prominent = false;

  const { oidc, loggedIn } = getOidc()
  const {t} = getLocalization()
</script>

<TopAppBar bind:this={topAppBar} {dense} {prominent} >
  <Row>
    <Section>
      <IconButton class="material-icons">menu</IconButton>
      <Title>Cipher Mountain</Title>
    </Section>
    <Section align="end" toolbar>
      {#if $loggedIn}
      <Button on:click={() => navigate("/dashboard", { replace: true })} color="secondary" variant="outlined">
        <Label>{$t('Exchange')}</Label>
      </Button>
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
  <slot></slot>
</AutoAdjust>