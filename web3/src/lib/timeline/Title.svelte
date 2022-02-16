<li class="timeline-title">
  <div class="timeline-circle timeline-title-circle">
    <slot name="icon" />
  </div>
  <slot />
</li>

<style>
  .timeline-title {
    position: relative;
    display: inline-block;
    cursor: crosshair;
    margin: -.15em 0 15px 28px
  }

  .timeline-title:not(:first-child) {
    margin-top: 28px;
  }

  .timeline-title-circle {
    left: -36px;
    top: .15em;
    width: 16px;
    height: 16px;
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