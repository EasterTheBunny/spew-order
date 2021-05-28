import { derived, Readable, Writable, writable } from "svelte/store"
import type { User } from "oidc-client"

export default class UserService {
  private state: Writable<UserState>
  public loggedIn: Readable<boolean>
  public user: Readable<User>

  constructor(oidc: OidcService) {
    oidc.initialize()
    this.state = this.createState()
    this.loggedIn = this.createLoggedInReadable()
    this.user = this.createUserReadable(oidc)

    oidc.manager.events.addUserLoaded(this.onUserLoaded)
    oidc.manager.events.addSilentRenewError(this.onSilentRenewError)
    oidc.manager.events.addAccessTokenExpired(this.onAccessTokenExpired)
    oidc.manager.events.addUserUnloaded(this.onUserUnloaded)
    oidc.manager.events.addUserSignedOut(this.onUserSignedOut)

    oidc.manager.getUser().then((u: User) => {
      if (!!u && !u.expired) {
        this.onUserLoaded(u)
      }
    })
  }

  private createState(): Writable<UserState> {
    const { subscribe, set, update } = writable<UserState>({
      user: null,
      isLoadingUser: false,
    })

    return { set, subscribe, update }
  }

  private createLoggedInReadable(): Readable<boolean> {
    return derived([this.state], ([$state]) => {
      return !$state.isLoadingUser && !!$state.user
    })
  }

  // callback for the userManager's getUser.catch
  private errorCallback(error: Error): void {
    console.error(`svelte-oidc: Error loading user in oidcMiddleware: ${error.message}`);
  }

  private createUserReadable(oidc: OidcService): Readable<User> {
    return derived([this.state], ([$state]) => {
      return $state.user
    })
  }

  // event callback when the user has been loaded (on silent renew or redirect)
  private onUserLoaded: (user: any) => void = (user) => {
    const { update } = this.state
    update((state: UserState) => {
      return {
        user: user,
        isLoadingUser: false,
      }
    })
  }

  // event callback when silent renew errored
  private onSilentRenewError: (error: any) => void = (error) => {
    const { update } = this.state
    update((state: UserState) => {
      state.user = null
      state.isLoadingUser = false
      return state
    })
  }

  // event callback when the access token expired
  private onAccessTokenExpired: () => void = () => {
    const { update } = this.state
    update((state: UserState) => {
      state.user = null
      state.isLoadingUser = false
      return state
    })
  }

  // event callback when the user is logged out
  private onUserUnloaded: () => void = () => {
    const { update } = this.state
    update((state: UserState) => {
      state.user = null
      state.isLoadingUser = false
      return state
    })
  }

  // event callback when the user is signed out
  private onUserSignedOut: () => void = () => {
    const { update } = this.state
    update((state: UserState) => {
      state.user = null
      state.isLoadingUser = false
      return state
    })
  }

}
