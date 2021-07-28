import {
  UserManager,
  UserManagerSettings,
  WebStorageStateStore,
} from "oidc-client"

export default class OidcService {
  public manager: UserManager
  private settings: UserManagerSettings = {
    response_type: 'token id_token',
    scope: 'openid profile',
    silent_redirect_uri: `${window.location.protocol}//${window.location.hostname}${window.location.port ? `:${window.location.port}` : ''}/silent_renew.html`,
    automaticSilentRenew: true,
    filterProtocolClaims: true,
    loadUserInfo: true,
    stateStore: new WebStorageStateStore({ store: window.localStorage }),
    userStore: new WebStorageStateStore({ store: window.localStorage }),
  }

  constructor(config: UserManagerSettings) {
    const settings: UserManagerSettings = {
      redirect_uri: this.buildURI("/login"),
    }
    this.settings = Object.assign(settings, this.settings, config)
  }

  private buildURI(path: string): string {
    const base = `${window.location.protocol}//${window.location.hostname}${window.location.port ? `:${window.location.port}` : ''}`
    return base + path
  }

  public initialize(): void {
    this.manager = new UserManager(this.settings)
  }

  public signIn(): void {
    this.manager.signinRedirect()
  }

  public claim(id: string): void {
    this.manager.signinRedirect({ redirect_uri: this.buildURI("/claim?token="+id) })
  }

  public signOut(): void {
    //this.manager.removeUser()
    //this.manager.revokeAccessToken()
    this.manager.signoutRedirect().then(() => {
      //console.log("signed out")
    })
  }
}
  