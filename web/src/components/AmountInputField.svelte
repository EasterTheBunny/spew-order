<script type="ts">
  import Textfield from '@smui/textfield';
  import HelperText from '@smui/textfield/helper-text/index';

  export let label = "Label"
  export let value = "0.00000"
  export let subtext = ""
  export let symbol = "BTC"

  const validate = (val: string): boolean => {
    let v = parseFloat(val)
    if (v <= 0) {
      return false
    }

    return true
  }

  $: invalid = !validate(value)
</script>

<Textfield
  class="shaped-outlined"
  variant="outlined"
  style="width: 100%;"
  helperLine$style="width: 100%;"
  bind:value={value}
  {invalid}
  label="{label}"
>
  <span slot="suffix">{symbol}</span>
  <HelperText slot="helper">{subtext}</HelperText>
</Textfield>