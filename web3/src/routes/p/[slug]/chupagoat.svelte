<script lang="ts" context="module">
  import type { Load } from '@sveltejs/kit'
  import { getProject } from '$lib/api'
  import { authGuard } from '$lib/guards';

  export const load: Load = async ({ stuff }) => {
    const { project } = stuff

    if (!authGuard()) {
      return { status: 302, redirect: `/p/${project.prettyPath}` }
    }

    return { props: { project } }
  }
</script>

<script type="ts">
  import Button, { Label } from '@smui/button';
  import BotButton from '$lib/components/discord'
  import { installDiscordBot } from '$lib/api'

  export let project: Project

  $: inGuild = !!project.discord

</script>

{#if !inGuild && project.permissions.includes('discord.bot.install')}
<BotButton color="primary" variant="raised" />
{/if}

{#if inGuild}
<p>Chupagoat is installed</p>
{/if}
