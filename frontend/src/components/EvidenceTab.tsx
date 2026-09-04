import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { evidenceApi } from '@/lib/api-services'
import { queryKeys } from '@/lib/queryKeys'
import type { Evidence, EvidenceType } from '@/lib/types'

export function EvidenceTab({ caseId }: { caseId: number }) {
  const [type, setType] = useState('')
  const [page, setPage] = useState(1)
  const [showForm, setShowForm] = useState(false)

  const { data, isLoading } = useQuery({
    queryKey: queryKeys.evidenceList(caseId, { type, page }),
    queryFn: () => evidenceApi.list(caseId, { type, page, limit: 50 }),
  })

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <div className="flex gap-2">
          <select value={type} onChange={(e) => { setType(e.target.value); setPage(1) }} className="px-3 py-1 border rounded text-sm">
            <option value="">All types</option>
            <option value="text">Text</option>
            <option value="url">URL</option>
            <option value="image">Image</option>
            <option value="video">Video</option>
            <option value="document">Document</option>
            <option value="raw">Raw</option>
          </select>
        </div>
        <button onClick={() => setShowForm(true)} className="bg-brand-600 text-white px-3 py-1 rounded text-sm hover:bg-brand-700">
          Add Evidence
        </button>
      </div>

      {isLoading && <div className="text-slate-500">Loading...</div>}
      {data && (
        <div className="space-y-2">
          {data.data.length === 0 && <div className="text-slate-500 text-center py-4">No evidence yet.</div>}
          {data.data.map((e) => <EvidenceCard key={e.id} caseId={caseId} evidence={e} />)}
        </div>
      )}

      {showForm && <EvidenceForm caseId={caseId} onClose={() => setShowForm(false)} />}
    </div>
  )
}

function EvidenceCard({ caseId, evidence }: { caseId: number; evidence: Evidence }) {
  const qc = useQueryClient()
  const deleteMut = useMutation({
    mutationFn: () => evidenceApi.delete(caseId, evidence.id),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.evidence(caseId) }),
  })

  return (
    <div className="bg-white p-3 rounded shadow">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <span className="text-xs px-2 py-0.5 rounded bg-slate-100 text-slate-600">{evidence.type}</span>
          <span className="font-medium">{evidence.title}</span>
        </div>
        <button onClick={() => deleteMut.mutate()} className="text-xs text-red-500 hover:underline">Delete</button>
      </div>
      {evidence.description && <p className="text-sm text-slate-600 mt-1">{evidence.description}</p>}
      {evidence.content && evidence.type === 'url' && (
        <a href={evidence.content} target="_blank" rel="noreferrer" className="text-sm text-brand-600 hover:underline mt-1 block">
          {evidence.content}
        </a>
      )}
      <div className="text-xs text-slate-400 mt-2">
        {evidence.source && <span>Source: {evidence.source} · </span>}
        {evidence.evidence_date && <span>Date: {new Date(evidence.evidence_date).toLocaleDateString()} · </span>}
        {evidence.latitude && evidence.longitude && <span>Geo: {evidence.latitude.toFixed(4)}, {evidence.longitude.toFixed(4)}</span>}
      </div>
      {evidence.tags.length > 0 && (
        <div className="flex gap-1 mt-1">
          {evidence.tags.map((t) => <span key={t} className="text-xs bg-brand-50 text-brand-700 px-1.5 py-0.5 rounded">#{t}</span>)}
        </div>
      )}
    </div>
  )
}

function EvidenceForm({ caseId, onClose }: { caseId: number; onClose: () => void }) {
  const qc = useQueryClient()
  const [type, setType] = useState<EvidenceType>('text')
  const [title, setTitle] = useState('')
  const [content, setContent] = useState('')
  const [url, setUrl] = useState('')
  const [source, setSource] = useState('')
  const [description, setDescription] = useState('')
  const [date, setDate] = useState('')
  const [lat, setLat] = useState('')
  const [lng, setLng] = useState('')
  const [tags, setTags] = useState('')
  const [file, setFile] = useState<File | null>(null)
  const [error, setError] = useState('')

  const mutation = useMutation({
    mutationFn: async () => {
      if (file && (type === 'image' || type === 'video' || type === 'document')) {
        const fd = new FormData()
        fd.append('title', title)
        fd.append('type', type)
        fd.append('source', source)
        fd.append('description', description)
        if (date) fd.append('evidence_date', date)
        if (lat) fd.append('latitude', lat)
        if (lng) fd.append('longitude', lng)
        fd.append('file', file)
        return evidenceApi.createFile(caseId, fd)
      }
      return evidenceApi.create(caseId, {
        type,
        title,
        content: type === 'url' ? url : content,
        url: type === 'url' ? url : undefined,
        source,
        description,
        evidence_date: date || undefined,
        latitude: lat ? Number(lat) : undefined,
        longitude: lng ? Number(lng) : undefined,
        tags: tags.split(',').map((t) => t.trim()).filter(Boolean),
      })
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.evidence(caseId) })
      onClose()
    },
    onError: () => setError('Failed to create evidence'),
  })

  return (
    <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50" onClick={onClose}>
      <div className="bg-white rounded-lg p-6 w-full max-w-lg max-h-[90vh] overflow-y-auto" onClick={(e) => e.stopPropagation()}>
        <h2 className="text-lg font-bold mb-4">Add Evidence</h2>
        {error && <div className="mb-3 p-2 bg-red-50 text-red-700 rounded text-sm">{error}</div>}
        <div className="space-y-3">
          <div>
            <label className="block text-sm font-medium mb-1">Type</label>
            <select value={type} onChange={(e) => setType(e.target.value as EvidenceType)} className="w-full px-3 py-2 border rounded">
              <option value="text">Text Note</option>
              <option value="url">URL</option>
              <option value="image">Image</option>
              <option value="video">Video</option>
              <option value="document">Document (PDF)</option>
              <option value="raw">Raw Data (JSON)</option>
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium mb-1">Title</label>
            <input value={title} onChange={(e) => setTitle(e.target.value)} className="w-full px-3 py-2 border rounded" />
          </div>
          {type === 'url' && (
            <div>
              <label className="block text-sm font-medium mb-1">URL</label>
              <input value={url} onChange={(e) => setUrl(e.target.value)} placeholder="https://..." className="w-full px-3 py-2 border rounded" />
            </div>
          )}
          {type === 'text' && (
            <div>
              <label className="block text-sm font-medium mb-1">Content</label>
              <textarea value={content} onChange={(e) => setContent(e.target.value)} rows={4} className="w-full px-3 py-2 border rounded" />
            </div>
          )}
          {type === 'raw' && (
            <div>
              <label className="block text-sm font-medium mb-1">JSON Content</label>
              <textarea value={content} onChange={(e) => setContent(e.target.value)} rows={4} placeholder='{"key": "value"}' className="w-full px-3 py-2 border rounded font-mono text-sm" />
            </div>
          )}
          {(type === 'image' || type === 'video' || type === 'document') && (
            <div>
              <label className="block text-sm font-medium mb-1">File</label>
              <input type="file" onChange={(e) => setFile(e.target.files?.[0] || null)} className="w-full" />
            </div>
          )}
          <div>
            <label className="block text-sm font-medium mb-1">Source</label>
            <input value={source} onChange={(e) => setSource(e.target.value)} className="w-full px-3 py-2 border rounded" />
          </div>
          <div>
            <label className="block text-sm font-medium mb-1">Description</label>
            <textarea value={description} onChange={(e) => setDescription(e.target.value)} rows={2} className="w-full px-3 py-2 border rounded" />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-sm font-medium mb-1">Evidence Date</label>
              <input type="datetime-local" value={date} onChange={(e) => setDate(e.target.value)} className="w-full px-3 py-2 border rounded" />
            </div>
            <div>
              <label className="block text-sm font-medium mb-1">Tags (comma-separated)</label>
              <input value={tags} onChange={(e) => setTags(e.target.value)} className="w-full px-3 py-2 border rounded" />
            </div>
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-sm font-medium mb-1">Latitude</label>
              <input type="number" step="any" value={lat} onChange={(e) => setLat(e.target.value)} className="w-full px-3 py-2 border rounded" />
            </div>
            <div>
              <label className="block text-sm font-medium mb-1">Longitude</label>
              <input type="number" step="any" value={lng} onChange={(e) => setLng(e.target.value)} className="w-full px-3 py-2 border rounded" />
            </div>
          </div>
        </div>
        <div className="flex justify-end gap-2 mt-4">
          <button onClick={onClose} className="px-4 py-2 border rounded">Cancel</button>
          <button onClick={() => mutation.mutate()} disabled={!title || mutation.isPending} className="px-4 py-2 bg-brand-600 text-white rounded disabled:opacity-50">
            {mutation.isPending ? 'Adding...' : 'Add'}
          </button>
        </div>
      </div>
    </div>
  )
}
