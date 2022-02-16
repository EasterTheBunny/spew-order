<script lang="ts">
  import Button, { Label } from '@smui/button';
  import Textfield from '@smui/textfield';
  import HelperText from '@smui/textfield/helper-text';
 
  let clicked = 0;
  let valueA = "";
</script>

<div class="d-flex flex-row flex-column-fluid">
  <div class="aside py-9">
    <div class="aside-logo flex-column-auto px-9 mb-9">
      <a href="ciphermtn.com">
        <img alt="Logo" src="https://images.squarespace-cdn.com/content/v1/612695928e6a194fbf935d9e/8fb88a70-4d1c-4607-9e87-90d57cfb317c/header_logo.png?format=1500w" style="width: 100%;" />
      </a>
    </div>

    <div class="aside-menu flex-column-fluid ps-5 pe-3 mb-9">
      <div>Left</div>
      <div>
        <ul>
          <li>takes up left static width @medium and higher</li>
          <li>is drawer toggled by nav burger @small and lower</li>
        </ul>
      </div>
    </div>

    <div class="aside-footer flex-column-auto px-9">footer</div>
  </div>

  <div class="wrapper d-flex flex-column flex-row-fluid">
    <div class="header">
      <div class="container d-flex flex-stack flex-wrap gap-2">

        <div class="page-title d-flex flex-column align-items-start justify-content-center flex-wrap me-lg-2 pb-5 pb-lg-0">
          <h1 class="d-flex flex-column fs-1">
            Project Name
            <small class="text-muted fs-6">Other info</small>
          </h1>
        </div>

        <div class="d-flex align-items-center flex-shrink-0">
          <div>
            <Textfield
              class="shaped-outlined"
              variant="outlined"
              bind:value={valueA}
              label="Label"
            ></Textfield>
          </div>
          
          <Button on:click={() => clicked++} variant="raised">
            <Label>Raised</Label>
          </Button>
        </div>

      </div>
    </div>

    <div class="content d-flex flex-column flex-column-fluid">
      <div class="container-xxl">
        <h1>Main section with project info</h1>
        <p>middle dynamic width</p>
        <a href="/p/test">site link</a>
      </div>
    </div>

    <div class="footer py-4 d-flex flex-lg-column">
      <div class="container d-flex flex-column flex-md-row flex-stack">
        footer
      </div>
    </div>
  </div>

  <div class="sidebar">
    <div class="d-flex flex-column sidebar-body px-5 py-10">
      Right
      - takes up right static width @large
      - is drawer toggled by button in middle for smaller than @large
    </div>
  </div>
</div>
