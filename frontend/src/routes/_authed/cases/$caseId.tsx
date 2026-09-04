import { createFileRoute } from '@tanstack/react-router'
import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { caseApi } from '@/lib/api-services'
import { queryKeys } from '@/lib/queryKeys'
import { EvidenceTab } from '@/components/EvidenceTab'
import { MapTab } from '@/components/MapTab'
import { TimelineTab } from '@/components/TimelineTab'
import { EntitiesTab } from '@/components/EntitiesTab'
import { CommentsTab } from '@/components/CommentsTab'
import { MembersTab } from '@/components/MembersTab'
import { ExportTab } from '@/components/ExportTab'

export const Route = createFileRoute('/_authed/cases/$caseId')({
  component: CaseDetailPage,
})

type Tab = 'evidence' | 'map' | 'timeline' | 'entities' | 'comments' | 'members' | 'export'

const TABS: { key: Tab; label: string }[] = [
  { key: 'evidence', label: 'Evidence' },
  { key: 'map', label: 'Map' },
  { key: 'timeline', label: 'Timeline' },
  { key: 'entities', label: 'Entities' },
  { key: 'comments', label: 'Comments' },
  { key: 'members', label: 'Members' },
  { key: 'export', label: 'Export' },
]

function CaseDetailPage() {
  const { caseId } = Route.useParams()
  const id = Number(caseId)
  const [tab, setTab] = useState<Tab>('evidence')

  const { data: caseData, isLoading } = useQuery({
    queryKey: queryKeys.case(id),
    queryFn: () => caseApi.get(id),
  })

  if (isLoading) return <div className="text-slate-500">Loading case...</div>
  if (!caseData) return <div className="text-red-500">Case not found.</div>

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="bg-white p-5 rounded-lg shadow">
        <div className="flex items-center justify-between">
          <h1 className="text-2xl font-bold">{caseData.title}</h1>
          <span className={`text-xs px-2 py-1 rounded ${caseData.status === 'open' ? 'bg-green-100 text-green-700' : 'bg-slate-100 text-slate-600'}`}>
            {caseData.status}
          </span>
        </div>
        {caseData.description && <p className="text-slate-600 mt-2">{caseData.description}</p>}
        <div className="text-xs text-slate-400 mt-2">
          Owner: {caseData.owner_name} · {caseData.visibility} · Created {new Date(caseData.created_at).toLocaleDateString()}
        </div>
        {caseData.tags.length > 0 && (
          <div className="flex gap-1 mt-2">
            {caseData.tags.map((t) => <span key={t} className="text-xs bg-brand-50 text-brand-700 px-2 py-0.5 rounded">#{t}</span>)}
          </div>
        )}
      </div>

      {/* Tabs */}
      <div className="border-b">
        <nav className="flex gap-1">
          {TABS.map((t) => (
            <button
              key={t.key}
              onClick={() => setTab(t.key)}
              className={`px-4 py-2 text-sm font-medium border-b-2 ${tab === t.key ? 'border-brand-600 text-brand-600' : 'border-transparent text-slate-500 hover:text-slate-700'}`}
            >
              {t.label}
            </button>
          ))}
        </nav>
      </div>

      {/* Tab content */}
      <div>
        {tab === 'evidence' && <EvidenceTab caseId={id} />}
        {tab === 'map' && <MapTab caseId={id} />}
        {tab === 'timeline' && <TimelineTab caseId={id} />}
        {tab === 'entities' && <EntitiesTab caseId={id} />}
        {tab === 'comments' && <CommentsTab caseId={id} />}
        {tab === 'members' && <MembersTab caseId={id} />}
        {tab === 'export' && <ExportTab caseId={id} />}
      </div>
    </div>
  )
}
