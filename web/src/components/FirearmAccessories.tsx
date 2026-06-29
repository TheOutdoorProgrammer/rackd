import { useEffect, useState } from 'react'
import { linkAccessory, listAccessoriesForFirearm, listItems, unlinkAccessory } from '../api'
import type { Accessory, Item } from '../types'
import RelatedRow from './RelatedRow'
import AttachControl from './AttachControl'

// FirearmAccessories shows (and edits) the accessories mounted on a firearm.
// An accessory can be mounted on several guns up to its quantity, so candidates
// are simply those not already on this gun; the server enforces the per-accessory
// cap and rejects an over-assignment.
export default function FirearmAccessories({ firearmId }: { firearmId: number }) {
  const [linked, setLinked] = useState<Accessory[]>([])
  const [all, setAll] = useState<Accessory[]>([])
  const [err, setErr] = useState<string | null>(null)

  const reload = () => {
    listAccessoriesForFirearm(firearmId).then(setLinked).catch(() => {})
    listItems<Accessory>('accessories').then(setAll).catch(() => {})
  }
  useEffect(reload, [firearmId])

  const linkedIds = new Set(linked.map((a) => a.id))
  const available = all.filter((a) => !linkedIds.has(a.id))

  return (
    <div>
      <h3 className="mb-2 text-sm font-medium text-dracula-comment">Accessories on this firearm</h3>
      {err && <p className="mb-2 text-sm text-dracula-red">{err}</p>}
      {linked.length === 0 ? (
        <p className="mb-3 text-sm text-dracula-comment">No accessories attached.</p>
      ) : (
        <ul className="mb-3 space-y-2">
          {linked.map((a) => (
            <RelatedRow
              key={a.id}
              resource="accessories"
              item={a as Item}
              action={{ label: 'Detach', onClick: async () => { await unlinkAccessory(firearmId, a.id); reload() } }}
            />
          ))}
        </ul>
      )}
      <AttachControl
        resource="accessories"
        candidates={available as Item[]}
        placeholder="Attach accessory…"
        onAttach={async (accId) => {
          setErr(null)
          try {
            await linkAccessory(firearmId, accId)
            reload()
          } catch (e) {
            setErr(e instanceof Error ? e.message : 'Could not attach')
          }
        }}
      />
    </div>
  )
}
