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
    display: block;
    width: 100%;
    margin: -56px 0px;
    padding: 0;
    height: 100vh;
    background-image: url('/images/base_logo.png');
    background-size: cover;
    background-position: center;
  }

</style>