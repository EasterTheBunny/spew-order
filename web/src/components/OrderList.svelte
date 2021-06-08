<script type="ts">
  import DataTable, { Head, Body, Row, Cell } from '@smui/data-table'
  import { OrderType } from '../constants'

  export let orders: IfcOrderResource[] = []

  const price: (order: IfcOrderResource) => string = (order) => {
    if (isLimit(order.order.type)) {
      return order.order.type.price
    }

    return "n/a"
  }

  function isLimit(item: IfcMarketOrder | IfcLimitOrder): item is IfcLimitOrder {
    return (item as IfcLimitOrder).name === OrderType.Limit
  }
</script>

<DataTable table$aria-label="People list" style="width: 100%;">
  <Head>
    <Row>
      <Cell>Action</Cell>
      <Cell>Status</Cell>
      <Cell>Type</Cell>
      <Cell>Price</Cell>
      <Cell>Amount</Cell>
    </Row>
  </Head>
  <Body>
    {#if !!orders}
    {#each orders as order}
    <Row>
      <Cell>{order.order.action.toLowerCase()}</Cell>
      <Cell>{order.status.toLowerCase()}</Cell>
      <Cell>{order.order.type.name.toLowerCase()}</Cell>
      <Cell>{price(order)}</Cell>
      <Cell>{order.order.type.quantity}</Cell>
    </Row>
    {/each}
    {/if}
  </Body>
</DataTable>