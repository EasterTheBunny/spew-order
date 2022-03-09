import state from '$lib/state'; // stores related to app state, auth state
import type { LoadInput, LoadOutput } from '@sveltejs/kit/types.internal';

export const isServerSide = typeof window === "undefined"

let auth = { ...state.initialAuthState };
state.auth.subscribe(authState => auth = authState);

export async function testFunc({ page }: LoadInput): Promise<LoadOutput> {
  console.log("testFunc")

  return {}
}

export async function authGuard(): boolean {
  const loggedIn = auth.loggedIn;

  if (isServerSide) {
    // no server-side authentication at this time
    return true
  }

  if (loggedIn) {
    return true
  } else {
    return false
  }
}

export default {
  authGuard
}