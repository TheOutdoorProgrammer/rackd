import { useState } from 'react'
import { adjustAmmo } from '../api'
import type { Ammo } from '../types'

// AmmoStock shows rounds-on-hand with quick "use" (subtract) and "refill" (add)
// buttons, and flags the line when it drops to its low-stock threshold.
export default function AmmoStock({ ammo, onChange }: { ammo: Ammo; onChange: (a: Ammo) => void }) {
  const [amt, setAmt] = useState(20)
  const [busy, setBusy] = useState(false)

  const low = ammo.lowStockThreshold > 0 && ammo.quantityOnHand <= ammo.lowStockThreshold

  const apply = async (delta: number) => {
    if (!delta) return
    setBusy(true)
    try {
      onChange(await adjustAmmo(ammo.id, delta))
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="rounded-2xl border border-dracula-current p-4">
      <div className="mb-3 flex items-end justify-between gap-3">
        <div>
          <div className="text-sm text-dracula-comment">Rounds on hand</div>
          <div className={`text-3xl font-semibold ${low ? 'text-dracula-red' : 'text-dracula-green'}`}>
            {ammo.quantityOnHand}
          </div>
        </div>
        {low && (
          <span className="shrink-0 rounded-full bg-dracula-red/20 px-2.5 py-1 text-xs font-medium text-dracula-red">
            Low stock{ammo.lowStockThreshold ? ` (≤ ${ammo.lowStockThreshold})` : ''}
          </span>
        )}
      </div>
      <div className="flex flex-wrap items-center gap-2">
        <input
          type="number"
          min={1}
          value={amt}
          onChange={(e) => setAmt(Math.max(0, Number(e.target.value)))}
          className="w-20 rounded-lg border border-dracula-current bg-dracula-current/40 px-2 py-1.5 text-sm text-dracula-fg"
        />
        <button
          onClick={() => apply(-amt)}
          disabled={busy || amt <= 0}
          className="rounded-lg border border-dracula-red/50 px-3 py-1.5 text-sm text-dracula-red disabled:opacity-50"
        >
          − Use
        </button>
        <button
          onClick={() => apply(amt)}
          disabled={busy || amt <= 0}
          className="rounded-lg bg-dracula-green px-3 py-1.5 text-sm font-medium text-dracula-bg disabled:opacity-50"
        >
          + Add stock
        </button>
      </div>
    </div>
  )
}
