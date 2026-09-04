import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { caseApi } from '@/lib/api-services'
import { queryKeys } from '@/lib/queryKeys'
import type { EvidenceType } from '@/lib/types'

export function TimelineTab({ caseId }: { caseId: number }) {
  const [type, setType] = useState('')
  const { data, isLoading } = useQuery({
    queryKey: queryKeys.caseTimeline(caseId, { type }),
    queryFn: () => caseApi.timeline(caseId, type ? { type } : undefined),
  })

  if (isLoading) return <div className="text-slate-500">Loading timeline...</div>
  if (!data || data.length === 0) return <div className="text-slate-500 text-center py-4">No dated evidence.</div>

  return (
    <div className="space-y-3">
      <select value={type} onChange={(e) => setType(e.target.value)} className="px-3 py-1 border rounded text-sm">
        <option value="">All types</option>
        <option value="text">Text</option>
        <option value="url">URL</option>
        <option value="image">Image</option>
        <option value="video">Video</option>
        <option value="document">Document</option>
        <option value="raw">Raw</option>
      </select>

      {/* Horizontal timeline */}
      <div className="bg-white rounded-lg shadow p-4 overflow-x-auto">
        <div className="relative" style={{ minWidth: '600px' }}>
          {/* Line */}
          <div className="absolute left-0 right-0 top-1/2 h-0.5 bg-slate-300" />
          {/* Events */}
          <div className="flex justify-between items-center relative" style={{ height: '200px' }}>
            {data.map((event, i) => {
              const isTop = i % 2 === 0
              return (
                <div key={event.evidence_id} className="relative flex flex-col items-center" style={{ flex: 1 }}>
                  {isTop ? (
                    <>
                      <div className="text-xs text-slate-600 text-center mb-2 max-w-[120px]">
                        <div className="font-medium truncate">{event.title}</div>
                        {event.date && <div className="text-slate-400">{new Date(event.date).toLocaleDateString()}</div>}
                      </div>
                      <div className="w-px h-8 bg-slate-300" />
                      <div className={`w-3 h-3 rounded-full ${typeColor(event.type)} z-10`} />
                    </>
                  ) : (
                    <>
                      <div className={`w-3 h-3 rounded-full ${typeColor(event.type)} z-10`} />
                      <div className="w-px h-8 bg-slate-300" />
                      <div className="text-xs text-slate-600 text-center mt-2 max-w-[120px]">
                        <div className="font-medium truncate">{event.title}</div>
                        {event.date && <div className="text-slate-400">{new Date(event.date).toLocaleDateString()}</div>}
                      </div>
                    </>
                  )}
                </div>
              )
            })}
          </div>
        </div>
      </div>
    </div>
  )
}

function typeColor(type: EvidenceType): string {
  const colors: Record<EvidenceType, string> = {
    text: 'bg-blue-500',
    url: 'bg-green-500',
    image: 'bg-purple-500',
    video: 'bg-red-500',
    document: 'bg-amber-500',
    raw: 'bg-slate-500',
  }
  return colors[type] || 'bg-slate-500'
}
