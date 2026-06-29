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
  // reauth is set when an authenticated request 401s while the app is open — the
  // session expired or the vault re-locked. We overlay the PIN screen instead of
  // tearing the app down, so whatever form you were filling out survives the
  // re-unlock (no plaintext is ever persisted to do this — the data just stays in
  // memory behind the overlay).
  const [reauth, setReauth] = useState(false)
  const refresh = () => getStatus().then(setStatus).catch(() => setStatus(null))

  useEffect(() => {
    setUnauthorizedHandler(() => setReauth(true))
    void refresh()
  }, [])

  if (status === null) {
    return <div className="flex min-h-dvh items-center justify-center text-dracula-comment">Loading…</div>
  }

  // First run, or a fresh load that was never unlocked: take the whole screen.
  if (!status.unlocked) {
    return <UnlockScreen status={status} onUnlocked={() => { setReauth(false); refresh() }} />
  }

  return (
    <BrowserRouter>
      <AppShell onLock={async () => { setReauth(false); await lockVault(); await refresh() }}>
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
      {/* Session expired mid-use: re-unlock over the live app so the form you were
          filling is still there when you come back — just hit Save again. */}
      {reauth && (
        <div className="fixed inset-0 z-50 bg-dracula-bg">
          <UnlockScreen
            status={{ initialized: true, unlocked: false }}
            onUnlocked={() => { setReauth(false); void refresh() }}
          />
        </div>
      )}
    </BrowserRouter>
  )
}
