// AmmoSeek deep links.
//
// We never scrape AmmoSeek (their ToS bans bots); this only opens their
// by-caliber page in the user's own browser. The slugs are AmmoSeek's own
// canonical caliber slugs, verified from their live URLs — they're idiosyncratic
// (e.g. 45acp / 22lr have no hyphens, but 9mm-luger / 12-gauge do, and
// 300aac-blackout jams "300aac" together), so this is a lookup table, not an
// algorithm. Unknown calibers return null → no button rendered (never a 404).

export interface CaliberOption {
  label: string
  slug: string
}

// Doubles as the caliber dropdown source.
export const COMMON_CALIBERS: CaliberOption[] = [
  { label: '9mm Luger', slug: '9mm-luger' },
  { label: '.380 ACP', slug: '380-auto' },
  { label: '.38 Special', slug: '38-special' },
  { label: '.357 Magnum', slug: '357-magnum' },
  { label: '.40 S&W', slug: '40sw' },
  { label: '10mm Auto', slug: '10mm-auto' },
  { label: '.44 Magnum', slug: '44-magnum' },
  { label: '.45 ACP', slug: '45acp' },
  { label: '.22 LR', slug: '22lr' },
  { label: '.223 Remington', slug: '223-remington' },
  { label: '5.56x45mm NATO', slug: '5.56x45mm-nato' },
  { label: '.308 Winchester', slug: '308-winchester' },
  { label: '7.62x51mm NATO', slug: '308-winchester' },
  { label: '7.62x39mm', slug: '7.62x39mm' },
  { label: '.300 Blackout', slug: '300aac-blackout' },
  { label: '.350 Legend', slug: '350-legend' },
  { label: '6.5 Creedmoor', slug: '6.5mm-creedmoor' },
  { label: '.30-06 Springfield', slug: '30-06' },
  { label: '.270 Winchester', slug: '270-winchester' },
  { label: '12 Gauge', slug: '12-gauge' },
  { label: '20 Gauge', slug: '20-gauge' },
  { label: '.410 Bore', slug: '410-bore' },
]

const SLUGS: Record<string, string> = {}
for (const c of COMMON_CALIBERS) SLUGS[c.label.toLowerCase()] = c.slug

// Common things people type that differ from the canonical labels above.
const ALIASES: Record<string, string> = {
  '9mm': '9mm-luger',
  '9x19': '9mm-luger',
  '9mm parabellum': '9mm-luger',
  '380': '380-auto',
  '380 acp': '380-auto',
  '380 auto': '380-auto',
  '38 spl': '38-special',
  '357 mag': '357-magnum',
  '40 s&w': '40sw',
  '40sw': '40sw',
  '10mm': '10mm-auto',
  '44 mag': '44-magnum',
  '45 acp': '45acp',
  '45 auto': '45acp',
  '22 lr': '22lr',
  '22': '22lr',
  '223': '223-remington',
  '223 rem': '223-remington',
  '5.56': '5.56x45mm-nato',
  '5.56 nato': '5.56x45mm-nato',
  '5.56x45': '5.56x45mm-nato',
  '308': '308-winchester',
  '7.62x51': '308-winchester',
  '7.62 nato': '308-winchester',
  '7.62x39': '7.62x39mm',
  '300 blk': '300aac-blackout',
  '300 blackout': '300aac-blackout',
  '300 aac blackout': '300aac-blackout',
  '6.5 creedmoor': '6.5mm-creedmoor',
  '6.5mm creedmoor': '6.5mm-creedmoor',
  '30-06': '30-06',
  '270 win': '270-winchester',
  '12ga': '12-gauge',
  '12 ga': '12-gauge',
  '20ga': '20-gauge',
  '20 ga': '20-gauge',
  '350 legend': '350-legend',
  '.350 legend': '350-legend',
  '350': '350-legend',
  '410': '410-bore',
  '.410': '410-bore',
  '410 bore': '410-bore',
  '410ga': '410-bore',
}

const normalize = (caliber: string) => caliber.trim().toLowerCase().replace(/\s+/g, ' ')

/** ammoseekSlug returns AmmoSeek's slug for a caliber, or null if unrecognized. */
export function ammoseekSlug(caliber: string): string | null {
  const key = normalize(caliber)
  return SLUGS[key] ?? ALIASES[key] ?? null
}

/** ammoseekURL returns a deep link, or null when the caliber isn't recognized
 *  (so we never render a dead link). */
export function ammoseekURL(caliber: string): string | null {
  const slug = ammoseekSlug(caliber)
  return slug ? `https://ammoseek.com/ammo/${slug}` : null
}
