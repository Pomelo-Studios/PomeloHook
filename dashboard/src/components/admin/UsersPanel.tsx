import { useState, useEffect } from 'react'
import { Plus, Pencil, Trash2, RotateCcw, X } from 'lucide-react'
import { api } from '../../api/client'
import { ConfirmDialog } from './ConfirmDialog'
import type { User, ConfirmState } from '../../types'

interface Props { apiKey: string }

type FormState = { email: string; name: string; role: string }
const emptyForm: FormState = { email: '', name: '', role: 'member' }

export function UsersPanel({ apiKey }: Props) {
  const [users, setUsers] = useState<User[]>([])
  const [loading, setLoading] = useState(true)
  const [form, setForm] = useState<FormState | null>(null)
  const [editingID, setEditingID] = useState<string | null>(null)
  const [confirm, setConfirm] = useState<ConfirmState | null>(null)
  const [newKey, setNewKey] = useState<{ userEmail: string; key: string } | null>(null)
  const [error, setError] = useState('')

  function load() {
    api.admin.listUsers(apiKey).then(setUsers).catch(() => setError('Failed to load users')).finally(() => setLoading(false))
  }

  useEffect(() => { load() }, [apiKey])

  async function handleSave() {
    if (!form) return
    setError('')
    try {
      if (editingID) {
        await api.admin.updateUser(apiKey, editingID, form)
      } else {
        await api.admin.createUser(apiKey, form)
      }
      setForm(null)
      setEditingID(null)
      load()
    } catch {
      setError('Save failed')
    }
  }

  function confirmDelete(user: User) {
    setConfirm({
      message: `Delete user ${user.Email}?`,
      detail: 'This also deletes their personal tunnels and events.',
      onConfirm: async () => {
        setConfirm(null)
        await api.admin.deleteUser(apiKey, user.ID).catch(() => setError('Delete failed'))
        load()
      },
    })
  }

  function confirmRotate(user: User) {
    setConfirm({
      message: `Rotate API key for ${user.Email}?`,
      detail: 'The current key will stop working immediately.',
      onConfirm: async () => {
        setConfirm(null)
        const result = await api.admin.rotateKey(apiKey, user.ID).catch(() => { setError('Rotate failed'); return null })
        if (result) setNewKey({ userEmail: user.Email, key: result.api_key })
        load()
      },
    })
  }

  if (loading) return <div className="p-4 text-xs font-mono" style={{ color: 'var(--text-3)' }}>Loading…</div>

  return (
    <div className="flex flex-col h-full">
      {confirm && <ConfirmDialog message={confirm.message} detail={confirm.detail} onConfirm={confirm.onConfirm} onCancel={() => setConfirm(null)} />}

      <div
        className="h-[52px] flex items-center justify-between px-5 flex-shrink-0 border-b"
        style={{ background: 'var(--surface)', borderColor: 'var(--border)' }}
      >
        <div>
          <div className="text-[14px] font-bold" style={{ color: 'var(--text)' }}>Users</div>
          <div className="text-[11px]" style={{ color: 'var(--text-3)' }}>{users.length} users</div>
        </div>
        <button
          onClick={() => { setForm(emptyForm); setEditingID(null) }}
          className="flex items-center gap-[6px] rounded-lg px-[14px] py-[7px] text-[11px] font-bold transition-opacity hover:opacity-90"
          style={{ background: 'var(--coral)', color: 'white' }}
        >
          <Plus size={12} strokeWidth={2.5} />
          New user
        </button>
      </div>

      {error && (
        <div className="text-xs px-5 py-2 border-b flex-shrink-0" style={{ background: 'var(--err-bg)', color: 'var(--err-text)', borderColor: 'var(--selected-border)' }}>
          {error}
        </div>
      )}

      {form !== null && (
        <div className="border-b p-4 flex gap-3 items-end flex-shrink-0" style={{ borderColor: 'var(--border)', background: 'var(--bg)' }}>
          {(['email', 'name'] as const).map(field => (
            <div key={field} className="flex flex-col gap-1">
              <label className="text-[9px] font-bold tracking-[1.5px] uppercase" style={{ color: 'var(--text-3)' }}>
                {field === 'email' ? 'Email' : 'Name'}
              </label>
              <input
                value={form[field]}
                onChange={e => setForm(f => f && { ...f, [field]: e.target.value })}
                className={`rounded-lg px-3 py-[6px] text-xs font-mono outline-none ${field === 'email' ? 'w-44' : 'w-36'}`}
                style={{ background: 'var(--surface)', border: '1px solid var(--border)', color: 'var(--text)' }}
              />
            </div>
          ))}
          <div className="flex flex-col gap-1">
            <label className="text-[9px] font-bold tracking-[1.5px] uppercase" style={{ color: 'var(--text-3)' }}>Role</label>
            <select
              value={form.role}
              onChange={e => setForm(f => f && { ...f, role: e.target.value })}
              className="rounded-lg px-3 py-[6px] text-xs outline-none"
              style={{ background: 'var(--surface)', border: '1px solid var(--border)', color: 'var(--text)' }}
            >
              <option value="member">member</option>
              <option value="admin">admin</option>
            </select>
          </div>
          <button
            onClick={handleSave}
            className="rounded-lg px-3 py-[6px] text-[11px] font-bold transition-opacity hover:opacity-90"
            style={{ background: 'var(--coral)', color: 'white' }}
          >
            Save
          </button>
          <button
            onClick={() => { setForm(null); setEditingID(null) }}
            className="rounded-lg px-3 py-[6px] text-[11px]"
            style={{ border: '1px solid var(--border)', color: 'var(--text-2)' }}
          >
            Cancel
          </button>
        </div>
      )}

      {newKey && (
        <div className="border-b p-3 flex items-center gap-3 flex-shrink-0" style={{ background: 'var(--ok-bg)', borderColor: 'var(--border)' }}>
          <span className="text-xs" style={{ color: 'var(--text-2)' }}>New key for {newKey.userEmail}:</span>
          <code className="font-mono text-xs px-2 py-[2px] rounded select-all" style={{ color: 'var(--ok-text)', background: 'var(--surface)' }}>
            {newKey.key}
          </code>
          <button onClick={() => setNewKey(null)} className="ml-auto" style={{ color: 'var(--text-3)' }}>
            <X size={14} />
          </button>
        </div>
      )}

      <div className="flex-1 overflow-auto">
        <table className="w-full border-collapse">
          <thead className="sticky top-0">
            <tr style={{ background: 'var(--surface)' }}>
              {['Name', 'Email', 'Role', 'API Key', 'Actions'].map(h => (
                <th key={h} className="text-left text-[9px] font-bold tracking-[1.5px] uppercase px-4 py-2 border-b" style={{ color: 'var(--text-3)', borderColor: 'var(--border)' }}>
                  {h}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {users.map(u => (
              <tr key={u.ID} className="group transition-colors" style={{ borderBottom: '1px solid var(--border)' }}>
                <td className="px-4 py-3 text-xs font-semibold" style={{ color: 'var(--text)' }}>{u.Name}</td>
                <td className="px-4 py-3 text-xs font-mono" style={{ color: 'var(--text-2)' }}>{u.Email}</td>
                <td className="px-4 py-3">
                  <span
                    className="text-[10px] font-semibold px-2 py-[2px] rounded-full"
                    style={
                      u.Role === 'admin'
                        ? { background: 'var(--selected-bg)', color: 'var(--coral)' }
                        : { background: 'var(--surface2)', color: 'var(--text-3)' }
                    }
                  >
                    {u.Role}
                  </span>
                </td>
                <td className="px-4 py-3 text-[10px] font-mono" style={{ color: 'var(--text-3)' }}>{u.APIKey.slice(0, 8)}…</td>
                <td className="px-4 py-3">
                  <div className="flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                    <button
                      onClick={() => { setEditingID(u.ID); setForm({ email: u.Email, name: u.Name, role: u.Role }) }}
                      className="flex items-center gap-1 text-[10px] px-2 py-[3px] rounded-md"
                      style={{ border: '1px solid var(--border)', color: 'var(--text-2)' }}
                    >
                      <Pencil size={10} /> Edit
                    </button>
                    <button
                      onClick={() => confirmRotate(u)}
                      className="flex items-center gap-1 text-[10px] px-2 py-[3px] rounded-md"
                      style={{ border: '1px solid var(--border)', color: 'var(--text-2)' }}
                    >
                      <RotateCcw size={10} /> Rotate Key
                    </button>
                    <button
                      onClick={() => confirmDelete(u)}
                      className="flex items-center gap-1 text-[10px] px-2 py-[3px] rounded-md"
                      style={{ border: '1px solid var(--selected-border)', color: 'var(--err-text)' }}
                    >
                      <Trash2 size={10} />
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
