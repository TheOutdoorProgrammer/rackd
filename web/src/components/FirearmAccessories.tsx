import { useEffect, useState } from 'react'
import { listAccessoriesForFirearm, listItems, updateItem } from '../api'
import type { Accessory, Item } from '../types'
import RelatedRow from './RelatedRow'
import AttachControl from './AttachControl'

// FirearmAccessories shows (and edits) the accessories mounted on a firearm.
// Each accessory belongs to at most one gun, so only unattached ones can be added.
export default function FirearmAccessories({ firearmId }: { firearmId: number }) {
  const [linked, setLinked] = useState<Accessory[]>([])
  const [all, setAll] = useState<Accessory[]>([])

  const reload = () => {
    listAccessoriesForFirearm(firearmId).then(setLinked).catch(() => {})
    listItems<Accessory>('accessories').then(setAll).catch(() => {})
  }
  useEffect(reload, [firearmId])

  const available = all.filter((a) => a.firearmId == null)

  return (
    <div>
      <h3 className="mb-2 text-sm font-medium text-dracula-comment">Accessories on this firearm</h3>
      {linked.length === 0 ? (
        <p className="mb-3 text-sm text-dracula-comment">No accessories attached.</p>
      ) : (
        <ul className="mb-3 space-y-2">
          {linked.map((a) => (
            <RelatedRow
              key={a.id}
              resource="accessories"
              item={a as Item}
              action={{ label: 'Detach', onClick: async () => { await updateItem('accessories', a.id, { ...a, firearmId: null }); reload() } }}
            />
          ))}
        </ul>
      )}
      <AttachControl
        resource="accessories"
        candidates={available as Item[]}
        placeholder="Attach accessory…"
        onAttach={async (accId) => {
          const acc = all.find((x) => x.id === accId)
          if (!acc) return
          await updateItem('accessories', accId, { ...acc, firearmId })
          reload()
        }}
      />
    </div>
  )
}
