<script type="ts">
  import Tab, { Label } from '@smui/tab/styled'
  import TabBar from '@smui/tab-bar/styled'
  import { ActionType } from "../constants"
  import { getLocalization } from '../i18n'

  export let active: ActionType = ActionType.Buy;

  const {t} = getLocalization()
  const actionTypes = [ActionType.Buy, ActionType.Sell]
  const actionTypeLabels = {
    [ActionType.Buy]: t("Buy"),
    [ActionType.Sell]: t("Sell"),
  }
</script>

<TabBar tabs={actionTypes} let:tab bind:active>
  <!-- Note: the `tab` property is required! -->
  <Tab {tab}>
    <Label>{actionTypeLabels[tab]}</Label>
  </Tab>
</TabBar>

<style lang="scss">
  :global(.mdc-tab--active > .mdc-tab__content > .mdc-tab__text-label) {
    color: #fff;
  }
</style>