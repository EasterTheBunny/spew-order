<script lang="ts">
  import { onMount } from 'svelte';
  import TopAppBar, {
    Row,
    Section,
    Title,
    AutoAdjust,
    TopAppBarComponentDev,
  } from '@smui/top-app-bar';
  import IconButton, { Icon } from '@smui/icon-button';
  import Button, { Label } from '@smui/button';
  import LayoutGrid, { Cell } from '@smui/layout-grid';
  import Card, {
    Content,
    PrimaryAction,
    Actions,
    ActionButtons,
    ActionIcons,
  } from '@smui/card';
  import Paper, { Title, Content } from '@smui/paper';
  import { Svg } from '@smui/common/elements';
  import { mdiDiscord, mdiTwitter } from '@mdi/js';
  import Dialog, { Title, Content, Actions } from '@smui/dialog';

  import Hero from '$lib/hero-section';
  import FAQ from '$lib/sections';

  let topAppBar: TopAppBarComponentDev;
  let clicked = 0;
  let walletConnectText = "Connect";
  let open = false;
  
  const preMintPrice = "0.05 ETH";
  const mintPrice = "0.06 ETH";

  onMount(() => {})

  const getAddr = async () => {
    const accounts = await ethereum.request({ method: 'eth_requestAccounts' });
    const account = accounts[0];

    walletConnectText = shortenAddr(account)
  }

  const shortenAddr: (addr: string) => string = (addr) => {
    const firstFive = addr.slice(0, 5)
    const lastThree = addr.slice(-3)

    return `${firstFive}...${lastThree}`
  }

  const onConnectBtnClick = async () => {
    if (walletConnectText === "Connect" && typeof window.ethereum !== 'undefined') {
      if (ethereum.selectedAddress === null) {
        getAddr()
      } else {
        walletConnectText = shortenAddr(ethereum.selectedAddress)
      }
    }

  }
</script>

<TopAppBar bind:this={topAppBar} variant="short">
  <Row>
    <Section>
      <!--<IconButton class="material-icons">menu</IconButton>-->
      <Title>Cipher Mountain</Title>
      <Button href="#staking" variant="outlined" class="button-left">
        <Label>Stake</Label>
      </Button>
      <Button href="/whitepaper.pdf" target="_blank" variant="outlined" class="button-left">
        <Label>Whitepaper</Label>
      </Button>
      <!--
      <Button on:click={() => clicked++} variant="outlined" class="button-left">
        <Label>Roadmap</Label>
      </Button>
      -->
    </Section>
    <Section align="end" toolbar>
      <Button on:click={() => (open = true)} variant="outlined" class="button-right">
        <Label>Mint</Label>
      </Button>
      <Button on:click={onConnectBtnClick} variant="outlined" class="button-right">
        <Label>{walletConnectText}</Label>
      </Button>
    </Section>
  </Row>
</TopAppBar>

<AutoAdjust {topAppBar}>
  <Hero>
    <div class="hero-notice bottom">
      <h1 class="smui-button--color-secondary">Minting May 4th 2022</h1>
    </div>
    <!--
    <div class="hero-notice bottom">
      <h2>200 whitelist slots</h2>
    </div>
    -->
  </Hero>

  <LayoutGrid class="main-section">
    <Cell spanDevices={{ desktop: 4, tablet: 4, phone: 12 }}>
      <Card>
        <Content>
          <h2>7 Goat Levels</h2>
          <p>Stake, earn, level up, split, earn more. Expand your investment by earning rewards from staking.</p>
        </Content>
        <Actions fullBleed>
          <Button color="secondary" href="#goats">
            <Label>View the Goats</Label>
            <i class="material-icons" aria-hidden="true">arrow_forward</i>
          </Button>
        </Actions>
      </Card>
    </Cell>
    <Cell spanDevices={{ desktop: 4, tablet: 4, phone: 12 }}>
      <Card>
        <Content>
          <h2>$CMTN Tokens</h2>
          <p>Stake your Goats to earn $CMTN. Use $CMTN in our growing ecosystem of products.</p>
        </Content>
        <Actions fullBleed>
          <Button color="secondary" href="/whitepaper.pdf" target="_blank">
            <Label>Read the Whitepaper</Label>
            <i class="material-icons" aria-hidden="true">arrow_forward</i>
          </Button>
        </Actions>
      </Card>
    </Cell>
    <Cell spanDevices={{ desktop: 4, tablet: 4, phone: 12 }}>
      <Card>
        <Content>
          <h2>The DAO</h2>
          <p>Use $DMTN to level up and split Goats. Vote on community decisions with $DMTN.</p>
        </Content>
        <Actions fullBleed>
          <Button color="secondary" href="#dao">
            <Label>Explore the DAO</Label>
            <i class="material-icons" aria-hidden="true">arrow_forward</i>
          </Button>
        </Actions>
      </Card>
    </Cell>
  </LayoutGrid>

  <div class="limited-width">
  <LayoutGrid class="main-section">
    <Cell spanDevices={{ desktop: 8, tablet: 6, phone: 12 }}>
      <p>Owning a Goat gives you access to exclusive benefits in the entire ecosystem of Cipher Mountain.</p>
      <p>Purchasing a Goat at mint costs {mintPrice}. The pre-sale price is {preMintPrice}.</p>
      <p>Level based staking ensures that higher levels earn more rewards. $CMTN is the utility token and $DMTN is the DAO coin.</p>
      <p>The Goat minting contract is based on the ERC-721 standard on the Ethereum blockchain. All image assets are stored on IPFS.</p>
    </Cell>
    <Cell spanDevices={{ desktop: 4, tablet: 6, phone: 12 }}>
      <img src="/images/goat_nft_preview.gif" width="100%" />
    </Cell>
  </LayoutGrid>
  </div>

  <div class="main-section light-section" id="goats">
    <div class="limited-width">
      <LayoutGrid>
        <Cell spanDevices={{ desktop: 6, tablet: 6, phone: 12 }} style="text-align: center;">
          <img src="/images/beach_no_bg.png" width="100%" />
        </Cell>
        <Cell spanDevices={{ desktop: 6, tablet: 6, phone: 12 }}>
          <Paper color="primary" elevation={12}>
            <Title>Many outfits, much fun!</Title>
            <Content>
              Goats levels 1-6 are indicated by their shirt, if they have one.
              Level 7 Goats are stamped with a bunny signature.
              <br/><br/>
              Numerous features make each Goat unique with many more to come!
            </Content>
          </Paper>
        </Cell>
      </LayoutGrid>
    </div>
  </div>

  <LayoutGrid class="main-section" id="dao">
    <Cell spanDevices={{ desktop: 6, tablet: 6, phone: 12 }}>
      <Card>
        <Content>
          <h2>DAO Assets</h2>
          <p>The DAO will purchase blue chip assets to store in the DAO Vault. The DAO will make decisions on when and what to sell. Decision making will be determined using $DMTN tokens. All proceeds from a sale will be added to the DAO Vault.</p>
        </Content>
        <Actions fullBleed>
          <Button color="secondary" href="/whitepaper.pdf" target="_blank">
            <Label>Read the Whitepaper</Label>
            <i class="material-icons" aria-hidden="true">arrow_forward</i>
          </Button>
        </Actions>
      </Card>
    </Cell>
    <Cell spanDevices={{ desktop: 6, tablet: 6, phone: 12 }}>
      <Card>
        <Content>
          <h2>DAO Development</h2>
          <p>The Cipher Mountain ecosystem of products will be determined by the DAO. Committed additions, improvements, or removals will be voted on. Cost of labor will be paid from the DAO Vault and DAO members will be first in line for assigned work items.</p>
        </Content>
        <Actions fullBleed>
          <Button color="secondary" href="/whitepaper.pdf" target="_blank">
            <Label>Read the Whitepaper</Label>
            <i class="material-icons" aria-hidden="true">arrow_forward</i>
          </Button>
        </Actions>
      </Card>
    </Cell>
  </LayoutGrid>

  <div class="main-section light-section" id="staking">
    <div class="limited-width">
      <LayoutGrid class="main-section">
        <Cell span={12}>
          <h1>Staking Rewards</h1>
          <p>
            Our staking program allows you to grow your portfolio over time! Stake your Goats. Earn coin. Level them up to the MAX and split! Once a Goat is split, the resulting Goats can be further staked and leveled.
          </p>
          <img src="/images/staking_flow_chart.svg" width="100%" />
        </Cell>
      </LayoutGrid>
    </div>
  </div>

  <LayoutGrid class="main-section">
    <Cell span={12}>
      <img src="/images/roadmap.svg" />
    </Cell>
  </LayoutGrid>

  <div class="main-section">
    <div class="limited-width">
      <LayoutGrid class="main-section">
        <Cell span={12}>

          <Paper color="primary" elevation={12}>
            <Title>The Cipher Mountain Story</Title>
            <Content>
              <p>
              A collection of tools built to serve our community. Born from the idea of an exchange, the fast moving pace of cryptocurrency development pushed us toward NFTs and community building. As an initial first step toward our goal, we have developed a Discord/Twitter game bot, <i>Chupagoat</i>, to assist community managers build and maintain a following. While our latest outcome isn't what we planned, we still stay true to our original goal.
              </p>
              <p>
              Not only do we want to build for our community. We want to BE our community. As a DAO, we aim to grow together.
              </p>
            </Content>
          </Paper>

        </Cell>
      </LayoutGrid>
    </div>
  </div>

  <div class="main-section light-section">
    <div class="limited-width">
      <LayoutGrid>
        <Cell span={12}>
          <h1>Frequently Asked Questions</h1>
          <FAQ mintPrice={mintPrice} preMintPrice={preMintPrice} />
        </Cell>
      </LayoutGrid>
    </div>
  </div>

  <LayoutGrid class="main-section">
    <Cell span={12}>
      <div style="text-align: center">
        <IconButton mini href="https://discord.gg/6gfNxC9Hj5" target="_blank" ripple={false}>
          <Icon component={Svg} viewBox="0 0 24 24">
            <path fill="currentColor" d={mdiDiscord} />
          </Icon>
        </IconButton>
        <IconButton mini href="https://twitter.com/CipherMountain" target="_blank" ripple={false}>
          <Icon component={Svg} viewBox="0 0 24 24">
            <path fill="currentColor" d={mdiTwitter} />
          </Icon>
        </IconButton>
      </div>
      <p style="text-align: center;font-size: 1.0rem;">
        <small>(c) 2022 Cipher Mountain LLC</small>
      </p>
    </Cell>
  </LayoutGrid>
</AutoAdjust>

<Dialog
  bind:open
  aria-labelledby="simple-title"
  aria-describedby="simple-content"
>
  <!-- Title cannot contain leading whitespace due to mdc-typography-baseline-top() -->
  <Title id="simple-title">Minting Notice</Title>
  <Content id="simple-content">Pre-sale mint begins on May 3rd 2022. Public sale begins May 4th 2022.</Content>
  <Actions>
    <Button on:click={() => (open = false)}>
      <Label>Ok</Label>
    </Button>
  </Actions>
</Dialog>

<style>
  /* Hide everything above this component. */
  :global(app),
  :global(body),
  :global(html) {
    display: block !important;
    height: auto !important;
    width: auto !important;
    position: static !important;
  }

  .hero-notice {
    position: absolute;
    left: 0;
    background-color: rgba(0, 0, 0, 0.7);
    width: 100%;
    padding: 15px 35px;
  }

  .hero-notice.top {
    top: 0;
  }

  .hero-notice.bottom {
    bottom: 0;
  }

  .hero-notice > h1 {
    opacity: 1.0;
  }

  .hero-notice h1 {
    font-size: 3em;
    font-family: 'Magistral-Medium', Roboto, sans-serif;
    font-style: italic;
    text-align: center;
  }

  :global(.main-section) {
    margin-top: 50px;
  }

  .limited-width {
    margin: 0 auto;
  }

  @media screen and (min-width: 900px) {
    .limited-width {
      width: 900px;
    }
  }

  :global(.custom-white) {
    color: #ffffff;
  }
</style>