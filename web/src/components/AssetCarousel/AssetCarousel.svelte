<script type="ts">
  import Carousel from '@beyonk/svelte-carousel';
  import { ChevronLeftIcon, ChevronRightIcon } from 'svelte-feather-icons';
  import Card from './Card.svelte';
  import { assetColorBindings } from './constants';

  export let assets = [];
</script>

<Carousel perPage={1} dots={false}>
  <span class="control" slot="left-control">
    <ChevronLeftIcon />
  </span>
  {#each assets as asset}
  <div class="slide-content">
    <div class="card-display">
      <div class="card-container">
        <Card color={assetColorBindings[asset.name]} name={asset.name} amount={asset.amount} nominal={asset.nominal} />
      </div>
    </div>
  </div>
  {/each}
  <span class="control" slot="right-control">
    <ChevronRightIcon />
  </span>
</Carousel>

<style>
  .control :global(svg) {
    width: 100%;
    height: 100%;
    color: #fff;
    border: 2px solid #fff;
    border-radius: 32px;
  }

  * :global(.card-display) {
    margin: 90px;
    margin-top: 15px;
    margin-bottom: 0;
  }
</style>