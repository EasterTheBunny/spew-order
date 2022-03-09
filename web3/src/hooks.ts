/** @type {import('@sveltejs/kit').GetSession} */
import { registerDiscordOauth } from '$lib/api'

export async function getSession(event) {
  let retryOauthClient = true

  // specific server-side processing of oauth requests
  if (event.path === "/chupagoat/install") {
    if (event.query.has('code') && event.query.has('state')) {
      const code = event.query.get('code')
      const state = event.query.get('state')

      try {
        // attempting to register the bot
        const response = await registerDiscordOauth(state, code)
        retryOauthClient = false
      } catch(e) {
        console.log(e.message)
      }
    }
  }

  return event.locals.oauth
    ? {
        oauth: {
          retryOauthClient,
        }
      }
    : { oauth: {retryOauthClient}};
}