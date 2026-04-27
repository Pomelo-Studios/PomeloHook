import { useState, useEffect } from 'react'
import { api } from '../../api/client'
import type { Org } from '../../types'

interface Props { apiKey: string }

export function OrgsPanel({ apiKey }: Props) {
  const [org, setOrg] = useState<Org | null>(null)
  const [editing, setEditing] = useState(false)
  const [name, setName] = useState('')
  const [error, setError] = useState('')

  useEffect(() => {
    api.admin.getOrg(apiKey).then(o => { setOrg(o); setName(o.Name) }).catch(() => setError('Failed to load org'))
  }, [apiKey])

  async function handleSave() {
    if (!org) return
    setError('')
    try {
      const updated = await api.admin.updateOrg(apiKey, org.ID, name)
      setOrg(updated)
      setEditing(false)
    } catch {
      setError('Save failed')
    }
  }

  return (
    <div className="flex flex-col h-full">
      <div className="h-11 border-b border-zinc-800 flex items-center px-4 flex-shrink-0 bg-zinc-900/30">
        <span className="text-zinc-300 text-xs font-semibold">Organization</span>
      </div>
      {error && <div className="bg-red-950 text-red-400 text-xs px-4 py-2 border-b border-red-900">{error}</div>}
      {org && (
        <div className="p-6 max-w-sm">
          <div className="flex flex-col gap-4">
            <div>
              <p className="text-[9px] text-zinc-500 uppercase tracking-widest mb-1">ID</p>
              <p className="text-xs text-zinc-400 font-mono">{org.ID}</p>
            </div>
            <div>
              <p className="text-[9px] text-zinc-500 uppercase tracking-widest mb-1">Name</p>
              {editing ? (
                <div className="flex gap-2 items-center">
                  <input value={name} onChange={e => setName(e.target.value)}
                    className="bg-zinc-800 border border-zinc-700 rounded px-2 py-1 text-xs text-zinc-200 outline-none focus:border-zinc-500 w-48" />
                  <button onClick={handleSave} className="text-[10px] px-2.5 py-1 bg-emerald-700 hover:bg-emerald-600 text-emerald-50 rounded font-medium">Save</button>
                  <button onClick={() => { setEditing(false); setName(org.Name) }} className="text-[10px] px-2.5 py-1 border border-zinc-700 text-zinc-400 rounded">Cancel</button>
                </div>
              ) : (
                <div className="flex items-center gap-2">
                  <p className="text-xs text-zinc-200">{org.Name}</p>
                  <button onClick={() => setEditing(true)} className="text-[10px] px-2 py-0.5 border border-zinc-700 text-zinc-500 rounded hover:text-zinc-300">Edit</button>
                </div>
              )}
            </div>
            <div>
              <p className="text-[9px] text-zinc-500 uppercase tracking-widest mb-1">Created</p>
              <p className="text-xs text-zinc-400 font-mono">{org.CreatedAt}</p>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
