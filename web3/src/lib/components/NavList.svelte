<script type="ts">
  import { onMount, setContext } from 'svelte';
  import List, { Item, Text } from '@smui/list';
  import { navigating, session } from '$app/stores';
  import { auth } from '$lib/state';

  export let routes = [];

  $: authenticated = $auth.loggedIn;

  navigating.subscribe((n) => {
    if (!!n) {
      for (let i = 0; i < routes.length; i++) {
        if (n.to.path == routes[i].url) {
          routes[i].active = true
        } else {
          routes[i].active = false
        }
      }
    }
  })

  session.subscribe(s => {
    authenticated = !!s.user?.token && s.user?.token != '';
  })

  setContext('SMUI:list:item:nav', true)
</script>

<List>
  {#each routes as route}
  {#if !route.restricted || authenticated}
  <Item href={route.url} activated={route.active} color="primary">
    <Text>{route.text}</Text>
  </Item>
  {/if}
  {/each}
</List>
