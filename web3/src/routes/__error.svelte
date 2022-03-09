<script context="module">
  /** @type {import('@sveltejs/kit').ErrorLoad} */
  export function load({ error, status }) {
    return {
      props: { status, error }
    };
  }
</script>

<script type="ts">
  import Hero from '$lib/hero-section';

  export let status: number = 404;
  export let error: string = "page not found";
</script>

<Hero style="align-items: center; justify-content: center;">
  <div>
    <p style="font-size: 8rem;">{status}</p>
    <p>How did you get here?</p>
  </div>
</Hero>
