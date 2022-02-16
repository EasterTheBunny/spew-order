<script lang="ts">
  import Accordion, { Panel, Header, Content } from '@smui-extra/accordion';
  import IconButton, { Icon } from '@smui/icon-button';

  let panel1Open = false;
  let panel2Open = false;
  let panel3Open = false;
  let panel4Open = false;
  let panel5Open = false;
  let panel6Open = false;
  let panel7Open = false;
</script>

<Accordion class="demo-small-titles">
  <Panel
    square
    variant="outlined"
    color="secondary"
    extend
    bind:open={panel1Open}
  >
    <Header>
      How many Goats are available?
      <span slot="description">2000</span>
      <IconButton slot="icon" toggle pressed={panel1Open}>
        <Icon class="material-icons" on>unfold_less</Icon>
        <Icon class="material-icons">unfold_more</Icon>
      </IconButton>
    </Header>
    <Content>At mint there are 2000 Goats available. More will be created through the leveling and splitting process over time.</Content>
  </Panel>
  <Panel
    square
    variant="outlined"
    color="secondary"
    extend
    bind:open={panel2Open}
  >
    <Header>
      How much to mint a Goat?
      <span slot="description">0.05 ETH</span>
      <IconButton slot="icon" toggle pressed={panel2Open}>
        <Icon class="material-icons" on>unfold_less</Icon>
        <Icon class="material-icons">unfold_more</Icon>
      </IconButton>
    </Header>
    <Content>The public sale price to mint is 0.05 ETH. A special pre-sale price will be 0.04 ETH available only to whitelisted addresses.</Content>
  </Panel>
  <Panel
    square
    variant="outlined"
    color="secondary"
    extend
    bind:open={panel3Open}
  >
    <Header>
      How about that whitelist?
      <span slot="description">200 slots</span>
      <IconButton slot="icon" toggle pressed={panel3Open}>
        <Icon class="material-icons" on>unfold_less</Icon>
        <Icon class="material-icons">unfold_more</Icon>
      </IconButton>
    </Header>
    <Content>There are only 200 whitelist slots and will be made available via Twitter and Discord. The whitelist allows access to the pre-sale with a special pre-sale price.</Content>
  </Panel>
  <Panel
    square
    variant="outlined"
    color="secondary"
    extend
    bind:open={panel4Open}
  >
    <Header>
      When is the pre-sale and public sale?
      <span slot="description">May 3rd and 4th 2022</span>
      <IconButton slot="icon" toggle pressed={panel4Open}>
        <Icon class="material-icons" on>unfold_less</Icon>
        <Icon class="material-icons">unfold_more</Icon>
      </IconButton>
    </Header>
    <Content>The pre-sale date is May 3rd 2022 and pre-sale availability will be determined by a whitelist with a special mint price of 0.04 ETH. The public sale will begin on May 4th 2022 and will be available for anyone to mint.</Content>
  </Panel>
  <Panel
    square
    variant="outlined"
    color="secondary"
    extend
    bind:open={panel5Open}
  >
    <Header>
      What is the staking contract?
      <IconButton slot="icon" toggle pressed={panel5Open}>
        <Icon class="material-icons" on>unfold_less</Icon>
        <Icon class="material-icons">unfold_more</Icon>
      </IconButton>
    </Header>
    <Content>The staking contract rewards holders of Goat NFTs. Earn $CMTN to buy things in the Cipher Mountain ecosystem. Earn $DMTN to get active in the DAO. Use $DMTN to level up or split your staked Goats, increasing your yield.</Content>
  </Panel>
  <Panel
    square
    variant="outlined"
    color="secondary"
    extend
    bind:open={panel6Open}
  >
    <Header>
      What is in the Ecosystem?
      <IconButton slot="icon" toggle pressed={panel6Open}>
        <Icon class="material-icons" on>unfold_less</Icon>
        <Icon class="material-icons">unfold_more</Icon>
      </IconButton>
    </Header>
    <Content>The Cipher Mountain ecosystem is constantly expanding. As a starting point, we have developed a Discord bot called Chupagoat that provides a fun game for community members to play in Discord and Twitter as well as tools for community managers to run promotions and outreach.</Content>
  </Panel>
  <Panel
    square
    variant="outlined"
    color="secondary"
    extend
    bind:open={panel7Open}
  >
    <Header>
      What is the DAO?
      <IconButton slot="icon" toggle pressed={panel7Open}>
        <Icon class="material-icons" on>unfold_less</Icon>
        <Icon class="material-icons">unfold_more</Icon>
      </IconButton>
    </Header>
    <Content>The DAO is a way to provide decision making capabilities to the community and collectively reap the rewards. Assets will be collected and held by the DAO and all proceeds from asset sales will go back to the DAO.</Content>
  </Panel>
</Accordion>

<style>
  *
    :global(.demo-small-titles
      .smui-accordion__header__title--with-description) {
    flex-basis: 20% !important;
    max-width: 200px !important;
  }
</style>
