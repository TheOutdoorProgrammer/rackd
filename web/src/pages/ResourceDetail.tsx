import { useEffect, useState } from 'react'
import { Link, Navigate, useNavigate, useParams } from 'react-router-dom'
import { deleteItem, getItem } from '../api'
import { RESOURCES, type Field } from '../resources'
import { money, titleCase } from '../format'
import PhotoManager from '../components/PhotoManager'
import AmmoLinks from '../components/AmmoLinks'
import FirearmAccessories from '../components/FirearmAccessories'
import AmmoStock from '../components/AmmoStock'
import AmmoFirearms from '../components/AmmoFirearms'
import AccessoryFirearms from '../components/AccessoryFirearms'
import AmmoSeekButton from '../components/AmmoSeekButton'
import type { Ammo, Item } from '../types'

function formatValue(f: Field, v: any): string {
  if (v === null || v === undefined || v === '') return ''
  if (Array.isArray(v)) return v.length ? v.join(', ') : ''
  if (f.type === 'money') return v ? money(v) : ''
  if (f.type === 'number') return v ? String(v) : ''
  if (f.type === 'bool') return v ? 'Yes' : ''
  if (f.type === 'select') return titleCase(String(v))
  return String(v)
}

export default function ResourceDetail() {
  const { resource, id } = useParams()
  const cfg = resource ? RESOURCES[resource] : undefined
  const nav = useNavigate()
  const [item, setItem] = useState<Item | null>(null)
  const [missing, setMissing] = useState(false)

  useEffect(() => {
    if (resource && id && RESOURCES[resource]) {
      getItem<Item>(resource, Number(id)).then(setItem).catch(() => setMissing(true))
    }
  }, [resource, id])

  if (!cfg) return <Navigate to="/" replace />
  if (missing) return <p className="text-dracula-comment">Not found.</p>
  if (!item) return <p className="text-dracula-comment">Loading…</p>

  const onDelete = async () => {
    if (!confirm(`Delete this ${cfg.singular.toLowerCase()}?`)) return
    await deleteItem(resource!, Number(id))
    nav(`/${resource}`)
  }

  // firearmRef → shown as a related-item card; ammo's rounds → shown by AmmoStock.
  const fields = cfg.fields.filter(
    (f) => f.type !== 'firearmRef' && !(cfg.key === 'ammo' && f.name === 'quantityOnHand'),
  )

  return (
    <div className="space-y-5">
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <h2 className="text-2xl font-semibold text-dracula-fg">{cfg.title(item)}</h2>
          <p className="text-dracula-comment">{cfg.subtitle(item) || ' '}</p>
        </div>
        <div className="flex shrink-0 gap-2">
          <Link to={`/${resource}/${id}/edit`} className="rounded-lg border border-dracula-current px-3 py-1.5 text-sm text-dracula-fg">Edit</Link>
          <button onClick={onDelete} className="rounded-lg border border-dracula-red/50 px-3 py-1.5 text-sm text-dracula-red">Delete</button>
        </div>
      </div>

      <PhotoManager owner={cfg.key} id={Number(id)} />

      <dl className="divide-y divide-dracula-current rounded-2xl border border-dracula-current">
        {fields.map((f) => {
          const val = formatValue(f, item[f.name])
          if (!val) return null
          return (
            <div key={f.name} className="flex justify-between gap-4 px-4 py-2.5">
              <dt className="text-sm text-dracula-comment">{f.label}</dt>
              <dd className="text-right text-dracula-fg">{val}</dd>
            </div>
          )
        })}
      </dl>

      {cfg.key === 'ammo' && <AmmoStock ammo={item as Ammo} onChange={(a) => setItem(a)} />}

      {/* Relationships — every related item rendered the same way (RelatedRow). */}
      {cfg.key === 'firearms' && (
        <>
          <AmmoLinks firearmId={Number(id)} />
          <FirearmAccessories firearmId={Number(id)} />
        </>
      )}
      {cfg.key === 'ammo' && <AmmoFirearms ammoId={Number(id)} />}
      {cfg.key === 'accessories' && <AccessoryFirearms accessoryId={Number(id)} quantity={Number(item.quantity) || 0} />}

      {/* External links always sit at the very bottom. */}
      {cfg.key === 'firearms' && item.caliber && (
        <AmmoSeekButton caliber={item.caliber} label={`Find ${item.caliber} ammo on AmmoSeek`} />
      )}
      {cfg.key === 'ammo' && (
        <div className="rounded-2xl border border-dracula-current p-4">
          {item.quantityOnHand > 0 && item.acquiredPriceCents > 0 && (
            <p className="mb-3 text-sm text-dracula-comment">
              Your cost:{' '}
              <span className="font-medium text-dracula-green">
                {money(Math.round(item.acquiredPriceCents / item.quantityOnHand))}/rd
              </span>
            </p>
          )}
          {item.caliber ? (
            <AmmoSeekButton caliber={item.caliber} label={`Check ${item.caliber} prices on AmmoSeek`} />
          ) : (
            <p className="text-sm text-dracula-comment">Add a caliber to check live prices on AmmoSeek.</p>
          )}
        </div>
      )}
    </div>
  )
}
