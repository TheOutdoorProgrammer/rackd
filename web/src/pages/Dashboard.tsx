import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { getSummary } from '../api'
import { RESOURCE_KEYS, RESOURCES } from '../resources'
import { money } from '../format'
import type { Summary } from '../types'

export default function Dashboard() {
  const [sum, setSum] = useState<Summary | null>(null)
  useEffect(() => {
    getSummary().then(setSum).catch(() => {})
  }, [])

  return (
    <div>
      <div className="mb-5 rounded-2xl border border-dracula-current bg-dracula-current/30 p-5">
        <div className="text-sm text-dracula-comment">Estimated collection value</div>
        <div className="text-3xl font-semibold text-dracula-green">{sum ? money(sum.totalValueCents) : '—'}</div>
      </div>
      <div className="grid grid-cols-2 gap-3">
        {RESOURCE_KEYS.map((k) => (
          <Link
            key={k}
            to={`/${k}`}
            className="rounded-2xl border border-dracula-current p-4 transition hover:border-dracula-purple"
          >
            <div className="text-3xl">{RESOURCES[k].emoji}</div>
            <div className="mt-2 text-lg font-medium text-dracula-fg">{RESOURCES[k].label}</div>
            <div className="text-sm text-dracula-comment">{sum ? (sum.counts[k] ?? 0) : '—'} items</div>
            {k === 'ammo' && sum && sum.lowStockAmmo > 0 && (
              <div className="mt-1 text-xs font-medium text-dracula-red">{sum.lowStockAmmo} low on stock</div>
            )}
          </Link>
        ))}
      </div>
    </div>
  )
}
