<script type="ts">
  import Fab, { Icon } from '@smui/fab/styled'
  import { Anchor } from '@smui/menu-surface/styled'
  import Menu from '@smui/menu/styled'
  import { getOidc } from "../oidc";
  import List, { Item, Text } from '@smui/list/styled'
  import { getLocalization } from '../i18n'

  let menu
  let anchor;
  let anchorClasses = {};

  const { oidc } = getOidc()
  const {t} = getLocalization()
</script>

<div
  class={Object.keys(anchorClasses).join(' ')}
  use:Anchor={{
    addClass: (className) => {
      if (!anchorClasses[className]) {
        anchorClasses[className] = true;
      }
    },
    removeClass: (className) => {
      if (anchorClasses[className]) {
        delete anchorClasses[className];
        anchorClasses = anchorClasses;
      }
    },
  }}
  bind:this={anchor}
>
<div class="flexy">
  <div class="margins">
    <Fab style="margin-left: 15px;" on:click={() => menu.setOpen(true)} mini>
      <Icon class="material-icons">person</Icon>
    </Fab>
    <Menu
      bind:this={menu}
      anchor={false}
      bind:anchorElement={anchor}
      anchorCorner="BOTTOM_LEFT"
    >
      <List>
        <Item on:SMUI:action={() => oidc.signOut()}>
          <Text>{$t('Signout')}</Text>
        </Item>
      </List>
    </Menu>
  </div>
</div>
</div>