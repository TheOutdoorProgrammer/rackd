import { useEffect, useState } from 'react'
import { linkAccessory, listFirearmsForAccessory, listItems, unlinkAccessory } from '../api'
import type { Firearm, Item } from '../types'
import RelatedRow from './RelatedRow'
import AttachControl from './AttachControl'

// AccessoryFirearms shows and edits which guns an accessory is mounted on. An
// accessory can go on up to `quantity` distinct firearms (one physical unit each);
// once every unit is assigned, the picker hides until a slot is freed or the
// quantity is raised.
export default function AccessoryFirearms({ accessoryId, quantity }: { accessoryId: number; quantity: number }) {
  const [linked, setLinked] = useState<Firearm[]>([])
  const [all, setAll] = useState<Firearm[]>([])
  const [err, setErr] = useState<string | null>(null)

  const reload = () => {
    listFirearmsForAccessory(accessoryId).then(setLinked).catch(() => {})
    listItems<Firearm>('firearms').then(setAll).catch(() => {})
  }
  useEffect(reload, [accessoryId])

  const linkedIds = new Set(linked.map((f) => f.id))
  const available = all.filter((f) => !linkedIds.has(f.id))
  const cap = Math.max(1, quantity || 0)
  const atCap = linked.length >= cap

  return (
    <div>
      <h3 className="mb-2 text-sm font-medium text-dracula-comment">On firearms</h3>
      {err && <p className="mb-2 text-sm text-dracula-red">{err}</p>}
      {linked.length === 0 ? (
        <p className="mb-3 text-sm text-dracula-comment">Not mounted on any firearm.</p>
      ) : (
        <ul className="mb-3 space-y-2">
          {linked.map((f) => (
            <RelatedRow
              key={f.id}
              resource="firearms"
              item={f as Item}
              action={{ label: 'Detach', onClick: async () => { await unlinkAccessory(f.id, accessoryId); reload() } }}
            />
          ))}
        </ul>
      )}
      {atCap ? (
        <p className="text-sm text-dracula-comment">
          {cap === 1
            ? 'Raise the quantity to mount this on more than one gun.'
            : `All ${cap} assigned — detach one or raise the quantity to add more.`}
        </p>
      ) : (
        <AttachControl
          resource="firearms"
          candidates={available as Item[]}
          placeholder="Mount on a gun…"
          onAttach={async (firearmId) => {
            setErr(null)
            try {
              await linkAccessory(firearmId, accessoryId)
              reload()
            } catch (e) {
              setErr(e instanceof Error ? e.message : 'Could not mount')
            }
          }}
        />
      )}
    </div>
  )
}
