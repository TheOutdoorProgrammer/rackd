import { useEffect, useState } from 'react'
import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import { getStatus, lockVault, setUnauthorizedHandler } from './api'
import type { Status } from './types'
import UnlockScreen from './components/UnlockScreen'
import AppShell from './components/AppShell'
import Dashboard from './pages/Dashboard'
import ResourceList from './pages/ResourceList'
import ResourceForm from './pages/ResourceForm'
import ResourceDetail from './pages/ResourceDetail'
import SearchPage from './pages/SearchPage'
import Settings from './pages/Settings'
import ErrorBoundary from './components/ErrorBoundary'

export default function App() {
  const [status, setStatus] = useState<Status | null>(null)
  const refresh = () => getStatus().then(setStatus).catch(() => setStatus(null))

  useEffect(() => {
    // A 401 mid-session means the vault locked (session expiry or a server
    // restart) — flip back to the PIN screen instead of leaving a broken page.
    setUnauthorizedHandler(() => setStatus((s) => (s ? { ...s, unlocked: false } : s)))
    void refresh()
  }, [])

  if (status === null) {
    return <div className="flex min-h-dvh items-center justify-center text-dracula-comment">Loading…</div>
  }
  if (!status.unlocked) {
    return <UnlockScreen status={status} onUnlocked={refresh} />
  }

  return (
    <BrowserRouter>
      <AppShell onLock={async () => { await lockVault(); await refresh() }}>
        <ErrorBoundary>
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="/search" element={<SearchPage />} />
          <Route path="/settings" element={<Settings />} />
          <Route path="/:resource" element={<ResourceList />} />
          <Route path="/:resource/new" element={<ResourceForm />} />
          <Route path="/:resource/:id" element={<ResourceDetail />} />
          <Route path="/:resource/:id/edit" element={<ResourceForm />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
        </ErrorBoundary>
      </AppShell>
    </BrowserRouter>
  )
}
