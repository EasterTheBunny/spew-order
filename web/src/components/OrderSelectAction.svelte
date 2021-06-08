<script type="ts">
  import SegmentedButton, { Segment } from "@smui/segmented-button";
  import { Label } from "@smui/common";

  import { ActionType } from "../constants"

  export let selected: ActionType = ActionType.Buy;

  const actionTypes = [ActionType.Buy, ActionType.Sell]
  const actionTypeLabels = {
    [ActionType.Buy]: "Buy",
    [ActionType.Sell]: "Sell",
  }
</script>

<SegmentedButton segments={actionTypes} let:segment singleSelect bind:selected={selected} on:click>
  <!-- Note: the `segment` property is required! -->
  <Segment {segment}>
    <Label>{actionTypeLabels[segment]}</Label>
  </Segment>
</SegmentedButton>