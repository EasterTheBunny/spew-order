<svelte:window bind:innerHeight={height}/>

<section
  bind:this={element}
  class={classMap({
    [className]: true,
    'hero': true,
    ...internalClasses,
  })}
  style={Object.entries(internalStyles)
    .map(([name, value]) => `${name}: ${value};`)
    .concat([style])
    .join(' ')}
>
    <slot/>
</section>

<script type="ts">
  import {
    classMap,
  } from '@smui/common/internal';

  let className = '';
  export { className as class };
  export let style = '';
  let element: HTMLElement;

  let height = 800;

  let internalClasses: { [k: string]: boolean } = {};
  let internalStyles: { [k: string]: string } = {}; // { height: `${height}px` };

  $: ((h) => {
    // internalStyles["height"] = `${h}px`
  })(height)
</script>

<style>
  :global(.hero) {
    position: relative;
    display: flex;
    width: 100%;
    padding: 0;
    height: 100vh;
    margin-top: -56px;
  }

</style>