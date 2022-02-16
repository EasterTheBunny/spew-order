<script type="ts">
  import Card, {
    Content,
    PrimaryAction,
    Media,
    MediaContent,
    Actions,
    ActionButtons,
    ActionIcons,
  } from '@smui/card/styled';
  import Button, { Label } from '@smui/button/styled';
  import IconButton, { Icon } from '@smui/icon-button/styled';
  import LayoutGrid, { Cell } from '@smui/layout-grid';

  let clicked = 0;

  export let name = "Bitcoin";
  export let color = "#000000";
  export let amount = "10.2345892";
  export let nominal = "";
</script>

<Card>
  <PrimaryAction on:click={() => clicked++}>
    <Content class="mdc-typography--body1 card-media" style={"background-color: "+color+";"}>
      <LayoutGrid>
        <Cell span={6}>
          <h1 class="mdc-typography--headline6" style="margin: 0;">
            {name}
          </h1>
        </Cell>
        <Cell align="middle" span={6}>
          <h3 class="mdc-typography--subtitle2" style="margin: 0; text-align: right;">
            {amount}
          </h3>
          <h4 style="text-align: right;">{nominal}</h4>
        </Cell>
      </LayoutGrid>
    </Content>
  </PrimaryAction>

  <Actions>
    <ActionButtons>
      <Button on:click={() => clicked++}>
        <Label>Action</Label>
      </Button>
      <Button on:click={() => clicked++}>
        <Label>Another</Label>
      </Button>
    </ActionButtons>
    <ActionIcons>
      <IconButton
        on:click={() => clicked++}
        toggle
        aria-label="Add to favorites"
        title="Add to favorites"
      >
        <Icon class="material-icons" on>favorite</Icon>
        <Icon class="material-icons">favorite_border</Icon>
      </IconButton>
      <IconButton
        class="material-icons"
        on:click={() => clicked++}
        title="Share">share</IconButton
      >
      <IconButton
        class="material-icons"
        on:click={() => clicked++}
        title="More options">more_vert</IconButton
      >
    </ActionIcons>
  </Actions>
</Card>
