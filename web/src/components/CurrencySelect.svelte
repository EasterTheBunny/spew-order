<script type="ts">
  import Select, { Option } from '@smui/select';
  import { Currency } from '../constants';

  export let selected: Currency

  const currencies = [{
    value: Currency.Ethereum,
    name: 'Ethereum',
  }, {
    value: Currency.Bitcoin,
    name: 'Bitcoin',
  }]

</script>

<Select bind:value={selected} label="Select Currency">
  {#each currencies as c}
    <Option value={c.value}>{c.name}</Option>
  {/each}
</Select>