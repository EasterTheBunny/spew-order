<script lang="ts" context="module">
  import type { Load } from '@sveltejs/kit'
  import { getProject } from '$lib/api'
  import { authGuard } from '$lib/guards';

  export const load: Load = async ({ stuff }) => {
    const { project } = stuff

    if (!authGuard()) {
      return { status: 302, redirect: `/p/${project.prettyPath}` }
    }

    if (!project.permissions.includes('project.info.update')) {
      return { status: 302, redirect: `/p/${project.prettyPath}`}
    }

    return { props: { project }, maxage: 0 }
  }
</script>

<script type="ts">
  import ProjectForm from '$lib/components/forms'

  export let project: Project
</script>

<ProjectForm showConnect={false} {project} extended={true} />