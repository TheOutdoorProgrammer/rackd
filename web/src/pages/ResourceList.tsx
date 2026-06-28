import { useEffect, useState } from 'react'
import { Link, Navigate, useParams } from 'react-router-dom'
import { listItems } from '../api'
import { RESOURCES } from '../resources'
import CoverThumb from '../components/CoverThumb'
import type { Item } from '../types'

export default function ResourceList() {
  const { resource } = useParams()
  const cfg = resource ? RESOURCES[resource] : undefined
  const [items, setItems] = useState<Item[] | null>(null)

  useEffect(() => {
    if (resource && RESOURCES[resource]) {
      listItems<Item>(resource).then(setItems).catch(() => setItems([]))
    }
  }, [resource])

  if (!cfg) return <Navigate to="/" replace />

  return (
    <div>
      <div className="mb-4 flex items-center justify-between">
        <h2 className="text-xl font-semibold text-dracula-fg">{cfg.label}</h2>
        <Link to={`/${cfg.key}/new`} className="rounded-lg bg-dracula-purple px-3 py-1.5 text-sm font-medium text-dracula-bg">
          + Add
        </Link>
      </div>
      {items === null ? (
        <p className="text-dracula-comment">Loading…</p>
      ) : items.length === 0 ? (
        <div className="rounded-2xl border border-dashed border-dracula-current p-8 text-center text-dracula-comment">
          No {cfg.label.toLowerCase()} yet. <Link to={`/${cfg.key}/new`} className="text-dracula-purple">Add one</Link>.
        </div>
      ) : (
        <ul className="space-y-2">
          {items.map((it) => (
            <li key={it.id}>
              <Link
                to={`/${cfg.key}/${it.id}`}
                className="flex items-center gap-3 rounded-xl border border-dracula-current p-3 transition hover:border-dracula-purple"
              >
                <CoverThumb owner={cfg.key} id={it.id} emoji={cfg.emoji} />
                <div className="min-w-0 flex-1">
                  <div className="flex items-center gap-2">
                    <span className="truncate font-medium text-dracula-fg">{cfg.title(it)}</span>
                    {cfg.key === 'ammo' && it.lowStockThreshold > 0 && it.quantityOnHand <= it.lowStockThreshold && (
                      <span className="shrink-0 rounded-full bg-dracula-red/20 px-2 py-0.5 text-xs font-medium text-dracula-red">Low</span>
                    )}
                  </div>
                  <div className="truncate text-sm text-dracula-comment">{cfg.subtitle(it) || '—'}</div>
                </div>
              </Link>
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
