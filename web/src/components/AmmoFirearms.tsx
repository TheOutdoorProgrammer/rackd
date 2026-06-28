import { useEffect, useState } from 'react'
import { linkAmmo, listFirearmsForAmmo, listItems, unlinkAmmo } from '../api'
import type { Firearm, Item } from '../types'
import RelatedRow from './RelatedRow'
import AttachControl from './AttachControl'

// AmmoFirearms is the reverse of a firearm's "Ammo this runs" — it shows and
// edits which guns a single ammo line feeds. One ammo can serve many guns.
export default function AmmoFirearms({ ammoId }: { ammoId: number }) {
  const [linked, setLinked] = useState<Firearm[]>([])
  const [all, setAll] = useState<Firearm[]>([])

  const reload = () => {
    listFirearmsForAmmo(ammoId).then(setLinked).catch(() => {})
    listItems<Firearm>('firearms').then(setAll).catch(() => {})
  }
  useEffect(reload, [ammoId])

  const linkedIds = new Set(linked.map((f) => f.id))
  const available = all.filter((f) => !linkedIds.has(f.id))

  return (
    <div>
      <h3 className="mb-2 text-sm font-medium text-dracula-comment">Guns that run this ammo</h3>
      {linked.length === 0 ? (
        <p className="mb-3 text-sm text-dracula-comment">Not linked to any firearm.</p>
      ) : (
        <ul className="mb-3 space-y-2">
          {linked.map((f) => (
            <RelatedRow
              key={f.id}
              resource="firearms"
              item={f as Item}
              action={{ label: 'Unlink', onClick: async () => { await unlinkAmmo(f.id, ammoId); reload() } }}
            />
          ))}
        </ul>
      )}
      <AttachControl
        resource="firearms"
        candidates={available as Item[]}
        placeholder="Link to a gun…"
        onAttach={async (firearmId) => { await linkAmmo(firearmId, ammoId, ''); reload() }}
      />
    </div>
  )
}
