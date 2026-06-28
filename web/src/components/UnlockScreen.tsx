import { useState, type FormEvent } from 'react'
import { setupVault, unlockVault } from '../api'
import type { Status } from '../types'

export default function UnlockScreen({
  status,
  onUnlocked,
}: {
  status: Status
  onUnlocked: () => void
}) {
  const isSetup = !status.initialized
  const [pin, setPin] = useState('')
  const [confirm, setConfirm] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [busy, setBusy] = useState(false)

  const submit = async (e: FormEvent) => {
    e.preventDefault()
    setError(null)
    if (pin.length !== 6) {
      setError('Enter a 6-digit PIN')
      return
    }
    if (isSetup && pin !== confirm) {
      setError('PINs do not match')
      return
    }
    setBusy(true)
    try {
      if (isSetup) await setupVault(pin)
      else await unlockVault(pin)
      onUnlocked()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Something went wrong')
      setPin('')
      setConfirm('')
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="flex min-h-dvh flex-col items-center justify-center px-6">
      <form onSubmit={submit} className="w-full max-w-sm">
        <h1 className="mb-1 text-center text-3xl font-semibold tracking-wide text-dracula-purple">rackd</h1>
        <p className="mb-8 text-center text-sm text-dracula-comment">
          {isSetup ? 'Create a 6-digit PIN to secure your vault' : 'Enter your PIN to unlock'}
        </p>

        <PinInput label={isSetup ? 'New PIN' : 'PIN'} value={pin} onChange={setPin} autoFocus />
        {isSetup && <PinInput label="Confirm PIN" value={confirm} onChange={setConfirm} />}

        {error && <p className="mt-3 text-center text-sm text-dracula-red">{error}</p>}

        <button
          type="submit"
          disabled={busy}
          className="mt-6 w-full rounded-xl bg-dracula-purple py-3 font-medium text-dracula-bg transition active:scale-[.99] disabled:opacity-50"
        >
          {busy ? '…' : isSetup ? 'Create vault' : 'Unlock'}
        </button>
      </form>
    </div>
  )
}

function PinInput({
  label,
  value,
  onChange,
  autoFocus,
}: {
  label: string
  value: string
  onChange: (v: string) => void
  autoFocus?: boolean
}) {
  return (
    <label className="mb-3 block">
      <span className="mb-1 block text-xs uppercase tracking-wider text-dracula-comment">{label}</span>
      <input
        type="password"
        inputMode="numeric"
        pattern="[0-9]*"
        autoComplete="off"
        autoFocus={autoFocus}
        maxLength={6}
        value={value}
        onChange={(e) => onChange(e.target.value.replace(/\D/g, '').slice(0, 6))}
        className="w-full rounded-xl border border-dracula-current bg-dracula-current/40 px-4 py-3 text-center text-2xl tracking-[0.5em] text-dracula-fg outline-none focus:border-dracula-purple"
      />
    </label>
  )
}
