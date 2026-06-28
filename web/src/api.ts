import type { Status, Attachment, AmmoLink, Summary, SearchResults, SpecSearchResult, SpecPage } from './types'

async function req<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(path, init)
  if (!res.ok) {
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
export const photoURL = (id: number) => `/api/photos/${id}`
export const thumbURL = (id: number) => `/api/photos/${id}/thumb`
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

// --- meta ---
export const getSummary = () => req<Summary>('/api/summary')
export const search = (q: string) => req<SearchResults>(`/api/search?q=${encodeURIComponent(q)}`)

// --- firearm spec lookup (Wikipedia + DBpedia, cached, key-less) ---
export const specsSearch = (q: string) => req<SpecSearchResult[]>(`/api/specs/search?q=${encodeURIComponent(q)}`)
export const specsPage = (title: string) => req<SpecPage>(`/api/specs/page?title=${encodeURIComponent(title)}`)
export const specCacheStats = () => req<{ count: number }>('/api/specs/cache')
export const specClearCache = () => req<{ cleared: number }>('/api/specs/cache', { method: 'DELETE' })
