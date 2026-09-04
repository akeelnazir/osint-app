import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { caseApi } from '@/lib/api-services'
import { queryKeys } from '@/lib/queryKeys'

export function CommentsTab({ caseId }: { caseId: number }) {
  const qc = useQueryClient()
  const [body, setBody] = useState('')

  const { data: comments } = useQuery({
    queryKey: queryKeys.caseComments(caseId),
    queryFn: () => caseApi.listComments(caseId),
  })

  const mut = useMutation({
    mutationFn: () => caseApi.createComment(caseId, body),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.caseComments(caseId) })
      setBody('')
    },
  })

  return (
    <div className="space-y-3">
      <div className="bg-white rounded-lg shadow p-4">
        <h2 className="font-semibold mb-3">Add Comment</h2>
        <textarea value={body} onChange={(e) => setBody(e.target.value)} rows={3} placeholder="Write a comment... Use @username to mention." className="w-full px-3 py-2 border rounded" />
        <button onClick={() => mut.mutate()} disabled={!body || mut.isPending} className="mt-2 bg-brand-600 text-white px-4 py-1 rounded text-sm disabled:opacity-50">
          {mut.isPending ? 'Posting...' : 'Post'}
        </button>
      </div>

      <div className="space-y-2">
        {comments && comments.length === 0 && <div className="text-slate-500 text-center py-4">No comments yet.</div>}
        {comments?.map((c) => (
          <div key={c.id} className="bg-white p-3 rounded shadow">
            <div className="flex items-center justify-between">
              <span className="font-medium text-sm">{c.author}</span>
              <span className="text-xs text-slate-400">{new Date(c.created_at).toLocaleString()}</span>
            </div>
            <p className="text-sm text-slate-700 mt-1">{c.body}</p>
            {c.mentions.length > 0 && (
              <div className="text-xs text-brand-600 mt-1">Mentions: {c.mentions.map((m) => `@${m}`).join(' ')}</div>
            )}
          </div>
        ))}
      </div>
    </div>
  )
}
