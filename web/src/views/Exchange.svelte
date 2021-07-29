<script lang="ts">
  import { Router, Route } from 'svelte-navigator'
  import Dashboard from './Dashboard.svelte'
  import { marketFromString } from '../constants'
</script>

<Router>
  <Route path=":market" let:params>
    <Dashboard market={marketFromString(params.market)} />
  </Route>
  <Route path="/"><h1>No market selected</h1></Route>
</Router>