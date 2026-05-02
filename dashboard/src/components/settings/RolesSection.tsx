import { useEffect, useState } from 'react'
import type { OrgRole } from '../../types'
import { api } from '../../api/client'

const ALL_PERMISSIONS = [
  'view_events',
  'replay_events',
  'create_org_tunnel',
  'delete_org_tunnel',
  'manage_members',
  'change_member_role',
  'edit_org_settings',
  'manage_roles',
] as const

interface Props {
  apiKey: string
  can: (perm: string) => boolean
}

export function RolesSection({ apiKey, can }: Props) {
  const [roles, setRoles] = useState<OrgRole[]>([])
  const [loading, setLoading] = useState(true)
  const [editing, setEditing] = useState<string | null>(null)
  const [editPerms, setEditPerms] = useState<string[]>([])
  const [showNew, setShowNew] = useState(false)
  const [newName, setNewName] = useState('')
  const [newDisplay, setNewDisplay] = useState('')
  const [newPerms, setNewPerms] = useState<string[]>([])

  useEffect(() => {
    api.org.listRoles(apiKey).then(r => { setRoles(r); setLoading(false) })
  }, [apiKey])

  function startEdit(role: OrgRole) {
    setEditing(role.name)
    setEditPerms([...role.permissions])
  }

  async function saveEdit(role: OrgRole) {
    const updated = await api.org.updateRole(apiKey, role.name, {
      display_name: role.display_name,
      permissions: editPerms,
    })
    setRoles(r => r.map(x => x.name === updated.name ? updated : x))
    setEditing(null)
  }

  async function handleDelete(name: string) {
    await api.org.deleteRole(apiKey, name)
    setRoles(r => r.filter(x => x.name !== name))
  }

  async function handleCreate() {
    const created = await api.org.createRole(apiKey, {
      name: newName,
      display_name: newDisplay,
      permissions: newPerms,
    })
    setRoles(r => [...r, created])
    setShowNew(false)
    setNewName('')
    setNewDisplay('')
    setNewPerms([])
  }

  function togglePerm(perm: string, current: string[], set: (p: string[]) => void) {
    set(current.includes(perm) ? current.filter(p => p !== perm) : [...current, perm])
  }

  if (loading) return <div style={{ color: '#555', fontSize: 12 }}>Loading…</div>

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 12 }}>
        <span style={{ fontSize: 12, color: '#555' }}>{roles.length} role{roles.length !== 1 ? 's' : ''}</span>
        {can('manage_roles') && (
          <button onClick={() => setShowNew(true)} style={btnStyle}>+ New Role</button>
        )}
      </div>

      {roles.map(role => (
        <div key={role.name} style={{ background: '#222', borderRadius: 6, padding: '10px 14px', marginBottom: 8 }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <div>
              <span style={{ fontSize: 12, color: '#ccc', fontWeight: 500 }}>{role.display_name}</span>
              <span style={{ fontSize: 10, color: '#444', marginLeft: 6 }}>{role.name}</span>
              {role.is_system && <span style={{ fontSize: 9, color: '#444', marginLeft: 4 }}>(system)</span>}
            </div>
            {can('manage_roles') && (
              <div style={{ display: 'flex', gap: 6 }}>
                {editing === role.name ? (
                  <>
                    <button onClick={() => saveEdit(role)} style={btnStyle}>Save</button>
                    <button onClick={() => setEditing(null)} style={cancelBtnStyle}>Cancel</button>
                  </>
                ) : (
                  <>
                    <button onClick={() => startEdit(role)} style={iconBtnStyle} title="Edit permissions">✎</button>
                    <button
                      onClick={() => !role.is_system && handleDelete(role.name)}
                      disabled={role.is_system}
                      title={role.is_system ? 'System roles cannot be deleted' : 'Delete role'}
                      style={{ ...iconBtnStyle, opacity: role.is_system ? 0.3 : 1, cursor: role.is_system ? 'not-allowed' : 'pointer' }}
                    >
                      ✕
                    </button>
                  </>
                )}
              </div>
            )}
          </div>

          {editing === role.name ? (
            <div style={{ display: 'flex', flexWrap: 'wrap', gap: 10, marginTop: 10 }}>
              {ALL_PERMISSIONS.map(p => (
                <label key={p} style={{ display: 'flex', alignItems: 'center', gap: 4, fontSize: 11, color: '#888', cursor: 'pointer' }}>
                  <input
                    type="checkbox"
                    checked={editPerms.includes(p)}
                    onChange={() => togglePerm(p, editPerms, setEditPerms)}
                  />
                  {p}
                </label>
              ))}
            </div>
          ) : (
            <div style={{ display: 'flex', flexWrap: 'wrap', gap: 4, marginTop: 6 }}>
              {role.permissions.length === 0 ? (
                <span style={{ fontSize: 10, color: '#444' }}>no permissions</span>
              ) : role.permissions.map(p => (
                <span key={p} style={{ fontSize: 9, padding: '1px 6px', background: '#2a2a2a', borderRadius: 3, color: '#555' }}>{p}</span>
              ))}
            </div>
          )}
        </div>
      ))}

      {showNew && (
        <div style={modalOverlay}>
          <div style={modalBox}>
            <h3 style={{ fontSize: 13, color: '#ccc', marginBottom: 12 }}>New Role</h3>
            <label style={labelStyle}>Role ID (lowercase, underscores)</label>
            <input
              value={newName}
              onChange={e => setNewName(e.target.value.toLowerCase().replace(/[^a-z0-9_]/g, '_'))}
              style={inputStyle}
              placeholder="e.g. viewer"
            />
            <label style={labelStyle}>Display Name</label>
            <input value={newDisplay} onChange={e => setNewDisplay(e.target.value)} style={inputStyle} placeholder="e.g. Viewer" />
            <label style={labelStyle}>Permissions</label>
            <div style={{ display: 'flex', flexWrap: 'wrap', gap: 10, marginBottom: 12 }}>
              {ALL_PERMISSIONS.map(p => (
                <label key={p} style={{ display: 'flex', alignItems: 'center', gap: 4, fontSize: 11, color: '#888', cursor: 'pointer' }}>
                  <input type="checkbox" checked={newPerms.includes(p)} onChange={() => togglePerm(p, newPerms, setNewPerms)} />
                  {p}
                </label>
              ))}
            </div>
            <div style={{ display: 'flex', gap: 8 }}>
              <button onClick={handleCreate} disabled={!newName || !newDisplay} style={{ ...btnStyle, opacity: (!newName || !newDisplay) ? 0.5 : 1 }}>
                Create
              </button>
              <button onClick={() => setShowNew(false)} style={cancelBtnStyle}>Cancel</button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

const btnStyle: React.CSSProperties = { fontSize: 11, padding: '4px 12px', background: 'rgba(255,107,107,0.13)', color: '#FF6B6B', border: '1px solid rgba(255,107,107,0.3)', borderRadius: 6, cursor: 'pointer' }
const iconBtnStyle: React.CSSProperties = { fontSize: 12, padding: '2px 6px', background: 'transparent', color: '#555', border: '1px solid #2a2a2a', borderRadius: 4, cursor: 'pointer' }
const cancelBtnStyle: React.CSSProperties = { fontSize: 11, padding: '4px 12px', background: 'transparent', color: '#888', border: '1px solid #2a2a2a', borderRadius: 6, cursor: 'pointer' }
const inputStyle: React.CSSProperties = { display: 'block', width: '100%', marginBottom: 8, padding: '5px 8px', background: '#222', color: '#ccc', border: '1px solid #2a2a2a', borderRadius: 4, fontSize: 12, boxSizing: 'border-box' }
const labelStyle: React.CSSProperties = { display: 'block', fontSize: 11, color: '#555', marginBottom: 3 }
const modalOverlay: React.CSSProperties = { position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.6)', display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 50 }
const modalBox: React.CSSProperties = { background: '#1a1a1a', border: '1px solid #2a2a2a', borderRadius: 8, padding: 20, width: 360, maxWidth: '90vw' }
