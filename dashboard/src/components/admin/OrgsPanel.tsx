import { useState, useEffect } from 'react'
import { Pencil } from 'lucide-react'
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
      const updated = await api.admin.updateOrg(apiKey, name)
      setOrg(updated)
      setName(updated.Name)
      setEditing(false)
    } catch {
      setError('Save failed')
    }
  }

  return (
    <div className="flex flex-col h-full">
      <div
        className="h-[52px] flex items-center px-5 flex-shrink-0 border-b"
        style={{ background: 'var(--surface)', borderColor: 'var(--border)' }}
      >
        <div className="text-[14px] font-bold" style={{ color: 'var(--text)' }}>Organization</div>
      </div>

      {error && (
        <div className="text-xs px-5 py-2 border-b" style={{ background: 'var(--err-bg)', color: 'var(--err-text)', borderColor: 'var(--selected-border)' }}>
          {error}
        </div>
      )}

      {org && (
        <div className="p-6 flex flex-col gap-5 max-w-sm">
          {[{ label: 'ID', value: org.ID }, { label: 'Created', value: org.CreatedAt }].map(({ label, value }) => (
            <div key={label}>
              <p className="text-[9px] font-bold tracking-[1.5px] uppercase mb-1" style={{ color: 'var(--text-3)' }}>{label}</p>
              <p className="text-xs font-mono" style={{ color: 'var(--text-2)' }}>{value}</p>
            </div>
          ))}
          <div>
            <p className="text-[9px] font-bold tracking-[1.5px] uppercase mb-1" style={{ color: 'var(--text-3)' }}>Name</p>
            {editing ? (
              <div className="flex items-center gap-2">
                <input
                  value={name}
                  onChange={e => setName(e.target.value)}
                  className="rounded-lg px-3 py-[6px] text-xs font-mono outline-none w-48"
                  style={{ background: 'var(--surface)', border: '1px solid var(--border)', color: 'var(--text)' }}
                />
                <button
                  onClick={handleSave}
                  className="rounded-lg px-3 py-[6px] text-[11px] font-bold transition-opacity hover:opacity-90"
                  style={{ background: 'var(--coral)', color: 'white' }}
                >
                  Save
                </button>
                <button
                  onClick={() => { setEditing(false); setName(org.Name) }}
                  className="rounded-lg px-3 py-[6px] text-[11px]"
                  style={{ border: '1px solid var(--border)', color: 'var(--text-2)' }}
                >
                  Cancel
                </button>
              </div>
            ) : (
              <div className="flex items-center gap-2">
                <p className="text-xs" style={{ color: 'var(--text)' }}>{org.Name}</p>
                <button
                  onClick={() => setEditing(true)}
                  className="flex items-center gap-1 text-[10px] px-2 py-[3px] rounded-md"
                  style={{ border: '1px solid var(--border)', color: 'var(--text-2)' }}
                >
                  <Pencil size={10} /> Edit
                </button>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  )
}
