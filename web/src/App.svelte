<script lang="ts">
	import Root from "./views/Root.svelte"
	import Router from "./views/Router.svelte"
  import oidc from "./oidc"
  import initExchangeAPI from "./exchange"
  import initMarketDataController from "./market"
  import { initLocalizationContext } from './i18n';

  export let client_id: string = process.env.AUTH0_CLIENTID;
  export let authority: string = process.env.AUTH0_DOMAIN;

  initLocalizationContext()
  const user = oidc({
    client_id: client_id,
    authority: authority,
    metadata: {
      end_session_endpoint: `${authority}/v2/logout?client_id=${client_id}`,
      issuer: `${authority}/`,
      authorization_endpoint: `${authority}/authorize`,
      token_endpoint: `${authority}/oauth/token`,
      userinfo_endpoint: `${authority}/userinfo`,
      jwks_uri: `${authority}/.well-known/jwks.json`,
      registration_endpoint: `${authority}/oidc/register`,
      revocation_endpoint: `${authority}/oauth/revoke`,
      scopes_supported: ["openid","profile","offline_access","name","given_name","family_name","nickname","email","email_verified","picture","created_at","identities","phone","address"],
      response_types_supported: ["code","token","id_token","code token","code id_token","token id_token","code token id_token"],
      code_challenge_methods_supported: ["S256","plain"],
      response_modes_supported: ["query","fragment","form_post"],
      subject_types_supported: ["public"],
      id_token_signing_alg_values_supported: ["HS256","RS256"],
      token_endpoint_auth_methods_supported: ["client_secret_basic","client_secret_post"],
      claims_supported: ["aud","auth_time","created_at","email","email_verified","exp","family_name","given_name","iat","identities","iss","name","nickname","phone_number","picture","sub"],
    },
  })
  initExchangeAPI(user)
  initMarketDataController()

</script>

<Root>
	<Router url="" />
</Root>