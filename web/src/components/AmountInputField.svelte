<script type="ts">
  import Textfield from '@smui/textfield';
  import HelperText from '@smui/textfield/helper-text/index';
  import { createEventDispatcher } from 'svelte';

  export let label = "Label"
  export let value = "0.00000"
  export let subtext = ""
  export let symbol = "BTC"

  const validMessageName = "valid"
  const dispatch = createEventDispatcher();
  const validate = (val: string): boolean => {
    let v = parseFloat(val)
    if (v <= 0) {
      dispatch(validMessageName, false);
      return false
    }

    dispatch(validMessageName, true);
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
  on:keyup
>
  <span slot="suffix">{symbol}</span>
  <HelperText slot="helper">{subtext}</HelperText>
</Textfield>