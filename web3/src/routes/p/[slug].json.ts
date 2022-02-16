import type { RequestHandler } from '@sveltejs/kit'

const sites: Project[] = [{ slug: "test" }]

export const get: RequestHandler = ({ params }) => {
  const { slug } = params

  const project = sites.find((itm) => itm.slug === slug)

  if (slug) return { body: project }

  // Not returning is equivalent to a 404 response.
  // https://kit.svelte.dev/docs#routing-endpoints
}