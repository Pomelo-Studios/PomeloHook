import { useState, useEffect } from 'react'
import { api } from '../../api/client'
import { ConfirmDialog } from './ConfirmDialog'
import type { User } from '../../types'

interface Props { apiKey: string }

type FormState = { email: string; name: string; role: string }
const emptyForm: FormState = { email: '', name: '', role: 'member' }

export function UsersPanel({ apiKey }: Props) {
  const [users, setUsers] = useState<User[]>([])
  const [loading, setLoading] = useState(true)
  const [form, setForm] = useState<FormState | null>(null)
  const [editingID, setEditingID] = useState<string | null>(null)
  const [confirm, setConfirm] = useState<{ message: string; detail?: string; onConfirm: () => void } | null>(null)
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

  if (loading) return <div className="p-4 text-zinc-600 text-xs">Loading…</div>

  return (
    <div className="flex flex-col h-full">
      {confirm && <ConfirmDialog message={confirm.message} detail={confirm.detail} onConfirm={confirm.onConfirm} onCancel={() => setConfirm(null)} />}
      <div className="h-11 border-b border-zinc-800 flex items-center justify-between px-4 flex-shrink-0 bg-zinc-900/30">
        <span className="text-zinc-300 text-xs font-semibold">Users</span>
        <button onClick={() => { setForm(emptyForm); setEditingID(null) }} className="text-[10px] px-3 py-1 bg-emerald-700 hover:bg-emerald-600 text-emerald-50 rounded font-medium">
          + New User
        </button>
      </div>
      {error && <div className="bg-red-950 text-red-400 text-xs px-4 py-2 border-b border-red-900">{error}</div>}
      {form !== null && (
        <div className="border-b border-zinc-800 p-4 flex gap-3 items-end bg-zinc-900/20 flex-shrink-0">
          <div className="flex flex-col gap-1">
            <label className="text-[9px] text-zinc-500 uppercase tracking-wide">Email</label>
            <input value={form.email} onChange={e => setForm(f => f && { ...f, email: e.target.value })}
              className="bg-zinc-800 border border-zinc-700 rounded px-2 py-1 text-xs text-zinc-200 outline-none focus:border-zinc-500 w-44" />
          </div>
          <div className="flex flex-col gap-1">
            <label className="text-[9px] text-zinc-500 uppercase tracking-wide">Name</label>
            <input value={form.name} onChange={e => setForm(f => f && { ...f, name: e.target.value })}
              className="bg-zinc-800 border border-zinc-700 rounded px-2 py-1 text-xs text-zinc-200 outline-none focus:border-zinc-500 w-36" />
          </div>
          <div className="flex flex-col gap-1">
            <label className="text-[9px] text-zinc-500 uppercase tracking-wide">Role</label>
            <select value={form.role} onChange={e => setForm(f => f && { ...f, role: e.target.value })}
              className="bg-zinc-800 border border-zinc-700 rounded px-2 py-1 text-xs text-zinc-200 outline-none focus:border-zinc-500">
              <option value="member">member</option>
              <option value="admin">admin</option>
            </select>
          </div>
          <button onClick={handleSave} className="text-[10px] px-3 py-1.5 bg-emerald-700 hover:bg-emerald-600 text-emerald-50 rounded font-medium">Save</button>
          <button onClick={() => { setForm(null); setEditingID(null) }} className="text-[10px] px-3 py-1.5 border border-zinc-700 text-zinc-400 rounded hover:text-zinc-200">Cancel</button>
        </div>
      )}
      {newKey && (
        <div className="border-b border-zinc-800 p-3 bg-emerald-950/30 flex items-center gap-3 flex-shrink-0">
          <span className="text-zinc-400 text-xs">New key for {newKey.userEmail}:</span>
          <code className="text-emerald-400 text-xs font-mono bg-zinc-900 px-2 py-0.5 rounded select-all">{newKey.key}</code>
          <button onClick={() => setNewKey(null)} className="text-zinc-600 text-xs hover:text-zinc-400 ml-auto">✕</button>
        </div>
      )}
      <div className="flex-1 overflow-auto">
        <table className="w-full border-collapse">
          <thead className="sticky top-0">
            <tr className="bg-zinc-900/80">
              {['Name', 'Email', 'Role', 'API Key', 'Actions'].map(h => (
                <th key={h} className="text-left text-[9px] uppercase tracking-widest text-zinc-500 px-3 py-2 border-b border-zinc-800 font-semibold">{h}</th>
              ))}
            </tr>
          </thead>
          <tbody>
            {users.map(u => (
              <tr key={u.ID} className="hover:bg-zinc-900/40 group">
                <td className="px-3 py-2 text-xs text-zinc-200">{u.Name}</td>
                <td className="px-3 py-2 text-xs text-zinc-400">{u.Email}</td>
                <td className="px-3 py-2">
                  <span className={`text-[9px] px-1.5 py-0.5 rounded font-semibold uppercase ${u.Role === 'admin' ? 'bg-orange-950 text-orange-400' : 'bg-zinc-800 text-zinc-500'}`}>{u.Role}</span>
                </td>
                <td className="px-3 py-2 text-[10px] text-zinc-600 font-mono">{u.APIKey.slice(0, 8)}…</td>
                <td className="px-3 py-2">
                  <div className="flex gap-1 opacity-0 group-hover:opacity-100">
                    <button onClick={() => { setEditingID(u.ID); setForm({ email: u.Email, name: u.Name, role: u.Role }) }}
                      className="text-[10px] px-2 py-0.5 border border-zinc-700 text-zinc-400 rounded hover:text-zinc-200">Edit</button>
                    <button onClick={() => confirmRotate(u)}
                      className="text-[10px] px-2 py-0.5 border border-zinc-700 text-zinc-400 rounded hover:text-zinc-200">Rotate Key</button>
                    <button onClick={() => confirmDelete(u)}
                      className="text-[10px] px-2 py-0.5 border border-red-900 text-red-500 rounded hover:text-red-300">Delete</button>
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
