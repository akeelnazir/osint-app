import { createFileRoute, Outlet, Link, useNavigate, redirect } from '@tanstack/react-router'
import { useAuth } from '@/lib/auth'

export const Route = createFileRoute('/_authed')({
  beforeLoad: () => {
    const token = typeof window !== 'undefined' ? localStorage.getItem('access_token') : null
    if (!token) {
      throw redirect({ to: '/login' })
    }
  },
  component: AuthedLayout,
})

function AuthedLayout() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()

  return (
    <div className="min-h-screen flex flex-col">
      <header className="bg-white border-b shadow-sm">
        <div className="max-w-7xl mx-auto px-4 py-3 flex items-center justify-between">
          <Link to="/" className="text-xl font-bold text-brand-600">OSINT Research</Link>
          <nav className="flex items-center gap-4 text-sm">
            <Link to="/" className="hover:text-brand-600" activeProps={{ className: 'text-brand-600 font-medium' }}>
              Dashboard
            </Link>
            <Link to="/cases" className="hover:text-brand-600" activeProps={{ className: 'text-brand-600 font-medium' }}>
              Cases
            </Link>
            <Link to="/search" className="hover:text-brand-600" activeProps={{ className: 'text-brand-600 font-medium' }}>
              Search
            </Link>
            {user && (
              <div className="flex items-center gap-3">
                <span className="text-slate-600">{user.username} ({user.role})</span>
                <button
                  onClick={() => { logout(); navigate({ to: '/login' }) }}
                  className="text-slate-500 hover:text-red-600"
                >
                  Logout
                </button>
              </div>
            )}
          </nav>
        </div>
      </header>
      <main className="flex-1 max-w-7xl mx-auto w-full px-4 py-6">
        <Outlet />
      </main>
    </div>
  )
}
