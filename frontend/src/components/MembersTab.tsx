import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { caseApi } from '@/lib/api-services'
import { queryKeys } from '@/lib/queryKeys'

export function MembersTab({ caseId }: { caseId: number }) {
  const qc = useQueryClient()
  const [username, setUsername] = useState('')
  const [role, setRole] = useState('viewer')

  const { data: members } = useQuery({
    queryKey: queryKeys.caseMembers(caseId),
    queryFn: () => caseApi.listMembers(caseId),
  })

  const addMut = useMutation({
    mutationFn: () => caseApi.addMember(caseId, { username, role }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.caseMembers(caseId) })
      setUsername('')
    },
  })

  const removeMut = useMutation({
    mutationFn: (userId: number) => caseApi.removeMember(caseId, userId),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.caseMembers(caseId) }),
  })

  return (
    <div className="space-y-3">
      <div className="bg-white rounded-lg shadow p-4">
        <h2 className="font-semibold mb-3">Add Member</h2>
        <div className="flex gap-2">
          <input value={username} onChange={(e) => setUsername(e.target.value)} placeholder="Username" className="flex-1 px-3 py-2 border rounded" />
          <select value={role} onChange={(e) => setRole(e.target.value)} className="px-3 py-2 border rounded">
            <option value="viewer">Viewer</option>
            <option value="editor">Editor</option>
          </select>
          <button onClick={() => addMut.mutate()} disabled={!username || addMut.isPending} className="bg-brand-600 text-white px-4 py-2 rounded disabled:opacity-50">
            Add
          </button>
        </div>
      </div>

      <div className="bg-white rounded-lg shadow p-4">
        <h2 className="font-semibold mb-3">Members</h2>
        {members && members.length === 0 && <div className="text-slate-500 text-sm">No members (only owner).</div>}
        <ul className="space-y-2">
          {members?.map((m) => (
            <li key={m.user_id} className="flex items-center justify-between py-1">
              <div>
                <span className="font-medium text-sm">{m.username}</span>
                <span className="text-xs text-slate-500 ml-2">({m.email})</span>
              </div>
              <div className="flex items-center gap-2">
                <span className="text-xs px-2 py-0.5 rounded bg-slate-100">{m.role}</span>
                {m.role !== 'owner' && (
                  <button onClick={() => removeMut.mutate(m.user_id)} className="text-xs text-red-500 hover:underline">Remove</button>
                )}
              </div>
            </li>
          ))}
        </ul>
      </div>
    </div>
  )
}
