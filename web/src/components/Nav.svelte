<script>
  import TopAppBar, {
    Row,
    Section,
    Title,
    AutoAdjust,
  } from "@smui/top-app-bar";
  import IconButton from "@smui/icon-button";
  import Button, { Label } from '@smui/button';
  import { getOidc } from "../oidc";
  
  let topAppBar;
  let dense = true;
  let prominent = false;
  let clicked = 0;

  const { oidc, loggedIn } = getOidc()
</script>

<TopAppBar bind:this={topAppBar} {dense} {prominent} >
  <Row>
    <Section>
      <IconButton class="material-icons">menu</IconButton>
      <Title>Cipher Mountain</Title>
    </Section>
    <Section align="end" toolbar>
      {#if $loggedIn}
      <Button color="secondary" on:click={() => clicked++} variant="outlined">
        <Label>Exchange</Label>
      </Button>
      {:else}
      <Button on:click={() => oidc.signIn()} variant="unelevated">
        <Label>Login</Label>
      </Button>
      <Button color="secondary" on:click={() => clicked++} variant="unelevated">
        <Label>Signup</Label>
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