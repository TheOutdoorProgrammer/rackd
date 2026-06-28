import { useEffect, useState } from 'react'
import { specCacheStats, specClearCache } from '../api'

export default function Settings() {
  const [count, setCount] = useState<number | null>(null)
  const [busy, setBusy] = useState(false)

  const refresh = () => specCacheStats().then((s) => setCount(s.count)).catch(() => setCount(null))
  useEffect(() => { refresh() }, [])

  const clear = async () => {
    setBusy(true)
    try {
      await specClearCache()
      await refresh()
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="space-y-4">
      <h2 className="text-xl font-semibold text-dracula-fg">Settings</h2>
      <div className="rounded-2xl border border-dracula-current p-4">
        <div className="mb-1 font-medium text-dracula-fg">Spec-lookup cache</div>
        <p className="text-sm text-dracula-comment">
          Wikipedia / DBpedia responses are cached indefinitely so repeat lookups are instant and don't re-hit the
          network. Clear it to force fresh data.
        </p>
        <div className="mt-3 flex items-center justify-between">
          <span className="text-sm text-dracula-comment">{count === null ? '…' : `${count} cached responses`}</span>
          <button
            onClick={clear}
            disabled={busy}
            className="rounded-lg border border-dracula-current px-3 py-1.5 text-sm text-dracula-fg disabled:opacity-50"
          >
            {busy ? 'Clearing…' : 'Clear cache'}
          </button>
        </div>
      </div>
    </div>
  )
}
