<script lang="ts" context="module">
  import type { Load } from '@sveltejs/kit'
  import { getProject, getIdentity } from '$lib/api'
  import { isServerSide, authGuard } from '$lib/guards';
  import { auth } from '$lib/state';

  let loggedIn = false;
  auth.subscribe(authState => loggedIn = authState.loggedIn);

  export const load: Load = async ({ params }) => {
    const { slug } = params
    
    let identity: Identity = null
    let project: Project = null
    
    try {
      project = await getProject(slug)
    } catch(e) {
      console.log(e)
      return { status: 404, error: "project not found" }
    }

    if (!isServerSide) {
      try {
        identity = await getIdentity()
      } catch(e) {
        console.log(e)
      }
    }

    return { props: { project, slug, identity }, stuff: { project, identity } }
  }
</script>

<script lang="ts">
  import { setContext } from 'svelte';
  import type { MenuComponentDev } from '@smui/menu';
  import IconButton from '@smui/icon-button';
  import Textfield from '@smui/textfield';

  import NavList from '$lib/components/NavList.svelte';
  import ConnectDialog, { ConnectButton } from '$lib/web3'
  import { page } from '$app/stores';

  let clicked = 0;
  let valueA = "";
  let drawerOn = "";

  export let identity: Identity = null;
  export let project: Project = null;
  export let slug: string

  const routes = [
    { url: `/p/${slug}`, text: 'About', restricted: false, active: $page.url.pathname == `/p/${slug}` },
    { url: `/p/${slug}/chupagoat`, text: 'Chupagoat', restricted: true, active: $page.url.pathname == `/p/${slug}/chupagoat` },
  ];

  if (project.permissions.includes('project.info.update')) {
    routes.push({ url: `/p/${slug}/update`, text: 'Update Project', restricted: true, active: $page.url.pathname == `/p/${slug}/update` })
  }

  setContext('SMUI:list:item:nav', true)
</script>

<div class="d-flex flex-row flex-column-fluid">
  <div class="aside py-9 {drawerOn}">
    <div class="aside-logo flex-column-auto px-9 mb-9">
      <a href="/">
        <img alt="Logo" src="/images/header_logo.png" style="width: 100%;" />
      </a>
    </div>

    <div class="aside-menu flex-column-fluid ps-5 pe-3 mb-9" style="align-items: start;">
      <!-- <div>Left</div> -->
      <div class="left-nav">
        <NavList routes={routes} />
      </div>
    </div>

    <div class="aside-footer flex-column-auto px-9">
      <ConnectButton color="secondary" variant="outlined" text="Connect Wallet" style="width: 100%;" {identity} />
    </div>
  </div>

  <div class="wrapper d-flex flex-column flex-row-fluid" on:click={() => (drawerOn = '')}>
    <div class="header">
      <div class="container d-flex flex-stack flex-wrap gap-2">
        <div class="page-title max d-flex flex-column align-items-start justify-content-center flex-wrap me-lg-2 pb-5 pb-lg-0">
          <h1 class="d-flex flex-column fs-1">
            {project.name}
            <small class="text-muted fs-6">{project.motto}</small>
          </h1>
        </div>

        <div class="d-flex d-lg-none align-items-center ms-n2 me-2">
          <IconButton class="material-icons" on:click$stopPropagation={() => (drawerOn = 'drawer drawer-start drawer-on')}>menu</IconButton>
          <a href="/" class="d-flex align-items-center"><img src="/images/header_logo.png" style="height: 20px;" class="h-20px" /></a>
        </div>

        <div class="d-flex align-items-center flex-shrink-0">
          <!-- 
          <div>
            <Textfield
              class="shaped-outlined"
              variant="outlined"
              bind:value={valueA}
              label="Label"
            ></Textfield>
          </div>
         
          <Button on:click={() => clicked++} variant="raised">
            <Label>Raised</Label>
          </Button>
          -->
        </div>

      </div>
    </div>

    <div class="content d-flex flex-column flex-column-fluid">
      <div class="container-xxl">

        <div class="page-title min d-flex flex-column align-items-start justify-content-center flex-wrap me-lg-2 pb-5 pb-lg-0">
          <h1 class="d-flex flex-column fs-1">
            {project.name}
            <small class="text-muted fs-6">{project.motto}</small>
          </h1>
        </div>

        <slot></slot>
      </div>
    </div>

    <div class="footer py-4 d-flex flex-lg-column">
      <div class="container d-flex flex-column flex-md-row flex-stack">
        <!-- &copy; Cipher Mountain -->
      </div>
    </div>
  </div>

  <!--
  <div class="sidebar">
    <div class="d-flex flex-column sidebar-body px-5 py-10">
      Right
      - takes up right static width @large
      - is drawer toggled by button in middle for smaller than @large
    </div>
  </div>
  -->
</div>

<ConnectDialog />

<style>
  * :global(.left-nav) {
    width: 100%;
  }
</style>
