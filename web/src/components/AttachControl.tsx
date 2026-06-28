import { useState } from 'react'
import { RESOURCES } from '../resources'
import type { Item } from '../types'

// AttachControl is the shared "link another item" picker used by every
// relationship section: a dropdown of candidates (already named via the same
// RESOURCES config), an optional note, and an Add button.
export default function AttachControl({
  resource,
  candidates,
  placeholder,
  withNote = false,
  onAttach,
}: {
  resource: string
  candidates: Item[]
  placeholder: string
  withNote?: boolean
  onAttach: (id: number, note: string) => void | Promise<void>
}) {
  const [sel, setSel] = useState('')
  const [note, setNote] = useState('')
  const cfg = RESOURCES[resource]
  if (candidates.length === 0) return null

  const add = async () => {
    if (!sel) return
    await onAttach(Number(sel), note)
    setSel('')
    setNote('')
  }

  return (
    <div className="mt-1 flex flex-wrap items-center gap-2">
      <select
        value={sel}
        onChange={(e) => setSel(e.target.value)}
        className="rounded-lg border border-dracula-current bg-dracula-current/40 px-2 py-1.5 text-sm text-dracula-fg"
      >
        <option value="">{placeholder}</option>
        {candidates.map((c) => (
          <option key={c.id} value={c.id}>{cfg.title(c)}</option>
        ))}
      </select>
      {withNote && (
        <input
          value={note}
          onChange={(e) => setNote(e.target.value)}
          placeholder="note (e.g. zeroed)"
          className="min-w-0 flex-1 rounded-lg border border-dracula-current bg-dracula-current/40 px-2 py-1.5 text-sm text-dracula-fg"
        />
      )}
      <button onClick={add} disabled={!sel} className="rounded-lg bg-dracula-purple px-3 py-1.5 text-sm text-dracula-bg disabled:opacity-50">
        Add
      </button>
    </div>
  )
}
