<script lang="ts" context="module">
  import type { Load } from '@sveltejs/kit'

  export const load: Load = async ({ stuff }) => {
    const { project } = stuff

    return { props: { project } }
  }
</script>

<script lang="ts">
  export let project: Project
</script>

<p>{project.description}</p>
