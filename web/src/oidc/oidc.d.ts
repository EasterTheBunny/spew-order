declare class OidcService {
    public manager: UserManager
    constructor() {
        // empty
    }

    public initialize(): void
    public signIn(): void
    public claim(id: string): void
    public signOut(): void
}

declare class UserService {
    constructor() {
        // empty
    }
}

interface UserState {
    user: User
    isLoadingUser: boolean
}

interface OidcContext {
    oidc: OidcService
    user: Readable<User>
    loggedIn: Readable<boolean>
}