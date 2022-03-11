<script lang="ts" context="module">
  import type { Load } from '@sveltejs/kit'
  import { isServerSide } from '$lib/guards'
  import { registerDiscordOauth } from '$lib/api'
  import { getIdentity } from '$lib/api'

  export const load: Load = async ({ page, stuff }) => {
    const { query } = page

    const guild = query.get('guild_id')
    const permissions = query.get('permissions')
    const code = query.get('code')
    const state = query.get('state')

    let id: Identity = null

    if (stuff.hasOwnProperty('identity')) {
      id = stuff.identity
    }

    if (!isServerSide && id == null) {
      try {
        console.log("getting current identity")
        id = await getIdentity()
      } catch(e) {
        console.log(e)
      }
    }

    return { props: { guild, permissions, code, state, identity: id }}
  }
</script>

<script type="ts">
  // import BotButton from '$lib/components/discord'
  import { onMount } from 'svelte'
  import Select, { Option } from '@smui/select';
  import Button, { Label } from '@smui/button';
  import { session } from '$app/stores'
  import { auth } from '$lib/state';
  import Hero from '$lib/hero-section';
  import BotButton from '$lib/components/discord'
  import ProjectForm from '$lib/components/forms';
  import { util } from '$lib/web3';
  import { registerDiscordBot } from '$lib/api';
  import { goto } from '$app/navigation';

  export let guild = ""
  export let permissions = ""
  export let code = ""
  export let state = ""
  export let identity: Identity = null

  let registerAuthCode = false
  let project: Project = {
    slug: "",
    name: "",
    motto: "",
    description: "",
    permissions: [],
  }
  let guildFromQuery = !!identity && !!identity.discordUser ? identity.discordUser.guilds.find(e => e.id == guild) : null
  let selectedGuild: DiscordGuild = !!identity && !!identity.discordUser ? (typeof guildFromQuery !== "undefined" ? guildFromQuery : null) : null
  let createProject = false

  const { promptForAddress, loginWithEthereum } = util

  const projectFilter: (p: Project) => boolean = (p) => {
    if (p.discord == null) {
      return false
    }

    if (!p.permissions.includes("discord.bot.install")) {
      return false
    }

    return true
  }

  $: connected = $auth.address != ''
  $: loggedIn = $auth.loggedIn
  $: projects = !!identity ? identity.projects : []
  $: guilds = !!identity && !!identity.discordUser ? identity.discordUser?.guilds : []

  const showFormBtn = () => {
    createProject = true
    project = {
      slug: "",
      name: "",
      motto: "",
      description: "",
      permissions: [],
    }
  }

  const run = async () => {
    if (!connected) {
      if (code == "" || state == "") {
        return
      }

      try {
        await promptForAddress()
      } catch(e) {
        // TODO: handle errors
        console.log(e)
        return
      }
    }

    if (!loggedIn) {
      if (code == "" || state == "") {
        return
      }

      // prompt login
      try {
        await loginWithEthereum()
      } catch(e) {
        // TODO: handle errors
        console.log(e)
        return
      }
    }

    if (loggedIn) {
      if (!identity) {
        try {
          const id = await getIdentity()
          identity = {
            ...id,
          }

          // select the guild again
          selectedGuild = !!identity.discordUser ? (typeof guildFromQuery !== "undefined" ? guildFromQuery : null) : null
        } catch(e) {
          // TODO: handle errors
          console.log(e)
          return
        }
      }

      if ($session.oauth.retryOauthClient) {
        // if the code has already been used, the oauth request will fail
        // fall back to the old oauth if possible
        try {
          const newAuth = await registerDiscordOauth(state, code)
          identity.discordUser = newAuth
          identity = {
            ...identity
          }
        } catch(e) {
          console.log(identity)
          if (!identity.discordUser) {
            return
          }

          if (!identity.discordUser.oauthToken) {
            return
          }

          // TODO: handle errors
          console.log(e)
        }
      }

      const perms = parseInt(permissions, 10)
      if (!!selectedGuild && !isNaN(perms) && !!project && project.slug != "") {
        await registerDiscordBot(project.slug, selectedGuild.id, perms)
        goto(`/p/${project.prettyPath}/chupagoat`)
      }
    }
  }

  const cancelFunc = () => {
    createProject = false
    project = {
      slug: "",
      name: "",
      motto: "",
      description: "",
      permissions: [],
    }
  }

  onMount(() => {
    run()
  })

</script>

<Hero style="align-items: center; justify-content: center;">
  <div style="width: 70%;">

    {#if !loggedIn && code == "" && state ==""}
    <BotButton color="secondary" variant="outlined" />
    {/if}

    <div>
      <Select
        variant="outlined"
        bind:value={selectedGuild}
        on:select={run}
        label="Discord Server"
        style="width: 100%;"
      >
        <Option value={null}>Select a discord</Option>
        {#each guilds as g}
          <Option value={g}>{g.name}</Option>
        {/each}
      </Select>
    </div>

    {#if !createProject}
    {#if projects.length > 0}
    <div style="margin-top: 20px;">
      <Select
        variant="outlined"
        bind:value={project}
        on:select={run}
        label="Project"
        style="width: 100%;"
      >
        <Option value={null} />
        {#each projects as p}
          <Option value={p}>{p.name}</Option>
        {/each}
      </Select>
    </div>
    {/if}

    <div style="margin-top: 20px;">
      <Button on:click={showFormBtn} color="secondary" variant="outlined" style="margin-right: 10px;">
        <Label>Create a Project</Label>
      </Button>

      {#if !!selectedGuild && !!project && project.slug != ""}
      <Button on:click={run} color="secondary" variant="outlined" style="margin-right: 10px;">
        <Label>Install</Label>
      </Button>
      {/if}
    </div>
    {/if}

    {#if createProject}
    <div style="margin-top: 20px">
      <ProjectForm bind:project redirect={false} complete={run} cancel={cancelFunc} showConnect={false} />
    </div>
    {/if}
  </div>
</Hero>
