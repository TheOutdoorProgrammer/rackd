import { useEffect, useState } from 'react'
import { getItem } from '../api'
import type { Firearm, Item } from '../types'
import RelatedRow from './RelatedRow'

// FirearmRefCard resolves an accessory's firearmId into a proper related-item
// card (thumbnail + name + link) instead of a bare "#3". Editing the link still
// happens on the accessory's edit form.
export default function FirearmRefCard({ firearmId }: { firearmId: number }) {
  const [firearm, setFirearm] = useState<Firearm | null>(null)
  const [missing, setMissing] = useState(false)

  useEffect(() => {
    getItem<Firearm>('firearms', firearmId)
      .then(setFirearm)
      .catch(() => setMissing(true))
  }, [firearmId])

  return (
    <div>
      <h3 className="mb-2 text-sm font-medium text-dracula-comment">On this firearm</h3>
      {firearm ? (
        <ul className="space-y-2">
          <RelatedRow resource="firearms" item={firearm as Item} />
        </ul>
      ) : (
        <p className="text-sm text-dracula-comment">{missing ? 'Linked firearm not found.' : 'Loading…'}</p>
      )}
    </div>
  )
}
