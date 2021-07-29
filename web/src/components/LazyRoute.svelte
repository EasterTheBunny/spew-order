<script>
  import { Route } from "svelte-navigator";
  import Lazy from "./Lazy.svelte";

  export let component;
  export let delayMs = null;

  let props;
  $: {
    const { component, ...restProps } = $$props;
    props = restProps;
  }
</script>

<Route {...props}>
  <Lazy {component} {delayMs}>
    <slot />
  </Lazy>
</Route>
