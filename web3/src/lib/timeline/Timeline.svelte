<ul class="timeline">
  <slot />
</ul>

<style>
  .timeline {
    padding: 0;
    position: relative;
    list-style: none;
    -webkit-font-smoothing: antialiased;
    margin: 10px 20px;
  }

  .timeline:after {
    position: absolute;
    content: '';
    left: 0;
    top: 0;
    width: 1px;
    height: 100%;
    background-color: var(--timelineTheme);
  }
</style>