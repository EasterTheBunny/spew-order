<script type="ts">
  import Button, { Label } from '@smui/button/styled'
  import { getOidc } from "../oidc"
  import { getLocalization } from '../i18n'
  
  const { oidc } = getOidc()
  const {t} = getLocalization()
</script>

<h4>{$t('LoginToView')}</h4>
<Button on:click={() => oidc.signIn()} variant="unelevated">
  <Label>{$t('Login')}</Label>
</Button>
