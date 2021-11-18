<script type="ts">
  import { Router, Route } from "svelte-navigator"
  import Nav from "../components/Nav.svelte"
  import LazyRoute from "../components/LazyRoute.svelte"
  import Landing from "../views/Landing.svelte"
  import Login from "../views/Login.svelte"

  export let url: string

  const Funding = () => import("../views/Funding.svelte")
  const Exchange = () => import("../views/Exchange.svelte")
  const Accounts = () => import("../views/Account.svelte")
</script>

<Router url={url}>
  <Nav>
    <LazyRoute path="funding" component={Funding} delayMs={500}>
      <h4>Loading Funding...</h4>
    </LazyRoute>
    <LazyRoute path="dashboard" component={Accounts} delayMS={500}>
      <h4>Loading Dashboard...</h4>
    </LazyRoute>
    <LazyRoute path="dashboard/*" component={Exchange} delayMS={500}>
      <h4>Loading Exchange...</h4>
    </LazyRoute>
    <Route path="login"><Login /></Route>
    <Route path="/"><Landing /></Route>
  </Nav>
</Router>