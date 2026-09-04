import { createFileRoute } from '@tanstack/react-router'
import { useQuery } from '@tanstack/react-query'
import { dashboardApi } from '@/lib/api-services'
import { queryKeys } from '@/lib/queryKeys'
import { Link } from '@tanstack/react-router'

export const Route = createFileRoute('/_authed/')({
  component: DashboardPage,
})

function DashboardPage() {
  const { data, isLoading } = useQuery({
    queryKey: queryKeys.dashboard,
    queryFn: dashboardApi.stats,
  })

  if (isLoading) return <div className="text-center py-8 text-slate-500">Loading dashboard...</div>
  if (!data) return <div className="text-center py-8 text-red-500">Failed to load dashboard.</div>

  const { stats, recent_cases, activity } = data

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Dashboard</h1>
        <Link to="/cases" className="bg-brand-600 text-white px-4 py-2 rounded hover:bg-brand-700">
          New Case
        </Link>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <StatCard label="Cases" value={stats.cases} />
        <StatCard label="Evidence" value={stats.evidence} />
        <StatCard label="Entities" value={stats.entities} />
        <StatCard label="Users" value={stats.users} />
      </div>

      <div className="grid md:grid-cols-2 gap-6">
        {/* Recent cases */}
        <div className="bg-white rounded-lg shadow p-5">
          <h2 className="font-semibold mb-3">Recent Cases</h2>
          {recent_cases.length === 0 ? (
            <p className="text-slate-500 text-sm">No cases yet.</p>
          ) : (
            <ul className="space-y-2">
              {recent_cases.map((c) => (
                <li key={c.id}>
                  <Link to="/cases/$caseId" params={{ caseId: String(c.id) }} className="block p-2 rounded hover:bg-slate-50">
                    <div className="font-medium">{c.title}</div>
                    <div className="text-xs text-slate-500">
                      {c.status} · updated {new Date(c.updated_at).toLocaleDateString()}
                    </div>
                  </Link>
                </li>
              ))}
            </ul>
          )}
        </div>

        {/* Activity feed */}
        <div className="bg-white rounded-lg shadow p-5">
          <h2 className="font-semibold mb-3">Activity Feed</h2>
          {activity.length === 0 ? (
            <p className="text-slate-500 text-sm">No recent activity.</p>
          ) : (
            <ul className="space-y-2">
              {activity.map((a) => (
                <li key={a.id} className="text-sm">
                  <span className="font-medium">{a.username}</span>{' '}
                  <span className="text-slate-600">{a.action}</span>{' '}
                  <span className="text-slate-400">{new Date(a.created_at).toLocaleString()}</span>
                </li>
              ))}
            </ul>
          )}
        </div>
      </div>
    </div>
  )
}

function StatCard({ label, value }: { label: string; value: number }) {
  return (
    <div className="bg-white rounded-lg shadow p-5">
      <div className="text-3xl font-bold text-brand-600">{value}</div>
      <div className="text-sm text-slate-500 mt-1">{label}</div>
    </div>
  )
}
