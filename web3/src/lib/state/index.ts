import { writable } from 'svelte/store'

const isServerSide = typeof window === "undefined"
export const TOKEN_KEY = 'ciphermtn:api_token'

export const initialAuthState: AuthState = {
  userData: {},
  loggedIn: false,
  token: "",
  address: "",
  permissions: {},
}

export let auth = writable({
  ...initialAuthState
});

if (!isServerSide) {
  const { update } = auth

  const data = window.localStorage.getItem(TOKEN_KEY)

  if (data !== null) {
    update((state: AuthState) => {
      const storedState: AuthState = JSON.parse(data)

      return {
        ...state,
        ...storedState,
      }
    })
  }

  auth.subscribe((newState) => {
    if (!!window) {
      window.localStorage.setItem(TOKEN_KEY, JSON.stringify(newState))
    }
  })
}

export default {
  auth,
  initialAuthState,
}