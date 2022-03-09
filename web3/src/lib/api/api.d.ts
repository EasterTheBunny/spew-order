type TokenResult {
  token: string
}

type Project {
  slug: string
  name: string
  motto: string
  description: string
  prettyPath: string
  permissions: string[]
  discord?: any
}

type NewProjectInput {
  name?: string
  motto?: string
  description?: string
  prettyPath?: string
  website?: string
}

type DiscordBot {
  guildId: string
  permissions: string
  project?: Project
}

type Identity {
  discordUser?: DiscordUser
  projects: Project[]
}

type DiscordUser {
  id: string
  oauthToken: boolean
  oauthScopes: string
  guilds: DiscordGuild[]
}

type DiscordGuild {
  id: string
  name: string
}
