import { writable } from 'svelte/store'

let dialogOpen = writable(false)

export const close = () => {
  dialogOpen.update(() => false)
}

export const open = () => {
  dialogOpen.update(() => true)
}

export default {
  dialogOpen,
  close,
  open,
}