<script type="ts">
  import Select, { Option } from '@smui/select/styled';
  import { createEventDispatcher } from 'svelte';

  import { OrderType } from "../constants"
  
  export let value: OrderType = OrderType.Market;

  let orderTypes = [{
    value: OrderType.Market,
    label: 'Market Order',
  }, {
    value: OrderType.Limit,
    label: 'Limit Order'
  }]
  
  const dispatch = createEventDispatcher();
  $: if(value) {
    dispatch('select', orderTypes.filter((o) => o.value === value)[0])
  }
</script>

<Select bind:value label="Order Type" style="width: 100%;">
  {#each orderTypes as tp}
    <Option value={tp.value}>{tp.label}</Option>
  {/each}
</Select>