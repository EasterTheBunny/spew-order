<script lang="ts" context="module">
  import type { Load } from '@sveltejs/kit'
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
  import LayoutGrid, { Cell } from '@smui/layout-grid';
  import Paper, { Title, Content } from '@smui/paper';
  import dayjs from 'dayjs'
  import BotButton from '$lib/components/discord';
  import { TwitterBotForm } from '$lib/components/forms';
  import LineChart from '$lib/charts/line';
  import { getMetrics } from '$lib/api';

  export let project: Project
  export let chartData: ChartItem[] = [];

  $: inGuild = !!project.discord
  $: inTwitter = !!project.twitter && project.twitter.connected

  let open = false

  getMetrics({ projectSlugs: [project.slug], metricTypes: ['HASHTAG'], startTime: '2022-03-13T14:00:00+00:00', blockSize: '5m' }).then((data) => {
    chartData = data.map((d) => {
      return {
        time: dayjs(d.startTime).toDate(),
        value: d.tweetCount,
      }
    })
    console.log(data)
  })

</script>


<LayoutGrid>
    {#if !inGuild}
  <Cell spanDevices={{ desktop: 12, tablet: 12, phone: 12 }}>
    <Paper color="primary" elevation={12}>
      <Title>Discord</Title>
      <Content>
        The Discord extension is the foundation of the Chupagoat game. Use slash commands to run game actions.<br/><br/>
        Installing in Discord is free to all projects and comes complete with all game functions.

        <div style="display: flex; justify-content: space-between; margin-top: 30px;">
          <BotButton color="primary" variant="raised" label={inGuild ? "Installed" : "Install"} disabled={inGuild || !project.permissions.includes('discord.bot.install')} />
        </div>

      </Content>
    </Paper>
  </Cell>
    {/if}

    {#if !inTwitter}
  <Cell spanDevices={{ desktop: 12, tablet: 12, phone: 12 }}>
    <Paper color="secondary" elevation={12}>
      <Title>Twitter</Title>
      <Content>
        Configure the bot to listen for a custom hashtag. When that hashtag is used by a player, the bot responds with game results. Set up loot boxes and run loot campaigns to drive activity in your project!<br/><br/>
        The Twitter exension is an exclusive feature that currently requires a whitelist spot for access. Collect Chupagoat loot to get whitelisted!

        <div style="display: flex; justify-content: space-between; margin-top: 30px;">
          <Button color="primary" variant="raised" on:click={() => (open = true)} disabled={inTwitter || !project.permissions.includes('twitter.bot.install')}>
            <Label>{inTwitter ? "Connected" : "Connect"}</Label>
          </Button>
        </div>

      </Content>
    </Paper>

  </Cell>
  {/if}

  <Cell spanDevices={{ desktop: 12, tablet: 12, phone: 12 }} style="display:none;">
    <Paper elevation={0}>
      <Content>
        <LineChart bind:chartData />
      </Content>
    </Paper>
  </Cell>
</LayoutGrid>

<TwitterBotForm bind:open bind:project />

