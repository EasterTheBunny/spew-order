<script lang="ts" context="module">
  import type { Load } from '@sveltejs/kit'
  import { getProject, getMetrics } from '$lib/api'
  import dayjs from 'dayjs'

  export const load: Load = async ({ page }) => {
    const { query } = page

    let projectA: Project = null
    let projectB: Project = null
    let stats: any = null

    const filter = {
      projectSlugs: ["", ""],
      metricTypes: ['HASHTAG'],
      startTime: dayjs().subtract(1, 'day').format('YYYY-MM-DDTHH:mm:ssZ'),
      blockSize: "1d",
    }

    if (!query.has('a') || !query.has('b')) {
      return { status: 404, error: "projects not found" }
    }

    try {
      projectA = await getProject(query.get('a'))
      projectB = await getProject(query.get('b'))

      filter.projectSlugs = [projectA.slug, projectB.slug]
      const metrics = await getMetrics(filter)

      stats = metrics.reduce((obj, m) => {
        if (!obj.hasOwnProperty(m.project.slug)) {
          obj[m.project.slug] = { tweets: m.tweetCount, mentions: m.mentionCount, users: m.userCount }
        }
        return obj
      }, {})
    } catch(e) {
      console.log(e)
      return { status: 404, error: "project not found" }
    }

    return { props: { projectA, projectB, stats } }
  }
</script>

<script type="ts">
  import Hero from '$lib/hero-section';

  export let projectA: Project = null;
  export let projectB: Project = null;
  export let stats: any = [];
  
  const calculateScore = (tweets, mentions, users) => {
    return Math.floor(tweets + (mentions*1.5) + (users*1.2))
  }

  const getScore = (stat) => {
    const { tweets, mentions, users } = stat
    return calculateScore(tweets, mentions, users)
  }

  $: projectAScore = stats.hasOwnProperty(projectA.slug) ? getScore(stats[projectA.slug]) : 0
  $: projectBScore = stats.hasOwnProperty(projectB.slug) ? getScore(stats[projectB.slug]) : 0
  $: sum = projectAScore + projectBScore
</script>

<Hero style="align-items: center; justify-content: center;">
  <div style="width: 70%; text-align: center;">
    <div class="data-element">
      <h1>{projectA.name}</h1>
      <div style="font-size: 1.8rem;"><i>vs</i></div>
      <h1>{projectB.name}</h1>
    </div>
    <div style="display: flex; align-items: stretch; width: 100%;">
      <div style="display: inline; width: {projectAScore/sum*100}%; height: 30px; background-color: #672b7a;"></div>
      <div style="display: inline; width: {projectBScore/sum*100}%; height: 30px; background-color: #ffbb00;"></div>
    </div>
    <div class="data-element">
      <h3>#{projectA.twitter?.hashtag}</h3>
      <h3>#{projectB.twitter?.hashtag}</h3>
    </div>
    <div class="data-element">
      <i>Score: {projectAScore}</i>
      <i>Score: {projectBScore}</i>
    </div>

    <div class="data-element" style="margin-top: 30px; visibility: hidden;">
      <div style="text-align: left;">
        Rankings
        <ul class="ranking-list">
          <li>Player 1</li>
          <li>Player 2</li>
          <li>Player 3</li>
          <li>Player 4</li>
        </ul>
      </div>
      <div style="text-align: center;">
        Combined Rankings
        <ul class="ranking-list">
          <li>Player 3</li>
          <li>Player 1</li>
          <li>Player 2</li>
          <li>Player 4</li>
        </ul>
      </div>
      <div style="text-align: right;">
        Rankings
        <ul class="ranking-list">
          <li>Player 1</li>
          <li>Player 2</li>
          <li>Player 3</li>
          <li>Player 4</li>
        </ul>
      </div>
    </div>

  </div>
</Hero>

<style>
  .data-element {
    display: flex;
    justify-content: space-between;
  }

  .data-element > h3 {
    margin-bottom: 0;
  }

  .ranking-list {
    list-style: none;
    padding-left: 0;
  }
</style>
