<li class="timeline-item">
  <div class="timeline-circle">
    <slot name="icon" />
  </div>
  <slot />
</li>

<style>
  .timeline-item {
    position: relative;
    margin: 1.5em 0 0 28px;
    padding-bottom: 1.5em;
    border-bottom: 1px dotted var(--timelineTheme);
  }

  .timeline-item:last-child {
    border-bottom: none;
  }

  .timeline-circle {
    position: absolute;
    top: .28em;
    left: -34px;
    width: 10px;
    height: 10px;
    border-radius: 50%;
    border: 1px solid var(--timelineTheme);
    background-color: var(--timelineTheme);
    z-index: 1;
    box-sizing: content-box;
  }

  .timeline-circle.hollow {
    background-color: var(--timelineBg);
  }
  
  .timeline-others {
    width: 40px;
    height: auto;
    top: .2em;
    left: -48px;
    line-height: 1;
    padding: 2px 0;
    text-align: center;
    border: none;
    background-color: var(--timelineBg);
  }
</style>
