import { useState } from 'react'
import { specsPage, specsSearch } from '../api'
import type { SpecPage, SpecSearchResult } from '../types'

// SpecLookup lets you search Wikipedia for a firearm, review its (community-
// sourced, sometimes messy) spec sheet, and optionally fill make/model/caliber
// into the form. It deliberately never auto-fills — you review first.
export default function SpecLookup({ onFill }: { onFill: (fields: Record<string, string>) => void }) {
  const [open, setOpen] = useState(false)
  const [q, setQ] = useState('')
  const [results, setResults] = useState<SpecSearchResult[] | null>(null)
  const [page, setPage] = useState<SpecPage | null>(null)
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState<string | null>(null)

  const runSearch = async () => {
    if (!q.trim()) return
    setBusy(true)
    setErr(null)
    setPage(null)
    try {
      setResults(await specsSearch(q))
    } catch {
      setErr('Search failed — Wikipedia may be unreachable')
    } finally {
      setBusy(false)
    }
  }

  const pick = async (title: string) => {
    setBusy(true)
    setErr(null)
    try {
      setPage(await specsPage(title))
    } catch {
      setErr('Could not load specs for that page')
    } finally {
      setBusy(false)
    }
  }

  if (!open) {
    return (
      <button
        type="button"
        onClick={() => setOpen(true)}
        className="w-full rounded-lg border border-dashed border-dracula-cyan/50 py-2 text-sm text-dracula-cyan"
      >
        🔎 Look up specs (Wikipedia)
      </button>
    )
  }

  return (
    <div className="rounded-xl border border-dracula-current p-3">
      <div className="mb-2 flex items-center justify-between">
        <span className="text-sm font-medium text-dracula-cyan">Spec lookup</span>
        <button type="button" onClick={() => setOpen(false)} className="text-xs text-dracula-comment">
          close
        </button>
      </div>

      <div className="flex gap-2">
        <input
          value={q}
          onChange={(e) => setQ(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === 'Enter') {
              e.preventDefault()
              void runSearch()
            }
          }}
          placeholder="e.g. Glock 19, AK-47, M1911"
          className="flex-1 rounded-lg border border-dracula-current bg-dracula-current/40 px-2 py-1.5 text-sm text-dracula-fg outline-none focus:border-dracula-cyan"
        />
        <button
          type="button"
          onClick={() => void runSearch()}
          disabled={busy}
          className="rounded-lg bg-dracula-cyan px-3 py-1.5 text-sm font-medium text-dracula-bg disabled:opacity-50"
        >
          {busy ? '…' : 'Search'}
        </button>
      </div>

      {err && <p className="mt-2 text-sm text-dracula-red">{err}</p>}

      {page ? (
        <div className="mt-3">
          <div className="mb-2 flex items-center justify-between">
            <a href={page.url} target="_blank" rel="noopener noreferrer" className="text-sm text-dracula-purple">
              {page.title} ↗
            </a>
            <button type="button" onClick={() => setPage(null)} className="text-xs text-dracula-comment">
              back
            </button>
          </div>
          <dl className="divide-y divide-dracula-current rounded-lg border border-dracula-current text-sm">
            {(page.specs ?? []).length === 0 ? (
              <div className="px-3 py-2 text-dracula-comment">No infobox specs found for this page.</div>
            ) : (
              (page.specs ?? []).map((s) => (
                <div key={s.label} className="flex justify-between gap-3 px-3 py-1.5">
                  <dt className="text-dracula-comment">{s.label}</dt>
                  <dd className="text-right text-dracula-fg">{s.value}</dd>
                </div>
              ))
            )}
          </dl>
          <p className="mt-2 text-xs text-dracula-comment">
            ⚠️ Community data from Wikipedia — review before filling. Series pages (e.g. “Glock”) can be mushy.
          </p>
          <button
            type="button"
            onClick={() => {
              onFill(page.fill ?? {})
              setOpen(false)
            }}
            className="mt-2 w-full rounded-lg bg-dracula-green/90 py-2 text-sm font-medium text-dracula-bg"
          >
            Fill make / model / caliber
          </button>
        </div>
      ) : results ? (
        <ul className="mt-3 space-y-1">
          {results.length === 0 && <li className="text-sm text-dracula-comment">No matches.</li>}
          {results.map((r) => (
            <li key={r.title}>
              <button
                type="button"
                onClick={() => void pick(r.title)}
                className="block w-full rounded-lg border border-dracula-current px-3 py-2 text-left text-sm hover:border-dracula-cyan"
              >
                <span className="text-dracula-fg">{r.title}</span>
                {r.description && <span className="ml-1 text-dracula-comment">— {r.description.slice(0, 80)}</span>}
              </button>
            </li>
          ))}
        </ul>
      ) : null}
    </div>
  )
}
