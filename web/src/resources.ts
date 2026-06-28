import { titleCase } from './format'
import { COMMON_CALIBERS } from './ammoseek'

export type FieldType = 'text' | 'number' | 'money' | 'date' | 'textarea' | 'select' | 'bool' | 'firearmRef' | 'combo' | 'multiCheck'

export interface Field {
  name: string
  label: string
  type?: FieldType
  options?: { value: string; label: string }[]
  showIf?: (item: any) => boolean
}

export interface ResourceConfig {
  key: string
  label: string
  singular: string
  emoji: string
  fields: Field[]
  title: (item: any) => string
  subtitle: (item: any) => string
}

const opts = (vals: string[]) => vals.map((v) => ({ value: v, label: v ? titleCase(v) : '—' }))
const list = (vals: string[]) => vals.map((v) => ({ value: v, label: v }))

// Combo (dropdown + "Other…") option sets.
const caliberOptions = COMMON_CALIBERS.map((c) => ({ value: c.label, label: c.label }))
const gunMakerOptions = list([
  'Glock', 'Smith & Wesson', 'SIG Sauer', 'Ruger', 'Springfield Armory', 'Colt', 'Beretta', 'CZ',
  'Heckler & Koch', 'FN', 'Remington', 'Mossberg', 'Winchester', 'Savage Arms', 'Kimber', 'Walther',
  'Taurus', 'Browning', 'Benelli', 'Henry', 'Marlin', 'Daniel Defense', 'Aero Precision',
  'Palmetto State Armory', 'Rock Island Armory', 'Wilson Combat',
])
const ammoBrandOptions = list([
  'Federal', 'Winchester', 'Remington', 'Hornady', 'CCI', 'Speer', 'PMC', 'Fiocchi', 'Sellier & Bellot',
  'Magtech', 'Blazer', 'Wolf', 'TulAmmo', 'Norma', 'Barnes', 'Nosler', 'SIG Sauer', 'Aguila',
  'Browning', 'American Eagle',
])
const knifeMakerOptions = list([
  'Benchmade', 'Spyderco', 'Kershaw', 'Zero Tolerance', 'CRKT', 'Cold Steel', 'Gerber', 'Buck', 'ESEE',
  'KA-BAR', 'Ontario', 'SOG', 'Victorinox', 'Leatherman', 'Microtech', 'WE Knife', 'Civivi', 'Kizer',
  'Opinel', 'Case', 'Böker', 'Chris Reeve', 'Hinderer', 'Emerson', 'TOPS', 'Morakniv',
])
const knifeSteelOptions = list([
  'CPM MagnaCut', 'S30V', 'S35VN', 'S45VN', 'CPM-20CV', 'M390', '154CM', 'CPM-3V', 'D2', 'VG-10',
  'AUS-8', 'AUS-10', '14C28N', '8Cr13MoV', 'Nitro-V', '1095', '5160', 'A2', 'O1', 'Elmax', 'K390',
  'LC200N', 'ZDP-189', 'Maxamet',
])
const shellLengthOptions = list(['2½"', '2¾"', '3"', '3½"'])

export const RESOURCE_KEYS = ['firearms', 'ammo', 'knives', 'accessories'] as const

export const RESOURCES: Record<string, ResourceConfig> = {
  firearms: {
    key: 'firearms',
    label: 'Firearms',
    singular: 'Firearm',
    emoji: '🔫',
    fields: [
      { name: 'nickname', label: 'Nickname' },
      { name: 'manufacturer', label: 'Manufacturer', type: 'combo', options: gunMakerOptions },
      { name: 'model', label: 'Model' },
      { name: 'kind', label: 'Type', type: 'select', options: opts(['pistol', 'rifle', 'shotgun', 'nfa', 'other']) },
      { name: 'caliber', label: 'Caliber', type: 'combo', options: caliberOptions },
      { name: 'shellLengths', label: 'Shell lengths (chamber)', type: 'multiCheck', options: shellLengthOptions, showIf: (f) => f.kind === 'shotgun' },
      { name: 'serialNumber', label: 'Serial number' },
      { name: 'finish', label: 'Finish' },
      { name: 'status', label: 'Status', type: 'select', options: opts(['owned', 'sold', 'loaned', 'pending']) },
      { name: 'isNfa', label: 'NFA item', type: 'bool' },
      { name: 'nfaType', label: 'NFA type', type: 'select', options: opts(['', 'suppressor', 'sbr', 'sbs', 'mg', 'aow']) },
      { name: 'taxStampDate', label: 'Tax stamp date', type: 'date' },
      { name: 'acquiredDate', label: 'Acquired date', type: 'date' },
      { name: 'acquiredPriceCents', label: 'Price', type: 'money' },
      { name: 'acquiredFrom', label: 'Acquired from' },
      { name: 'notes', label: 'Notes', type: 'textarea' },
    ],
    title: (f) => f.nickname || [f.manufacturer, f.model].filter(Boolean).join(' ') || 'Firearm',
    subtitle: (f) =>
      [f.shellLengths?.length ? `${f.caliber} (${f.shellLengths.join(', ')})` : f.caliber, f.kind].filter(Boolean).join(' · '),
  },
  ammo: {
    key: 'ammo',
    label: 'Ammo',
    singular: 'Ammo',
    emoji: '📦',
    fields: [
      { name: 'name', label: 'Name' },
      { name: 'caliber', label: 'Caliber', type: 'combo', options: caliberOptions },
      { name: 'brand', label: 'Brand', type: 'combo', options: ammoBrandOptions },
      { name: 'bulletType', label: 'Bullet type', type: 'select', options: opts(['FMJ', 'JHP', 'HP', 'SP', 'match', 'birdshot', 'buckshot', 'slug', 'other']) },
      { name: 'shellLength', label: 'Shell length', type: 'combo', options: shellLengthOptions },
      { name: 'grainWeight', label: 'Grain weight', type: 'number' },
      { name: 'quantityOnHand', label: 'Rounds on hand', type: 'number' },
      { name: 'lowStockThreshold', label: 'Low-stock alert (rds)', type: 'number' },
      { name: 'lotNumber', label: 'Lot number' },
      { name: 'acquiredDate', label: 'Acquired date', type: 'date' },
      { name: 'acquiredPriceCents', label: 'Price', type: 'money' },
      { name: 'acquiredFrom', label: 'Acquired from' },
      { name: 'notes', label: 'Notes', type: 'textarea' },
    ],
    title: (a) => a.name || a.caliber || 'Ammo',
    subtitle: (a) => [a.caliber, a.shellLength, a.bulletType, a.grainWeight ? `${a.grainWeight}gr` : '', a.quantityOnHand ? `${a.quantityOnHand} rds` : ''].filter(Boolean).join(' · '),
  },
  knives: {
    key: 'knives',
    label: 'Knives',
    singular: 'Knife',
    emoji: '🔪',
    fields: [
      { name: 'nickname', label: 'Nickname' },
      { name: 'type', label: 'Type', type: 'select', options: opts(['folding', 'fixed', 'automatic', 'balisong', 'multitool', 'machete', 'other']) },
      { name: 'manufacturer', label: 'Manufacturer', type: 'combo', options: knifeMakerOptions },
      { name: 'model', label: 'Model' },
      { name: 'bladeSteel', label: 'Blade steel', type: 'combo', options: knifeSteelOptions },
      { name: 'bladeLengthIn', label: 'Blade length (in)' },
      { name: 'serialNumber', label: 'Serial number' },
      { name: 'status', label: 'Status', type: 'select', options: opts(['owned', 'sold', 'loaned', 'pending']) },
      { name: 'lastSharpenedDate', label: 'Last sharpened', type: 'date' },
      { name: 'acquiredDate', label: 'Acquired date', type: 'date' },
      { name: 'acquiredPriceCents', label: 'Price', type: 'money' },
      { name: 'acquiredFrom', label: 'Acquired from' },
      { name: 'notes', label: 'Notes', type: 'textarea' },
    ],
    title: (k) => k.nickname || [k.manufacturer, k.model].filter(Boolean).join(' ') || 'Knife',
    subtitle: (k) => [titleCase(k.type), k.bladeSteel].filter(Boolean).join(' · '),
  },
  accessories: {
    key: 'accessories',
    label: 'Accessories',
    singular: 'Accessory',
    emoji: '🔭',
    fields: [
      { name: 'name', label: 'Name' },
      { name: 'category', label: 'Category', type: 'select', options: opts(['optic', 'light', 'laser', 'sling', 'magazine', 'trigger', 'stock', 'choke', 'case', 'cleaning', 'other']) },
      { name: 'manufacturer', label: 'Manufacturer' },
      { name: 'model', label: 'Model' },
      { name: 'serialNumber', label: 'Serial number' },
      { name: 'firearmId', label: 'On firearm', type: 'firearmRef' },
      { name: 'quantity', label: 'Quantity', type: 'number' },
      { name: 'valueCents', label: 'Value', type: 'money' },
      { name: 'acquiredFrom', label: 'Acquired from' },
      { name: 'notes', label: 'Notes', type: 'textarea' },
    ],
    title: (a) => a.name || 'Accessory',
    subtitle: (a) => [titleCase(a.category), a.manufacturer].filter(Boolean).join(' · '),
  },
}
