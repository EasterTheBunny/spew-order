<script type="ts">
  import type { Writable } from "svelte/store"
  import { onDestroy, onMount } from "svelte"
  import DataTable, { Head, Body, Row, Cell } from '@smui/data-table/styled'
  import { Icon } from '@smui/common/styled'
  import Tooltip, { Wrapper } from '@smui/tooltip/styled'
  import OrderStatusIcon from './OrderStatusIcon.svelte'
  import { OrderStatus, OrderType } from '../constants'
  import { getDataCtx } from "../exchange";
  import { getLocalization } from '../i18n'

  let orderList: IfcOrderResource[] = []
  let unsubscribe_orders = () => {}

  const {
    orders,
  }: {
    orders: Writable<IfcOrderResource[] | IfcOrderResource>
  } = getDataCtx()
  const {t} = getLocalization()

  const price: (order: IfcOrderResource) => string = (order) => {
    if (isLimit(order.order.type)) {
      return order.order.type.price
    }

    return "n/a"
  }

  function isLimit(item: IfcMarketOrder | IfcLimitOrder): item is IfcLimitOrder {
    return (item as IfcLimitOrder).name === OrderType.Limit
  }

  const cancelOrder: (order: IfcOrderResource) => void = (order) => {
    orders.update((v: IfcOrderResource): IfcOrderResource => {
      if (v.guid === order.guid) {
        v.status = OrderStatus.Cancelled
      }
      return v
    })
  }

  onMount(() => {
    unsubscribe_orders = orders.subscribe((values: IfcOrderResource[]) => {
      orderList = values.filter((v) => v.status === OrderStatus.Open || v.status === OrderStatus.Partial)
    })
  })

  onDestroy(() => {
    unsubscribe_orders()
  })
</script>

<DataTable table$aria-label="People list" style="width: 100%;">
  <Head>
    <Row>
      <Cell>{$t('Action')}</Cell>
      <Cell>{$t('Type')}</Cell>
      <Cell>{$t('Price')}</Cell>
      <Cell>{$t('Amount')}</Cell>
      <Cell>{$t('Status')}</Cell>
      <Cell></Cell>
    </Row>
  </Head>
  <Body>
    {#each orderList as order}
    <Row>
      <Cell>{order.order.action.toLowerCase()}</Cell>
      <Cell>{order.order.type.name.toLowerCase()}</Cell>
      <Cell>{price(order)}</Cell>
      <Cell>{order.order.type.quantity}</Cell>
      <Cell>
        <OrderStatusIcon bind:value={order} />
      </Cell>
      <Cell>
        {#if order.status !== OrderStatus.Cancelled}
        <div style="color: red; cursor: pointer;" on:click={() => cancelOrder(order)}>
          <Wrapper>
            <Icon class="material-icons">cancel</Icon>
            <Tooltip>cancel</Tooltip>
          </Wrapper>
        </div>
        {/if}
      </Cell>
    </Row>
    {/each}
  </Body>
</DataTable>