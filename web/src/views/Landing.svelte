<script type="ts">
  import LayoutGrid, { Cell } from "@smui/layout-grid/styled"
</script>

<main>
  <img class="hero" src="hero.jpg" alt="Cipher Mountain" />

  <LayoutGrid>
    <Cell align="left" spanDevices={{ desktop: 4, tablet: 6, phone: 12 }}>
      <h2 class="no-focus-border">Simple Interface</h2>
      <p>Easy tools to get you started. Trade with Bitcoin and Ethereum with ease against a growing number of other currencies.</p>
    </Cell>
    <Cell spanDevices={{ desktop: 8, tablet: 6, phone: 12 }}>
      <img src="cipher-mountain-ui-snapshot-zoom-book.png" alt="cryptocurrency exchange" />
    </Cell>
  </LayoutGrid>
  
  <LayoutGrid>
    <Cell spanDevices={{ desktop: 8, tablet: 6, phone: 12 }}>
      <img src="cipher-mountain-ui-snapshot-zoom-order.png" alt="cryptocurrency exchange" />
    </Cell>
    <Cell align="left" spanDevices={{ desktop: 4, tablet: 6, phone: 12 }}>
      <h2 class="no-focus-border">Trade Popular Currencies</h2>
      <p>Open limit and market orders on some of your favorite cryptocurrencies.</p>
      <ul>
        <li>Bitcoin</li>
        <li>Ethereum</li>
        <li>Bitcoin Cash</li>
        <li>Dogecoin</li>
        <li>Uniswap</li>
      </ul>
    </Cell>
  </LayoutGrid>

  <LayoutGrid>
    <Cell spanDevices={{ desktop: 4, tablet: 12, phone: 12 }}>
      <h3>Biometric Security</h3>
      <p>Industry leading security for supported devices</p>
    </Cell>
    <Cell spanDevices={{ desktop: 4, tablet: 12, phone: 12 }}>
      <h3>Realtime Data</h3>
      <p>Receive up to the second market information</p>
    </Cell>
    <Cell spanDevices={{ desktop: 4, tablet: 12, phone: 12 }}>
      <h3>Integrity First</h3>
      <p>Your assets are safe and your data is secure</p>
    </Cell>
  </LayoutGrid>
</main>

<footer>

  <LayoutGrid>
    <Cell spanDevices={{ desktop: 6, tablet: 12, phone: 12 }}>
      <h4>About</h4>
      <p>
        <a href="https://www.ciphermtn.com/">Home</a>
      </p>
      <p>
        <a href="https://www.ciphermtn.com/our-vision">Our Vision</a>
      </p>
    </Cell>
    <Cell spanDevices={{ desktop: 6, tablet: 12, phone: 12 }}>
      <h4>Information</h4>
      <p>
        <a href="https://www.ciphermtn.com/materials-for-learning">Cryptocurrency Education</a>
      </p>
      <p>
        <a href="https://www.ciphermtn.com/privacy-policy">Privacy Policy</a>
      </p>
    </Cell>
  </LayoutGrid>

</footer>
  
<style>
  main {
		padding: 1em;
		max-width: 240px;
		margin: 0 auto;
	}

  footer {
    padding: 1em;
    width: 100%;
    background-color: #333333;
  }

	@media (min-width: 640px) {
		main {
			max-width: none;
		}
	}

  * :global(.info-content) {
    text-align: left;
  }

  * :global(.no-focus-border:focus-visible) {
    outline: none;
  }

  img {
    width: 100%;
    margin-top: 25px;
  }

  img.hero {
    width: 110%;
    margin-top: 0;
    margin: -20px;
  }
</style>