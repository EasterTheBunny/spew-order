<script type="ts">
  import { Router, Route } from "svelte-routing"
  import Nav from "../components/Nav.svelte"
  import Landing from "../views/Landing.svelte"
  import Login from "../views/Login.svelte"
  import Dashboard from "../views/Dashboard.svelte"
  import Funding from "../views/Funding.svelte"

  export let url: string
</script>

<Router url={url}>
  <Nav>
    <Route path="funding"><Funding /></Route>
    <Route path="dashboard"><Dashboard /></Route>
    <Route path="login"><Login /></Route>
    <Route path="/"><Landing /></Route>
  </Nav>
</Router>