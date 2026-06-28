import { useEffect, useState } from 'react'
import { linkAmmo, listFirearmAmmo, listItems, unlinkAmmo } from '../api'
import type { Ammo, AmmoLink } from '../types'

export default function AmmoLinks({ firearmId }: { firearmId: number }) {
  const [links, setLinks] = useState<AmmoLink[]>([])
  const [allAmmo, setAllAmmo] = useState<Ammo[]>([])
  const [sel, setSel] = useState('')
  const [note, setNote] = useState('')

  const refresh = () => listFirearmAmmo(firearmId).then(setLinks).catch(() => {})
  useEffect(() => {
    refresh()
    listItems<Ammo>('ammo').then(setAllAmmo).catch(() => {})
  }, [firearmId])

  const add = async () => {
    if (!sel) return
    await linkAmmo(firearmId, Number(sel), note)
    setSel('')
    setNote('')
    await refresh()
  }
  const remove = async (ammoId: number) => {
    await unlinkAmmo(firearmId, ammoId)
    await refresh()
  }

  const linkedIds = new Set(links.map((l) => l.ammo.id))
  const available = allAmmo.filter((a) => !linkedIds.has(a.id))

  return (
    <div>
      <h3 className="mb-2 text-sm font-medium text-dracula-comment">Ammo this runs</h3>
      {links.length === 0 ? (
        <p className="text-sm text-dracula-comment">No ammo linked.</p>
      ) : (
        <ul className="mb-3 space-y-1">
          {links.map((l) => (
            <li key={l.ammo.id} className="flex items-center justify-between rounded-lg border border-dracula-current px-3 py-2">
              <span className="text-dracula-fg">
                {l.ammo.name || l.ammo.caliber}
                {l.note ? <span className="text-dracula-comment"> — {l.note}</span> : null}
              </span>
              <button onClick={() => remove(l.ammo.id)} className="text-sm text-dracula-red">Remove</button>
            </li>
          ))}
        </ul>
      )}
      {available.length > 0 && (
        <div className="flex flex-wrap items-center gap-2">
          <select
            value={sel}
            onChange={(e) => setSel(e.target.value)}
            className="rounded-lg border border-dracula-current bg-dracula-current/40 px-2 py-1.5 text-sm text-dracula-fg"
          >
            <option value="">Add ammo…</option>
            {available.map((a) => (
              <option key={a.id} value={a.id}>{a.name || a.caliber}</option>
            ))}
          </select>
          <input
            value={note}
            onChange={(e) => setNote(e.target.value)}
            placeholder="note (e.g. zeroed)"
            className="flex-1 rounded-lg border border-dracula-current bg-dracula-current/40 px-2 py-1.5 text-sm text-dracula-fg"
          />
          <button onClick={add} disabled={!sel} className="rounded-lg bg-dracula-purple px-3 py-1.5 text-sm text-dracula-bg disabled:opacity-50">
            Link
          </button>
        </div>
      )}
    </div>
  )
}
