<script>
  import TopAppBar, {
    Row,
    Section,
    Title,
    AutoAdjust,
  } from "@smui/top-app-bar";
  import IconButton from "@smui/icon-button";
  import Button, { Label } from '@smui/button';
  import { navigate } from "svelte-routing"
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
  
<style>
  /* Hide everything above this component. */
  :global(app, body, html) {
    display: block !important;
    height: auto !important;
    width: auto !important;
    position: static !important;
  }
</style>