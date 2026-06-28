import { useEffect, useState } from 'react'
import { linkAmmo, listFirearmAmmo, listItems, unlinkAmmo } from '../api'
import type { Ammo, AmmoLink, Item } from '../types'
import RelatedRow from './RelatedRow'
import AttachControl from './AttachControl'

// AmmoLinks manages the ammo a firearm runs (many-to-many), with an optional
// per-link note like "zeroed / preferred load".
export default function AmmoLinks({ firearmId }: { firearmId: number }) {
  const [links, setLinks] = useState<AmmoLink[]>([])
  const [allAmmo, setAllAmmo] = useState<Ammo[]>([])

  const reload = () => {
    listFirearmAmmo(firearmId).then(setLinks).catch(() => {})
    listItems<Ammo>('ammo').then(setAllAmmo).catch(() => {})
  }
  useEffect(reload, [firearmId])

  const linkedIds = new Set(links.map((l) => l.ammo.id))
  const available = allAmmo.filter((a) => !linkedIds.has(a.id))

  return (
    <div>
      <h3 className="mb-2 text-sm font-medium text-dracula-comment">Ammo this runs</h3>
      {links.length === 0 ? (
        <p className="mb-3 text-sm text-dracula-comment">No ammo linked.</p>
      ) : (
        <ul className="mb-3 space-y-2">
          {links.map((l) => (
            <RelatedRow
              key={l.ammo.id}
              resource="ammo"
              item={l.ammo as Item}
              note={l.note || undefined}
              action={{ label: 'Remove', onClick: async () => { await unlinkAmmo(firearmId, l.ammo.id); reload() } }}
            />
          ))}
        </ul>
      )}
      <AttachControl
        resource="ammo"
        candidates={available as Item[]}
        placeholder="Add ammo…"
        withNote
        onAttach={async (ammoId, note) => { await linkAmmo(firearmId, ammoId, note); reload() }}
      />
    </div>
  )
}
