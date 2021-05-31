<script type="ts">
  import Textfield from '@smui/textfield';
  import HelperText from '@smui/textfield/helper-text/index';

  let amount = ""
  let price = ""
  let total = ""
</script>

<div class="form-section">
  <Textfield
    class="shaped-outlined"
    variant="outlined"
    style="width: 100%;"
    helperLine$style="width: 100%;"
    bind:value={amount}
    label="Amount"
  >
    <HelperText slot="helper">0.0000000</HelperText>
  </Textfield>
</div>

<div class="form-section">
  <Textfield
    class="shaped-outlined"
    variant="outlined"
    style="width: 100%;"
    helperLine$style="width: 100%;"
    bind:value={price}
    label="Price"
  >
    <HelperText slot="helper">0.0000000</HelperText>
  </Textfield>
</div>

<div class="form-section">
  <Textfield
    class="shaped-outlined"
    variant="outlined"
    style="width: 100%;"
    helperLine$style="width: 100%;"
    bind:value={total}
    label="Total"
  >
    <HelperText slot="helper">0.0000000</HelperText>
  </Textfield>
</div>