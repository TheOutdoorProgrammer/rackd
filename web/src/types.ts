export interface Status {
  initialized: boolean
  unlocked: boolean
}

export interface Firearm {
  id: number
  nickname: string
  manufacturer: string
  model: string
  kind: string
  caliber: string
  serialNumber: string
  finish: string
  acquiredDate: string
  acquiredPriceCents: number
  acquiredFrom: string
  status: string
  isNfa: boolean
  nfaType: string
  taxStampDate: string
  notes: string
  createdAt: string
  updatedAt: string
}

export interface Ammo {
  id: number
  name: string
  caliber: string
  brand: string
  bulletType: string
  shellLength: string
  shotSize: string
  shotWeight: string
  grainWeight: number
  fps: number
  quantityOnHand: number
  lowStockThreshold: number
  lotNumber: string
  acquiredDate: string
  acquiredPriceCents: number
  acquiredFrom: string
  notes: string
  createdAt: string
  updatedAt: string
}

export interface Knife {
  id: number
  nickname: string
  type: string
  manufacturer: string
  model: string
  bladeSteel: string
  bladeLengthIn: string
  serialNumber: string
  acquiredDate: string
  acquiredPriceCents: number
  acquiredFrom: string
  lastSharpenedDate: string
  status: string
  notes: string
  createdAt: string
  updatedAt: string
}

export interface Accessory {
  id: number
  name: string
  category: string
  manufacturer: string
  model: string
  serialNumber: string
  valueCents: number
  quantity: number
  acquiredFrom: string
  notes: string
  createdAt: string
  updatedAt: string
}

export interface Attachment {
  id: number
  ownerType: string
  ownerId: number
  kind: string
  filename: string
  contentType: string
  sizeBytes: number
  createdAt: string
  updatedAt: string
  cover: boolean
}

export interface AmmoLink {
  ammo: Ammo
  note: string
}

export interface Summary {
  counts: Record<string, number>
  totalValueCents: number
  lowStockAmmo: number
}

export interface SearchResults {
  firearms: Firearm[]
  ammo: Ammo[]
  knives: Knife[]
  accessories: Accessory[]
}

// A loosely-typed item for the generic list/form/detail components.
export type Item = Record<string, any> & { id: number }

export interface SpecSearchResult {
  title: string
  description: string
}

export interface SpecRow {
  label: string
  value: string
}

export interface SpecPage {
  title: string
  url: string
  specs: SpecRow[]
  fill: Record<string, string>
}
