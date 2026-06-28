import { useEffect, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { search } from '../api'
import { RESOURCES } from '../resources'
import type { SearchResults } from '../types'

export default function SearchPage() {
  const [params] = useSearchParams()
  const q = params.get('q') ?? ''
  const [res, setRes] = useState<SearchResults | null>(null)

  useEffect(() => {
    search(q).then(setRes).catch(() => setRes(null))
  }, [q])

  const groups: { key: string; items: any[] }[] = res
    ? [
        { key: 'firearms', items: res.firearms },
        { key: 'ammo', items: res.ammo },
        { key: 'knives', items: res.knives },
        { key: 'accessories', items: res.accessories },
      ]
    : []
  const total = groups.reduce((n, g) => n + g.items.length, 0)

  return (
    <div className="space-y-4">
      <h2 className="text-xl font-semibold text-dracula-fg">
        Search {q && <span className="text-dracula-comment">“{q}”</span>}
      </h2>
      {res === null ? (
        <p className="text-dracula-comment">Loading…</p>
      ) : total === 0 ? (
        <p className="text-dracula-comment">No matches.</p>
      ) : (
        groups
          .filter((g) => g.items.length > 0)
          .map((g) => {
            const cfg = RESOURCES[g.key]
            return (
              <div key={g.key}>
                <h3 className="mb-1 text-sm font-medium text-dracula-comment">{cfg.emoji} {cfg.label}</h3>
                <ul className="space-y-1">
                  {g.items.map((it) => (
                    <li key={it.id}>
                      <Link to={`/${g.key}/${it.id}`} className="block rounded-lg border border-dracula-current px-3 py-2 hover:border-dracula-purple">
                        <span className="text-dracula-fg">{cfg.title(it)}</span>
                        <span className="ml-2 text-sm text-dracula-comment">{cfg.subtitle(it)}</span>
                      </Link>
                    </li>
                  ))}
                </ul>
              </div>
            )
          })
      )}
    </div>
  )
}
