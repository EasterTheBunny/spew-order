<script type="ts">
  import { Icon } from '@smui/common'
  import Tooltip, { Wrapper } from '@smui/tooltip'
  import { OrderStatus } from '../constants'

  export let value: IfcOrderResource
  
  const setIcon: (o: IfcOrderResource) => string = (o) => {
    if (o === null) {
      return "remove_circle"
    }

    switch (o.status) {
      case OrderStatus.Cancelled:
        return "not_interested"
      case OrderStatus.Open:
        return "done"
      case OrderStatus.Partial:
        return "done_all"
      case OrderStatus.Filled:
        return "done_all"
      default:
        return "not_interested"
    }
  }
  
  const setColor: (o: IfcOrderResource) => string = (o) => {
    if (o === null) {
      return "remove_circle"
    }

    switch (o.status) {
      case OrderStatus.Cancelled:
        return "red"
      case OrderStatus.Open:
        return "gray"
      case OrderStatus.Partial:
        return "gray"
      case OrderStatus.Filled:
        return "blue"
      default:
        return "remove_circle"
    }
  }

  $: icon = setIcon(value)
  $: color = setColor(value)
</script>

<div style="color: {color}; cursor: pointer;">
  <Wrapper>
    <Icon class="material-icons">{icon}</Icon>
    <Tooltip>{value.status.toLowerCase()}</Tooltip>
  </Wrapper>
</div>