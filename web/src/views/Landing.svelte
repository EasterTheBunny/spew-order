<script type="ts">
  import LayoutGrid, { Cell } from "@smui/layout-grid/styled"
  import Paper, { Title, Content } from "@smui/paper/styled"
</script>

<main>
  <LayoutGrid>
    <Cell span={6}>
      <div class="info-container">
        <Paper color="primary">
          <Title class="no-focus-border">Current Features</Title>
          <Content class="info-content">
            <dl>
              <dt>Order Types:</dt>
              <dd>Limit and Market</dd>

              <dt>Currencies:</dt>
              <dd>BTC, ETH, BCH, DOGE</dd>

              <dt>Funding:</dt>
              <dd>BTC, ETH, BCH, DOGE</dd>

              <dt>Withdrawals:</dt>
              <dd>BTC, ETH, BCH, DOGE</dd>

              <dt>Other:</dt>
              <dd>Realtime market prices</dd>
            </dl>
            <img src="exchange-screenshot.png" alt="cryptocurrency exchange" />
          </Content>
        </Paper>
      </div>
    </Cell>
    <Cell span={6}>
      <iframe src="https://cdn.forms-content.sg-form.com/35a20b74-0f35-11ec-8a80-3a9a2a8bfcfe" style="width:100%;height:600px;border:none;"/>
    </Cell>
  </LayoutGrid>
</main>
  
<style>
  main {
		text-align: center;
		padding: 1em;
		max-width: 240px;
		margin: 0 auto;
	}

	@media (min-width: 640px) {
		main {
			max-width: none;
		}
	}
  .info-container {
    padding-top: 20px;
  }

  * :global(.info-content) {
    text-align: left;
  }

  * :global(.no-focus-border:focus-visible) {
    outline: none;
  }
  dl {
    width: 100%;
    overflow: hidden;
    padding: 0;
    margin: 0
  }
  dt {
    float: left;
    width: 35%;
    /* adjust the width; make sure the total of both is 100% */
    padding: 0;
    margin: 0;
    margin-right: 10px;
    text-align: right;
    font-weight: bold;
  }
  dd {
    float: left;
    width: 50%;
    /* adjust the width; make sure the total of both is 100% */
    padding: 0;
    margin: 0
  }

  img {
    width: 100%;
    margin-top: 25px;
  }
</style>