import { useState, type FormEvent, type ReactNode } from 'react'
import { Link, NavLink, useNavigate } from 'react-router-dom'
import { RESOURCE_KEYS, RESOURCES } from '../resources'

export default function AppShell({ onLock, children }: { onLock: () => void; children: ReactNode }) {
  const nav = useNavigate()
  const [q, setQ] = useState('')

  const onSearch = (e: FormEvent) => {
    e.preventDefault()
    nav(`/search?q=${encodeURIComponent(q)}`)
  }

  return (
    <div className="min-h-dvh">
      <header className="sticky top-0 z-10 border-b border-dracula-current bg-dracula-bg/90 backdrop-blur">
        <div className="mx-auto flex max-w-3xl items-center gap-3 px-4 py-3">
          <Link to="/" className="text-xl font-semibold tracking-wide text-dracula-purple">Boating Accident</Link>
          <form onSubmit={onSearch} className="flex-1">
            <input
              value={q}
              onChange={(e) => setQ(e.target.value)}
              placeholder="Search…"
              className="w-full rounded-lg border border-dracula-current bg-dracula-current/40 px-3 py-1.5 text-sm text-dracula-fg outline-none focus:border-dracula-purple"
            />
          </form>
          <Link to="/settings" title="Settings" className="rounded-lg border border-dracula-current px-2 py-1.5 text-sm text-dracula-comment hover:text-dracula-fg">
            ⚙
          </Link>
          <button onClick={onLock} className="rounded-lg border border-dracula-current px-3 py-1.5 text-sm text-dracula-comment hover:text-dracula-fg">
            Lock
          </button>
        </div>
        <nav className="mx-auto flex max-w-3xl gap-1 overflow-x-auto px-2 pb-2 text-sm">
          {RESOURCE_KEYS.map((k) => (
            <NavLink
              key={k}
              to={`/${k}`}
              className={({ isActive }) =>
                `whitespace-nowrap rounded-lg px-3 py-1.5 ${isActive ? 'bg-dracula-current text-dracula-fg' : 'text-dracula-comment hover:text-dracula-fg'}`
              }
            >
              {RESOURCES[k].emoji} {RESOURCES[k].label}
            </NavLink>
          ))}
        </nav>
      </header>
      <main className="mx-auto max-w-3xl px-4 py-5">{children}</main>
    </div>
  )
}
