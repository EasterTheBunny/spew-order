<script type="ts">
  import { onMount } from 'svelte'
  import FormField from '@smui/form-field';
  import Checkbox from '@smui/checkbox';
  import Textfield from '@smui/textfield';
  import HelperText from '@smui/textfield/helper-text';
  import CharacterCounter from '@smui/textfield/character-counter';
  import Button, { Label } from '@smui/button';
  import { auth } from '$lib/state';
  import { state, util } from '$lib/web3';
  import { createProject, updateProject } from '$lib/api';
  import { goto } from '$app/navigation';

  export let extended = false
  export let showConnect = true 
  export let redirect = true 
  export let cancel: () => any = null
  export let complete: () => any = () => {}
  export let project: Project = {
    slug: "",
    name: "",
    motto: "",
    description: "",
    prettyPath: "",
    permissions: [],
  }

  const { open } = state
  const { loginWithEthereum } = util

  $: connected = $auth.address != ''
  $: loggedIn = $auth.loggedIn
  $: disabled = project.name == "" || project.description == ""

  const formValid: (p: any) => boolean = (p) => {
    return p.name && p.description
  }

  const buttonClk = async () => {
    if (project.name == "" || project.description == "") {
      return
    }

    try {
      if (!loggedIn) {
        // sign the request and get a token
        await loginWithEthereum()
      }
      console.log(project)

      const projectInput: NewProjectInput = {
        name: project.name,
        prettyPath: project.prettyPath,
        description: project.description,
      }

      if (projectInput.prettyPath.length == 0) {
        projectInput.prettyPath = null
      } else if (projectInput.prettyPath.length < 8) {
        throw new Error("Pretty URL must be at minimum 8 characters")
      }

      if (project.slug == "") {
        // post new data to server
        project = await createProject(projectInput)
      } else {
        project = await updateProject(project.slug, projectInput)
      }

      complete()
      if (redirect) {
        goto(`/p/${project.prettyPath}`)
      }
    } catch(e) {
      console.log(e)
      // TODO: handle errors
    }
  }

  const connectCheck = () => {
    if (!connected) {
      open()
    }
  }

  onMount(() => {
    if (!project) {
      project = {
        slug: "",
        name: "",
        motto: "",
        description: "",
        prettyPath: "",
        permissions: [],
      }
    }
  })
</script>

{#if showConnect}
<FormField style="width: 100%;">
  <Checkbox on:click={connectCheck} bind:checked={connected} bind:disabled={connected} />
  <span slot="label">
    Connect your wallet
  </span>
</FormField>
{/if}

<Textfield
  color="secondary"
  variant="filled"
  bind:value={project.name}
  updateInvalid
  label="Project Name"
  input$maxlength={32}
  style="width: 100%;margin-top: 20px;"
  required
>
  <svelte:fragment slot="helper">
    <HelperText>Project Name</HelperText>
    <CharacterCounter>0 / 32</CharacterCounter>
  </svelte:fragment>
</Textfield>

{#if extended}
<Textfield
  color="secondary"
  variant="filled"
  bind:value={project.prettyPath}
  updateInvalid
  label="Pretty URL"
  input$maxlength={36}
  style="width: 100%;margin-top: 20px;"
>
  <svelte:fragment slot="helper">
    <HelperText>Pretty URL</HelperText>
    <CharacterCounter>0 / 36</CharacterCounter>
  </svelte:fragment>
</Textfield>
{/if}

<Textfield
  textarea
  color="secondary"
  variant="filled"
  input$maxlength={1000}
  bind:value={project.description}
  updateInvalid
  label="Project Description"
  style="width: 100%;margin-top: 20px;"
  required
>
  <CharacterCounter slot="internalCounter">0 / 1000</CharacterCounter>
</Textfield>

{#if !!cancel}
<Button on:click={cancel} color="primary" style="margin-top:20px;">
  <Label>Cancel</Label>
</Button>
{/if}

<Button on:click={buttonClk} color="secondary" style="margin-top:20px;" {disabled}>
  <Label>Ready Set GO!</Label>
  <i class="material-icons" aria-hidden="true">arrow_forward</i>
</Button>
