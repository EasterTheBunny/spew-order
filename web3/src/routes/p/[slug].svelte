<script lang="ts" context="module">
  import type { Load } from '@sveltejs/kit'

  export const load: Load = async ({ page, fetch }) => {
    const { slug } = page.params
    const response = await fetch(`/p/${slug}.json`)
    // return nothing if site was not found to fall through to __error.svelte
    if (response.ok) return { props: { project: await response.json() } }
  }
</script>

<script lang="ts">
  export let project: Project
</script>

<a href="/" class="back" sveltekit:prefetch>&laquo; back</a>
<h2>{project.slug}</h2>