<script type="ts">
  import Textfield from '@smui/textfield/styled'
  import HelperText from '@smui/textfield/helper-text/styled'
  import { createEventDispatcher } from 'svelte'
  import type { Currency } from '../constants'

  export let value = ""
  export let currency: Currency

  const validMessageName = "valid"
  const dispatch = createEventDispatcher();
  const validate = (val: string): boolean => {
    if (val.length == 0) {
      dispatch(validMessageName, false);
      return false
    }

    dispatch(validMessageName, true);
    return true
  }

  $: label = currency + " Send Address"
  $: invalid = !validate(value)
  
  let subtext = ""
</script>

<Textfield bind:value bind:label on:keyup {invalid}>
  <HelperText slot="helper">{subtext}</HelperText>
</Textfield>