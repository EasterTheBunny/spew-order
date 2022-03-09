<script type="ts">
  import Hero from '$lib/hero-section';
  import BotButton from '$lib/components/discord';
  import LayoutGrid, { Cell } from '@smui/layout-grid';
  import Paper, { Title, Content } from '@smui/paper';
</script>

<Hero style="align-items: center; justify-content: center;">
  <div style="width: 70%;">
    <h1 style="font-size: 4rem;">Bring humor to Discord</h1>
    <h1 style="font-size: 1.5rem;font-style: italic;">The bot that hits your funny bone</h1>
    <BotButton style="margin-top: 20px;" color="secondary" variant="outlined" />
  </div>
</Hero>

<div style="padding-top: 25px; padding-bottom: 25px;" class="light-section">
  <div style="margin: 0 auto; width: 75%;">
    <h1>Alternate Reality Game</h1>
    <h2><i>Available in Discord</i></h2>
    <strong>You are launched into a deadly multiverse. Every encounter could be your last! Maybe you'll stub your toe or be abducted by aliens. Or you might find a truly delicious sandwich! Your fate will decide.</strong>
    <LayoutGrid style="margin-top: 30px;">
      <Cell spanDevices={{ desktop: 6, tablet: 4, phone: 6 }}>
        <img src="/images/spawnin_command.png" alt="Spawnin Slash Command" style="width: 100%;" />
      </Cell>
      <Cell spanDevices={{ desktop: 6, tablet: 4, phone: 6 }}>
        <div  style="display: flex; position: relative; align-items: center; justify-content: center; height: 100%;">
          <i>Spawn in as a random character. Some characters have higher skills than others!</i>
        </div>
      </Cell>
      <Cell spanDevices={{ desktop: 6, tablet: 4, phone: 6 }}>
        <div style="display: flex; position: relative; align-items: center; justify-content: center; height: 100%;">
          <div>
            <i>Walk around in the metaverse. Sometimes you find stuff. Other times you have very dangerous encounters!</i>
            <br/><br/>
            <i>WARNING!<br/>You have a fairly high chance of dying. If you do, no worries! Just spawn in again and keep walking.</i>
          </div>
        </div>
      </Cell>
      <Cell spanDevices={{ desktop: 6, tablet: 4, phone: 6 }}>
        <img src="/images/walk_command.png" alt="Walk Slash Command" style="width: 100%" />
      </Cell>

      <Cell spanDevices={{ desktop: 6, tablet: 4, phone: 6 }}>
        <img src="/images/backpack_command.png" alt="Backpack Slash Command" style="width: 100%" />
      </Cell>
      <Cell spanDevices={{ desktop: 6, tablet: 4, phone: 6 }}>
        <div style="display: flex; position: relative; align-items: center; justify-content: center; height: 100%;">
          <div>
            <i>Collect items in your backpack. Build XP and earn cash to access purchasable items.</i>
            <br/>
            <br/>
            <i>Some items help you through the game. Others get you to the GRAND PRIZE!</i>
          </div>
        </div>
      </Cell>
    </LayoutGrid>

  </div>
</div>

<div style="padding-top: 25px; padding-bottom: 25px;">
  <div style="margin: 0 auto; width: 75%;">
    <h1>Exclusive Features</h1>
    <h2><i>Available in Twitter as #ChupagoatARG</i></h2>

    <strong>Allow your Twitter followers to continue playing the Chupagoat ARG right in Twitter with hashtags, replies, and retweets! Keep the fun going while driving engagement and outreach.</strong>

    <LayoutGrid style="margin-top: 30px;">
      <Cell spanDevices={{ desktop: 6, tablet: 4, phone: 6 }}>
        <div style="display: flex; position: relative; align-items: center; justify-content: center; height: 100%;">
          <div>
            <i>A custom Twitter hashtag runs the walk command. Incentivize your followers to use your hashtag by connecting it to gameplay!</i>
          </div>
        </div>
      </Cell>

      <Cell spanDevices={{ desktop: 6, tablet: 4, phone: 6 }}>
        <img src="/images/twitter_walk.png" alt="Twitter Hashtag Walk" style="width: 100%" />
      </Cell>

      <Cell spanDevices={{ desktop: 6, tablet: 4, phone: 6 }}>
        <img src="/images/twitter_chupagoat_response.png" alt="Twitter Chupagoat Response" style="width: 100%" />
      </Cell>

      <Cell spanDevices={{ desktop: 6, tablet: 4, phone: 6 }}>
        <div style="display: flex; position: relative; align-items: center; justify-content: center; height: 100%;">
          <div>
            <i>Chupagoat responds to every use of your hashtag with the same output in Discord.</i>
          </div>
        </div>
      </Cell>

      <Cell span={12}>
        <Paper color="primary" elevation={12} style="margin-top: 20px;">
          <Title>Twitter Loot Boxes</Title>
          <Content>
            <p>
            Add something a little extra to tweets with your hashtag! Loot boxes allow you to automate Twitter giveaways.
            </p>
            <p>
              <i>Configure a wide range of items to giveaway:
              <ul>
                <li>In-game stuff like Totally Legal Tender or game items</li>
                <li>Whitelist spots</li>
                <li>Any other custom item like an NFT or a pair of really cool shoes!</li>
              </ul>
            </p>
          </Content>
        </Paper>
      </Cell>

    </LayoutGrid>
  </div>
</div>