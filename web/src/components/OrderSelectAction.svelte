<script type="ts">
  import SegmentedButton, { Segment } from "@smui/segmented-button";
  import { Label } from "@smui/common";

  import { orderActionBuy, orderActionSell } from "../constants"

  export let selected = orderActionBuy;

  const actionTypes = [orderActionBuy, orderActionSell]
  const actionTypeLabels = {
    [orderActionBuy]: "Buy",
    [orderActionSell]: "Sell",
  }
</script>

<SegmentedButton segments={actionTypes} let:segment singleSelect bind:selected={selected} on:click>
  <!-- Note: the `segment` property is required! -->
  <Segment {segment}>
    <Label>{actionTypeLabels[segment]}</Label>
  </Segment>
</SegmentedButton>