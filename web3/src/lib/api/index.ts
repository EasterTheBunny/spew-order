import { createClient, dedupExchange, cacheExchange, fetchExchange } from '@urql/core';
import { authExchange } from '@urql/exchange-auth';
import type { Client, OperationResult } from '@urql/core';
import auth from './auth'

const client: Client = createClient({
  url: 'http://localhost:8080/graphql',
  exchanges: [
    dedupExchange,
    cacheExchange,
    authExchange({...auth}),
    fetchExchange,
  ],
});

export const getProject: (slug: string) => Promise<Project> = async (slug) => {
  const QUERY_GET_PROJECT = `
    query GetProjectData(
      $slug: ID!
    ) {
      project(slug: $slug) {
        slug
        name
        motto
        description
        prettyPath
        discord {
          guildId
          permissions
        }
        permissions
      }
    }`;

  try {
    const result = await client.query(QUERY_GET_PROJECT, { slug }).toPromise()

    if (result.error) {
      throw new Error(result.error)
    }

    if (!result.data) {
      throw new Error("no data found in result")
    }

    return result.data.project;
  } catch(e) {
    throw new Error(e)
  }
}

export const createProject: (input: NewProjectInput) => Promise<Project> = async (input) => {
  const MUTATION_NEW_PROJECT = `
    mutation NewProject(
      $input: NewProjectInput
    ) {
      createProject(
        input: $input
      ) {
        slug
        name
        motto
        description
        prettyPath
        permissions
      }
    }`;

  try {
    const result = await client.mutation(MUTATION_NEW_PROJECT, { input }).toPromise()

    if (result.error) {
      throw new Error(result.error)
    }

    if (!result.data) {
      throw new Error("no data found in result")
    }

    return result.data.createProject;
  } catch(e) {
    throw new Error(e)
  }
}

export const updateProject: (slug: string, input: NewProjectInput) => Promise<Project> = async (slug, input) => {
  const MUTATION_UPDATE_PROJECT = `
    mutation UpdateProject(
      $slug: ID!
      $input: NewProjectInput
    ) {
      updateProject(
        slug: $slug
        input: $input
      ) {
        slug
        name
        motto
        description
        prettyPath
        permissions
      }
    }`;

  try {
    const result = await client.mutation(MUTATION_UPDATE_PROJECT, { slug, input }).toPromise()

    if (result.error) {
      throw new Error(result.error)
    }

    if (!result.data) {
      throw new Error("no data found in result")
    }

    return result.data.updateProject;
  } catch(e) {
    throw new Error(e)
  }
}

export const createChallengeMessage: (address: string) => Promise<string> = async (address) => {
  const CHALLENGE_MSG = `
    query NewChallengeMessage($chain: ChainType!, $address: AddressScalar!) {
      challengeLoginMessage(chain: $chain, address: $address)
    }`;

  try {
    const result = await client.query(CHALLENGE_MSG, { chain: 'ETHEREUM', address }).toPromise()

    if (result.error) {
      throw new Error(result.error)
    }

    if (!result.data) {
      throw new Error("no data found in result")
    }

    return result.data.challengeLoginMessage;
  } catch(e) {
    throw new Error(e)
  }
}

export const verifyMessage: (address: string, message: string, signature: string) => Promise<TokenResult> = async (address, message, signature) => {
  const VERIFY_MSG = `
    mutation DoAddressLogin($chain: ChainType!, $address: AddressScalar!, $message: String!, $signature: String!) {
      loginFromSignedMessage(chain: $chain, address: $address, message: $message, signature: $signature) {
        token
      }
    }`

  const params = { chain: 'ETHEREUM', address, message, signature }

  try {
    const result = await client.mutation(VERIFY_MSG, params).toPromise()

    if (result.error) {
      throw new Error(result.error)
    }

    if (!result.data) {
      throw new Error("no data found in result")
    }

    return result.data.loginFromSignedMessage
  } catch(e) {
    throw new Error(e)
  }
}

export const registerDiscordBot: (slug: string, guild: string, permissions: number) => Promise<DiscordBot> = async (slug, guild, permissions) => {
  const DISCORD_CALLBACK = `  
    mutation DiscordCallback(
      $projectSlug: String!
      $guildID: String!
      $permissions: Int!
    ) {
      registerDiscordBot(
        projectSlug: $projectSlug
        guildID: $guildID
        permissions: $permissions
      ) {
        guildId
        permissions
        project {
          slug
        }
      }
    }`;

  const params = { projectSlug: slug, guildID: guild, permissions }

  try {
    const result = await client.mutation(DISCORD_CALLBACK, params).toPromise()

    if (result.error) {
      throw new Error(result.error)
    }

    if (!result.data) {
      throw new Error("no data found in result")
    }

    return result.data.registerDiscordBot
  } catch(e) {
    throw new Error(e)
  }
}

export const installDiscordBot: (extended: boolean) => Promise<string> = async (extended) => {
  const QUERY_INSTALL_BOT = `query InstallDiscordBot($extended: Boolean!) {
      initiateDiscordBotInstall(extended: $extended)
    }`;

  try {
    const result = await client.query(QUERY_INSTALL_BOT, { extended }).toPromise()

    if (result.error) {
      throw new Error(result.error)
    }

    if (!result.data) {
      throw new Error("no data found in result")
    }

    return result.data.initiateDiscordBotInstall;
  } catch(e) {
    throw new Error(e)
  }
}

export const getIdentity: () => Promise<Identity> = async () => {
  const QUERY_GET_IDENTITY = `query GetIdentity {
    identity {
      discordUser {
        id
        oauthToken
        oauthScopes
        guilds {
          id
          name
        }
      }
      projects {
        slug
        name
        motto
        description
        prettyPath
        discord {
          guildId
          permissions
        }
        permissions
      }
    }
  }`;

  try {
    const result = await client.query(QUERY_GET_IDENTITY).toPromise()

    if (result.error) {
      throw new Error(result.error)
    }

    if (!result.data) {
      throw new Error("no data found in result")
    }

    return result.data.identity;
  } catch(e) {
    throw new Error(e)
  }

  return null
}

export const registerDiscordOauth: (state: string, code: string) => Promise<DiscordUser> = async (state, code) => {
  const MUTATION_REGISTER = `mutation RegisterDiscordOauth($state: String!, $code: String!) {
    registerDiscordOauth(state: $state, code: $code) {
      id
      oauthToken
      oauthScopes
      guilds {
        id
        name
      }
    }
  }`;

  try {
    const result = await client.mutation(MUTATION_REGISTER, { state, code }).toPromise()

    if (result.error) {
      throw new Error(result.error)
    }

    if (!result.data) {
      throw new Error("no data found in result")
    }

    return result.data.registerDiscordOauth;
  } catch(e) {
    throw new Error(e)
  }
}
 
export default client