<script type="ts">
  import Select, { Option } from '@smui/select/styled';
  import { Currency } from '../constants';

  export let selected: Currency

  const currencies = [{
    value: Currency.Ethereum,
    name: 'Ethereum',
  }, {
    value: Currency.Bitcoin,
    name: 'Bitcoin',
  }, {
    value: Currency.BitcoinCash,
    name: 'Bitcoin Cash',
  }, {
    value: Currency.Dogecoin,
    name: 'Dogecoin',
  }]

</script>

<Select bind:value={selected} label="Select Currency">
  {#each currencies as c}
    <Option value={c.value}>{c.name}</Option>
  {/each}
</Select>