<script type="ts">
  import type { Readable } from "svelte/store"
  import dayjs from 'dayjs'
  import DataTable, { Head, Body, Row, Cell } from '@smui/data-table/styled'
  import { getDataCtx } from "../exchange";
  import { TransactionType } from '../constants';
  import { getLocalization } from '../i18n'

  const {
    transactions,
  }: {
    transactions: Readable<IfcTransactionResource[]>
  } = getDataCtx()
  const {t} = getLocalization()

  const format: (t: string) => string = (t) => {
    return dayjs(t, 'YYYY-MM-DDTHH:mm:ssZ').format('MM/DD/YYYY HH:mm ZZ')
  }

  $: filtered = $transactions.filter((t: IfcTransactionResource) => t.type === TransactionType.Deposit || t.type === TransactionType.Transfer )

</script>

<DataTable table$aria-label="People list" style="width: 100%;">
  <Head>
    <Row>
      <Cell>{$t('Date')}</Cell>
      <Cell>{$t('Currency')}</Cell>
      <Cell>{$t('Quantity')}</Cell>
    </Row>
  </Head>
  <Body>
    {#each filtered as t}
    <Row>
      <Cell>{format(t.timestamp)}</Cell>
      <Cell>{t.symbol}</Cell>
      <Cell>{t.quantity}</Cell>
    </Row>
    {/each}
  </Body>
</DataTable>