<script type="ts">
  import Select, { Option } from '@smui/select/styled';
  import { createEventDispatcher } from 'svelte';
  import { OrderType } from "../constants"
  import { getLocalization } from '../i18n'
  
  export let value: OrderType = OrderType.Market;

  const {t} = getLocalization()
  let orderTypes = [{
    value: OrderType.Market,
    label: t('MarketOrder'),
  }, {
    value: OrderType.Limit,
    label: t('LimitOrder'),
  }]
  
  const dispatch = createEventDispatcher();
  $: if(value) {
    dispatch('select', orderTypes.filter((o) => o.value === value)[0])
  }
</script>

<Select bind:value label={$t('OrderType')} style="width: 100%;">
  {#each orderTypes as tp}
    <Option value={tp.value}>{tp.label}</Option>
  {/each}
</Select>