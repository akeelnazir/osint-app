import { createFileRoute } from '@tanstack/react-router'
import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { searchApi } from '@/lib/api-services'
import { queryKeys } from '@/lib/queryKeys'
import { Link } from '@tanstack/react-router'

export const Route = createFileRoute('/_authed/search')({
  component: SearchPage,
})

function SearchPage() {
  const [q, setQ] = useState('')
  const [type, setType] = useState('')
  const [submitted, setSubmitted] = useState(false)

  const { data, isLoading } = useQuery({
    queryKey: queryKeys.search({ q, type }),
    queryFn: () => searchApi.search({ q, type }),
    enabled: submitted && q.length > 0,
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    setSubmitted(true)
  }

  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-bold">Search</h1>
      <form onSubmit={handleSubmit} className="flex gap-3 bg-white p-3 rounded shadow">
        <input
          placeholder="Search cases and evidence..."
          value={q}
          onChange={(e) => setQ(e.target.value)}
          className="flex-1 px-3 py-2 border rounded"
        />
        <select value={type} onChange={(e) => setType(e.target.value)} className="px-3 py-2 border rounded">
          <option value="">All</option>
          <option value="case">Cases</option>
          <option value="evidence">Evidence</option>
        </select>
        <button type="submit" className="bg-brand-600 text-white px-4 py-2 rounded hover:bg-brand-700">Search</button>
      </form>

      {isLoading && <div className="text-slate-500">Searching...</div>}
      {data && (
        <div className="space-y-2">
          {data.data.length === 0 && <div className="text-slate-500 text-center py-4">No results.</div>}
          {data.data.map((r, i) => (
            <div key={i} className="bg-white p-3 rounded shadow">
              <div className="flex items-center gap-2">
                <span className={`text-xs px-2 py-0.5 rounded ${r.type === 'case' ? 'bg-blue-100 text-blue-700' : 'bg-amber-100 text-amber-700'}`}>
                  {r.type}
                </span>
                {r.type === 'case' ? (
                  <Link to="/cases/$caseId" params={{ caseId: String(r.id) }} className="font-medium text-brand-600 hover:underline">
                    {r.title}
                  </Link>
                ) : (
                  <Link to="/cases/$caseId" params={{ caseId: String(r.case_id || 0) }} className="font-medium text-brand-600 hover:underline">
                    {r.title}
                  </Link>
                )}
              </div>
              {r.description && <p className="text-sm text-slate-600 mt-1">{r.description}</p>}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
