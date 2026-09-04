import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { evidenceApi } from '@/lib/api-services'
import { queryKeys } from '@/lib/queryKeys'

export function EntitiesTab({ caseId }: { caseId: number }) {
  const { data: evidence } = useQuery({
    queryKey: queryKeys.evidenceList(caseId, {}),
    queryFn: () => evidenceApi.list(caseId, { limit: 200 }),
  })
  const qc = useQueryClient()

  // Collect all entities across evidence.
  const allEntities = useQuery({
    queryKey: ['case-entities', caseId],
    queryFn: async () => {
      if (!evidence) return []
      const results = await Promise.all(
        evidence.data.map((e) => evidenceApi.listEntities(caseId, e.id).then((ents) => ({ evidenceId: e.id, evidenceTitle: e.title, entities: ents }))),
      )
      return results.filter((r) => r.entities.length > 0)
    },
    enabled: !!evidence,
  })

  const extractMut = useMutation({
    mutationFn: (evidenceId: number) => evidenceApi.extractEntities(caseId, evidenceId),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['case-entities', caseId] }),
  })

  if (!evidence) return <div className="text-slate-500">Loading...</div>

  return (
    <div className="space-y-3">
      <div className="bg-white rounded-lg shadow p-4">
        <h2 className="font-semibold mb-2">Extract Entities</h2>
        <p className="text-sm text-slate-600 mb-3">Run regex + gazetteer extraction on each evidence item:</p>
        <div className="space-y-1">
          {evidence.data.map((e) => (
            <div key={e.id} className="flex items-center justify-between py-1">
              <span className="text-sm">{e.title}</span>
              <button
                onClick={() => extractMut.mutate(e.id)}
                disabled={extractMut.isPending}
                className="text-xs bg-brand-50 text-brand-700 px-2 py-1 rounded hover:bg-brand-100"
              >
                Extract
              </button>
            </div>
          ))}
        </div>
      </div>

      <div className="bg-white rounded-lg shadow p-4">
        <h2 className="font-semibold mb-2">Extracted Entities</h2>
        {allEntities.isLoading && <div className="text-slate-500">Loading entities...</div>}
        {allEntities.data && allEntities.data.length === 0 && <div className="text-slate-500 text-sm">No entities extracted yet.</div>}
        {allEntities.data && allEntities.data.map((group) => (
          <div key={group.evidenceId} className="mb-3">
            <div className="text-sm font-medium text-slate-700 mb-1">{group.evidenceTitle}</div>
            <div className="flex flex-wrap gap-1">
              {group.entities.map((ent) => (
                <span
                  key={ent.id}
                  className={`text-xs px-2 py-0.5 rounded ${entityColor(ent.type)}`}
                >
                  {ent.type}: {ent.value}
                </span>
              ))}
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

function entityColor(type: string): string {
  const colors: Record<string, string> = {
    person: 'bg-blue-100 text-blue-700',
    organization: 'bg-purple-100 text-purple-700',
    location: 'bg-green-100 text-green-700',
    other: 'bg-slate-100 text-slate-700',
  }
  return colors[type] || 'bg-slate-100 text-slate-700'
}
