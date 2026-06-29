import type { Status, Attachment, AmmoLink, Accessory, Ammo, Firearm, Summary, SearchResults, SpecSearchResult, SpecPage } from './types'

// When an authenticated request 401s (session expired or the vault re-locked on
// a server restart), the app drops straight back to the PIN screen.
let onUnauthorized: (() => void) | null = null
export const setUnauthorizedHandler = (fn: () => void) => {
  onUnauthorized = fn
}

async function req<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(path, init)
  if (!res.ok) {
    if (res.status === 401) onUnauthorized?.()
    const body = (await res.json().catch(() => ({}))) as { error?: string }
    throw new Error(body.error ?? res.statusText)
  }
  if (res.status === 204) return undefined as T
  return res.json() as Promise<T>
}

const json = (method: string, body: unknown): RequestInit => ({
  method,
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify(body),
})

// --- auth ---
export const getStatus = () => req<Status>('/api/status')
export const setupVault = (pin: string) => req<Status>('/api/auth/setup', json('POST', { pin }))
export const unlockVault = (pin: string) => req<Status>('/api/auth/unlock', json('POST', { pin }))
export const lockVault = () => req<void>('/api/auth/lock', { method: 'POST' })

// --- generic resource CRUD ---
export const listItems = <T>(resource: string) => req<T[]>(`/api/${resource}`)
export const getItem = <T>(resource: string, id: number) => req<T>(`/api/${resource}/${id}`)
export const createItem = <T>(resource: string, body: unknown) => req<T>(`/api/${resource}`, json('POST', body))
export const updateItem = <T>(resource: string, id: number, body: unknown) =>
  req<T>(`/api/${resource}/${id}`, json('PUT', body))
export const deleteItem = (resource: string, id: number) =>
  req<void>(`/api/${resource}/${id}`, { method: 'DELETE' })

// --- photos ---
export const listPhotos = (owner: string, id: number) =>
  req<Attachment[]>(`/api/photos?owner=${owner}&id=${id}`)
export const deletePhoto = (id: number) => req<void>(`/api/photos/${id}`, { method: 'DELETE' })
export const setCover = (id: number) => req<void>(`/api/photos/${id}/cover`, { method: 'PUT' })
export const rotatePhoto = (id: number) => req<void>(`/api/photos/${id}/rotate`, { method: 'POST' })
// Optional version (an attachment's updatedAt) cache-busts the URL after a rotate.
export const photoURL = (id: number, v?: string) => `/api/photos/${id}${v ? `?v=${encodeURIComponent(v)}` : ''}`
export const thumbURL = (id: number, v?: string) => `/api/photos/${id}/thumb${v ? `?v=${encodeURIComponent(v)}` : ''}`
export async function uploadPhoto(owner: string, id: number, file: Blob, filename: string): Promise<Attachment> {
  const fd = new FormData()
  fd.append('file', file, filename)
  return req<Attachment>(`/api/photos?owner=${owner}&id=${id}`, { method: 'POST', body: fd })
}

// --- ammo links ---
export const listFirearmAmmo = (firearmId: number) => req<AmmoLink[]>(`/api/firearms/${firearmId}/ammo`)
export const linkAmmo = (firearmId: number, ammoId: number, note: string) =>
  req<void>(`/api/firearms/${firearmId}/ammo/${ammoId}`, json('PUT', { note }))
export const unlinkAmmo = (firearmId: number, ammoId: number) =>
  req<void>(`/api/firearms/${firearmId}/ammo/${ammoId}`, { method: 'DELETE' })

// --- ammo stock + reverse (which guns use an ammo line) ---
export const adjustAmmo = (id: number, delta: number) =>
  req<Ammo>(`/api/ammo/${id}/adjust`, json('POST', { delta }))
export const listFirearmsForAmmo = (ammoId: number) =>
  req<Firearm[]>(`/api/ammo/${ammoId}/firearms`)

// --- accessory ↔ firearm mounts (many-to-many, capped at the accessory's quantity) ---
export const listAccessoriesForFirearm = (firearmId: number) =>
  req<Accessory[]>(`/api/firearms/${firearmId}/accessories`)
export const listFirearmsForAccessory = (accessoryId: number) =>
  req<Firearm[]>(`/api/accessories/${accessoryId}/firearms`)
export const linkAccessory = (firearmId: number, accessoryId: number) =>
  req<void>(`/api/firearms/${firearmId}/accessories/${accessoryId}`, { method: 'PUT' })
export const unlinkAccessory = (firearmId: number, accessoryId: number) =>
  req<void>(`/api/firearms/${firearmId}/accessories/${accessoryId}`, { method: 'DELETE' })

// --- meta ---
export const getSummary = () => req<Summary>('/api/summary')
export const search = (q: string) => req<SearchResults>(`/api/search?q=${encodeURIComponent(q)}`)

// --- firearm spec lookup (Wikipedia + DBpedia, cached, key-less) ---
export const specsSearch = (q: string) => req<SpecSearchResult[]>(`/api/specs/search?q=${encodeURIComponent(q)}`)
export const specsPage = (title: string) => req<SpecPage>(`/api/specs/page?title=${encodeURIComponent(title)}`)
export const specCacheStats = () => req<{ count: number }>('/api/specs/cache')
export const specClearCache = () => req<{ cleared: number }>('/api/specs/cache', { method: 'DELETE' })
