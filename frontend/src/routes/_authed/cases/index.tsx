import { createFileRoute, Link } from '@tanstack/react-router'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { caseApi } from '@/lib/api-services'
import { queryKeys } from '@/lib/queryKeys'
import type { Case } from '@/lib/types'

export const Route = createFileRoute('/_authed/cases/')({
  component: CaseListPage,
})

function CaseListPage() {
  const [status, setStatus] = useState('')
  const [q, setQ] = useState('')
  const [page, setPage] = useState(1)
  const [showCreate, setShowCreate] = useState(false)

  const { data, isLoading } = useQuery({
    queryKey: queryKeys.caseList({ status, q, page }),
    queryFn: () => caseApi.list({ status, q, page, limit: 20 }),
  })

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Cases</h1>
        <button onClick={() => setShowCreate(true)} className="bg-brand-600 text-white px-4 py-2 rounded hover:bg-brand-700">
          New Case
        </button>
      </div>

      {/* Filters */}
      <div className="flex gap-3 bg-white p-3 rounded shadow">
        <input
          placeholder="Search..."
          value={q}
          onChange={(e) => { setQ(e.target.value); setPage(1) }}
          className="flex-1 px-3 py-1 border rounded"
        />
        <select value={status} onChange={(e) => { setStatus(e.target.value); setPage(1) }} className="px-3 py-1 border rounded">
          <option value="">All statuses</option>
          <option value="open">Open</option>
          <option value="closed">Closed</option>
          <option value="archived">Archived</option>
        </select>
      </div>

      {isLoading && <div className="text-slate-500">Loading...</div>}
      {data && (
        <>
          <div className="grid gap-3">
            {data.data.length === 0 && <div className="text-slate-500 text-center py-8">No cases found.</div>}
            {data.data.map((c) => <CaseCard key={c.id} case={c} />)}
          </div>
          {data.total > 20 && (
            <div className="flex justify-center gap-2">
              <button disabled={page <= 1} onClick={() => setPage(page - 1)} className="px-3 py-1 border rounded disabled:opacity-50">Prev</button>
              <span className="px-3 py-1">Page {page}</span>
              <button disabled={page * 20 >= data.total} onClick={() => setPage(page + 1)} className="px-3 py-1 border rounded disabled:opacity-50">Next</button>
            </div>
          )}
        </>
      )}

      {showCreate && <CreateCaseModal onClose={() => setShowCreate(false)} />}
    </div>
  )
}

function CaseCard({ case: c }: { case: Case }) {
  return (
    <Link
      to="/cases/$caseId"
      params={{ caseId: String(c.id) }}
      className="block bg-white p-4 rounded shadow hover:shadow-md transition"
    >
      <div className="flex items-center justify-between">
        <h3 className="font-semibold">{c.title}</h3>
        <span className={`text-xs px-2 py-0.5 rounded ${c.status === 'open' ? 'bg-green-100 text-green-700' : 'bg-slate-100 text-slate-600'}`}>
          {c.status}
        </span>
      </div>
      {c.description && <p className="text-sm text-slate-600 mt-1 line-clamp-2">{c.description}</p>}
      <div className="text-xs text-slate-400 mt-2">
        Owner: {c.owner_name} · Updated {new Date(c.updated_at).toLocaleDateString()}
      </div>
      {c.tags.length > 0 && (
        <div className="flex gap-1 mt-2">
          {c.tags.map((t) => <span key={t} className="text-xs bg-brand-50 text-brand-700 px-2 py-0.5 rounded">#{t}</span>)}
        </div>
      )}
    </Link>
  )
}

function CreateCaseModal({ onClose }: { onClose: () => void }) {
  const qc = useQueryClient()
  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [visibility, setVisibility] = useState('private')
  const [tags, setTags] = useState('')
  const [error, setError] = useState('')

  const mutation = useMutation({
    mutationFn: () => caseApi.create({
      title,
      description,
      visibility,
      tags: tags.split(',').map((t) => t.trim()).filter(Boolean),
    }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.cases })
      onClose()
    },
    onError: () => setError('Failed to create case'),
  })

  return (
    <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50" onClick={onClose}>
      <div className="bg-white rounded-lg p-6 w-full max-w-lg" onClick={(e) => e.stopPropagation()}>
        <h2 className="text-lg font-bold mb-4">New Case</h2>
        {error && <div className="mb-3 p-2 bg-red-50 text-red-700 rounded text-sm">{error}</div>}
        <div className="space-y-3">
          <div>
            <label className="block text-sm font-medium mb-1">Title</label>
            <input value={title} onChange={(e) => setTitle(e.target.value)} className="w-full px-3 py-2 border rounded" />
          </div>
          <div>
            <label className="block text-sm font-medium mb-1">Description</label>
            <textarea value={description} onChange={(e) => setDescription(e.target.value)} rows={3} className="w-full px-3 py-2 border rounded" />
          </div>
          <div>
            <label className="block text-sm font-medium mb-1">Visibility</label>
            <select value={visibility} onChange={(e) => setVisibility(e.target.value)} className="w-full px-3 py-2 border rounded">
              <option value="private">Private</option>
              <option value="public">Public</option>
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium mb-1">Tags (comma-separated)</label>
            <input value={tags} onChange={(e) => setTags(e.target.value)} className="w-full px-3 py-2 border rounded" placeholder="tag1, tag2" />
          </div>
        </div>
        <div className="flex justify-end gap-2 mt-4">
          <button onClick={onClose} className="px-4 py-2 border rounded">Cancel</button>
          <button onClick={() => mutation.mutate()} disabled={!title || mutation.isPending} className="px-4 py-2 bg-brand-600 text-white rounded disabled:opacity-50">
            {mutation.isPending ? 'Creating...' : 'Create'}
          </button>
        </div>
      </div>
    </div>
  )
}
