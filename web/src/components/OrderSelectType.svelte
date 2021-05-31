<script type="ts">
  import Select, { Option } from '@smui/select';
  import { createEventDispatcher } from 'svelte';

  import { orderTypeMarket, orderTypeLimit } from "../constants"
  
  export let value = orderTypeMarket;

  let orderTypes = [{
    value: orderTypeMarket,
    label: 'Market Order',
  }, {
    value: orderTypeLimit,
    label: 'Limit Order'
  }]
  
  const dispatch = createEventDispatcher();
  $: if(value) {
    dispatch('select', orderTypes.filter((o) => o.value === value)[0])
  }
</script>

<Select bind:value label="Order Type">
  {#each orderTypes as tp}
    <Option value={tp.value}>{tp.label}</Option>
  {/each}
</Select>