<script lang="ts">
  import Carousel from '@beyonk/svelte-carousel';
</script>

<Carousel loop={true} autoplay={3000} perPage={5} duration={3000} dots={false} controls={false} draggable={false}>
  <div class="slide-content"><img src="/images/0020.png" width="100%" /></div>
  <div class="slide-content"><img src="/images/0029.png" width="100%" /></div>
  <div class="slide-content"><img src="/images/0039.png" width="100%" /></div>
  <div class="slide-content"><img src="/images/0068.png" width="100%" /></div>
  <div class="slide-content"><img src="/images/0129.png" width="100%" /></div>
  <div class="slide-content"><img src="/images/0153.png" width="100%" /></div>
  <div class="slide-content"><img src="/images/0181.png" width="100%" /></div>
  <div class="slide-content"><img src="/images/0257.png" width="100%" /></div>
  <div class="slide-content"><img src="/images/0259.png" width="100%" /></div>
  <div class="slide-content"><img src="/images/1008.png" width="100%" /></div>
</Carousel>