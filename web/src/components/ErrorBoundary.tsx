import { Component, type ErrorInfo, type ReactNode } from 'react'

// ErrorBoundary keeps a render-time exception in one view from blanking the
// whole app. It shows a recoverable message instead.
export default class ErrorBoundary extends Component<{ children: ReactNode }, { error: Error | null }> {
  state: { error: Error | null } = { error: null }

  static getDerivedStateFromError(error: Error) {
    return { error }
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error('render error:', error, info)
  }

  render() {
    if (this.state.error) {
      return (
        <div className="flex min-h-[50vh] flex-col items-center justify-center gap-3 px-6 text-center">
          <p className="text-dracula-red">Something went wrong rendering this view.</p>
          <div className="flex gap-2">
            <button
              onClick={() => this.setState({ error: null })}
              className="rounded-lg border border-dracula-current px-3 py-1.5 text-sm text-dracula-fg"
            >
              Try again
            </button>
            <button
              onClick={() => window.location.reload()}
              className="rounded-lg bg-dracula-purple px-3 py-1.5 text-sm text-dracula-bg"
            >
              Reload
            </button>
          </div>
        </div>
      )
    }
    return this.props.children
  }
}
