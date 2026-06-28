import { Link } from 'react-router-dom'
import CoverThumb from './CoverThumb'
import { RESOURCES } from '../resources'
import type { Item } from '../types'

// RelatedRow is the single, shared way a related item is shown anywhere in the
// app: cover thumbnail + name + subtitle, linking to that item's own page. Used
// by every relationship section so guns, ammo, knives, and accessories all read
// identically wherever they're referenced.
export default function RelatedRow({
  resource,
  item,
  note,
  action,
}: {
  resource: string
  item: Item
  note?: string
  action?: { label: string; onClick: () => void }
}) {
  const cfg = RESOURCES[resource]
  if (!cfg) return null
  return (
    <li className="flex items-center gap-3 rounded-xl border border-dracula-current p-2.5">
      <Link to={`/${resource}/${item.id}`} className="flex min-w-0 flex-1 items-center gap-3">
        <CoverThumb owner={resource} id={item.id} emoji={cfg.emoji} />
        <div className="min-w-0">
          <div className="truncate font-medium text-dracula-fg">{cfg.title(item)}</div>
          <div className="truncate text-sm text-dracula-comment">{note || cfg.subtitle(item) || '—'}</div>
        </div>
      </Link>
      {action && (
        <button onClick={action.onClick} className="shrink-0 text-sm text-dracula-red">
          {action.label}
        </button>
      )}
    </li>
  )
}
