import { createChallengeMessage, verifyMessage } from '$lib/api'
import state from '$lib/state'; // stores related to app state, auth state

export const ethereumActive = typeof window !== "undefined" && typeof window.ethereum !== "undefined"

let auth = { ...state.initialAuthState };
state.auth.subscribe(authState => auth = authState);

if (ethereumActive) {
  ethereum.on('accountsChanged', function (accounts) {
    if (!accounts || accounts.length == 0) {
      clearLoginData()
    } else {
      updateAddressInAuth(accounts[0])
    }
  })
}

export const shortenAddress: (addr: string) => string = (addr) => {
  return `${addr.slice(0, 5)}...${addr.slice(-3)}`
}

export const updateAddressInAuth = (address) => {
  state.auth.update((oldAuth) => {
    return {
      ...oldAuth,
      token: address == oldAuth.address ? oldAuth.token : '',
      loggedIn: address == oldAuth.address ? oldAuth.token != '' : false,
      address,
    }
  })
}

export const updateTokenInAuth = (token) => {
  state.auth.update((oldAuth) => {
    return {
      ...oldAuth,
      token,
      loggedIn: token != '',
    }
  })
}

export const clearLoginData = () => {
  state.auth.update((oldAuth) => {
    return {
      ...oldAuth,
      address: '', // added because: when a user disconnects their wallet, the site said it was still connected
      token: '',
      loggedIn: false,
    }
  })
}

export const loginWithEthereum = async () => {
  if (ethereumActive) {
    try {
      const address = auth.address;
      const message = await createChallengeMessage(address)
      const signature = await ethereum.request({ method: 'personal_sign', params: [address, message] })
      const { token } = await verifyMessage(address, message, signature)

      updateTokenInAuth(token)
    } catch(e) {
      // handle error
      throw new Error(e)
    }
  }
}

export const checkConnectionStatus = async () => {
  if (ethereumActive) {
    if (ethereum.selectedAddress !== null) {
      updateAddressInAuth(ethereum.selectedAddress)
    } else {
      clearLoginData()
    }
  }
}

export const promptForAddress = async () => {
  if (ethereumActive) {
    try {
      const accounts = await ethereum.request({ method: 'eth_requestAccounts' });
      updateAddressInAuth(accounts[0])
    } catch(e) {
      throw new Error(e)
    }
  }
}

export const WALLET_TITLE = 'Connecting your Wallet';
export const WALLET_TEXT = 'This step enables READ ONLY access to the contents of your wallet. Any further actions or permissions will require a signature request. You can disconnect your wallet at any time using the disconnect feature in your wallet.';
export const AUTH_TITLE = 'Login with Web3';
export const AUTH_TEXT = 'To perform any administrative functions you will need to provide proof that you control the connected account. This is done in the form of a signing request with your wallet. Be CAUTIOUS of what you sign! Verify that the actions you are giving permissions for are what you expect and ensure that the domain matches what you see in the URL bar and consistent with https://www.ciphermtn.com.';

export default {
  loginWithEthereum,
  clearLoginData,
  updateAddressInAuth,
  updateTokenInAuth,
  shortenAddress,
  checkConnectionStatus,
  promptForAddress,
}