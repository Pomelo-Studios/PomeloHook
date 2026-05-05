import { useEffect, useState } from 'react'
import type { OrgMember, OrgRole } from '../../types'
import { api } from '../../api/client'

interface Props {
  apiKey: string
  can: (perm: string) => boolean
}

export function MembersSection({ apiKey, can }: Props) {
  const [members, setMembers] = useState<OrgMember[]>([])
  const [roles, setRoles] = useState<OrgRole[]>([])
  const [loading, setLoading] = useState(true)
  const assignableRoles = roles.filter(r => r.name !== 'admin')
  const [showInvite, setShowInvite] = useState(false)
  const [inviteEmail, setInviteEmail] = useState('')
  const [inviteName, setInviteName] = useState('')
  const [inviteRole, setInviteRole] = useState('member')
  const [inviteApiKey, setInviteApiKey] = useState<string | null>(null)

  useEffect(() => {
    Promise.all([api.org.listMembers(apiKey), api.org.listRoles(apiKey)]).then(([m, r]) => {
      setMembers(m)
      setRoles(r)
      setLoading(false)
    })
  }, [apiKey])

  async function handleInvite() {
    const res = await api.org.inviteMember(apiKey, { email: inviteEmail, name: inviteName, role: inviteRole })
    setInviteApiKey(res.api_key)
    api.org.listMembers(apiKey).then(setMembers)
  }

  async function handleRemove(id: string) {
    await api.org.removeMember(apiKey, id)
    setMembers(m => m.filter(x => x.ID !== id))
  }

  async function handleRoleChange(id: string, role: string) {
    await api.org.changeMemberRole(apiKey, id, role)
    setMembers(m => m.map(x => x.ID === id ? { ...x, Role: role } : x))
  }

  function closeInvite() {
    setShowInvite(false)
    setInviteApiKey(null)
    setInviteEmail('')
    setInviteName('')
    setInviteRole('member')
  }

  if (loading) return <div style={{ color: 'var(--text-3)', fontSize: 12 }}>Loading…</div>

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 12 }}>
        <span style={{ fontSize: 12, color: 'var(--text-3)' }}>{members.length} member{members.length !== 1 ? 's' : ''}</span>
        {can('manage_members') && (
          <button onClick={() => setShowInvite(true)} style={btnStyle}>+ Invite</button>
        )}
      </div>

      <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 12 }}>
        <thead>
          <tr>
            <th style={thStyle}>Name</th>
            <th style={thStyle}>Email</th>
            <th style={thStyle}>Role</th>
            <th style={thStyle}>Active Tunnel</th>
            {can('manage_members') && <th style={thStyle} />}
          </tr>
        </thead>
        <tbody>
          {members.map(m => (
            <tr key={m.ID} style={{ borderBottom: '1px solid var(--border)' }}>
              <td style={tdStyle}>{m.Name}</td>
              <td style={tdStyle}>{m.Email}</td>
              <td style={tdStyle}>
                {can('change_member_role') ? (
                  <select
                    value={m.Role}
                    onChange={e => handleRoleChange(m.ID, e.target.value)}
                    style={selectStyle}
                  >
                    {assignableRoles.map(r => <option key={r.name} value={r.name}>{r.display_name}</option>)}
                  </select>
                ) : m.Role}
              </td>
              <td style={{ ...tdStyle, fontFamily: 'monospace', color: 'var(--text-3)' }}>
                {m.ActiveTunnelSubdomain || '—'}
              </td>
              {can('manage_members') && (
                <td style={tdStyle}>
                  <button onClick={() => handleRemove(m.ID)} style={removeBtnStyle}>Remove</button>
                </td>
              )}
            </tr>
          ))}
        </tbody>
      </table>

      {showInvite && (
        <div style={modalOverlay}>
          <div style={modalBox}>
            {inviteApiKey ? (
              <>
                <p style={{ fontSize: 12, color: 'var(--text-2)', marginBottom: 8 }}>Member invited. Share this API key:</p>
                <code style={{ display: 'block', padding: 8, background: 'var(--surface2)', borderRadius: 4, fontSize: 11, wordBreak: 'break-all', color: 'var(--text-3)' }}>
                  {inviteApiKey}
                </code>
                <button onClick={closeInvite} style={{ ...btnStyle, marginTop: 12 }}>Done</button>
              </>
            ) : (
              <>
                <h3 style={{ fontSize: 13, color: 'var(--text-2)', marginBottom: 12 }}>Invite Member</h3>
                <label style={labelStyle}>Name</label>
                <input value={inviteName} onChange={e => setInviteName(e.target.value)} style={inputStyle} />
                <label style={labelStyle}>Email</label>
                <input value={inviteEmail} onChange={e => setInviteEmail(e.target.value)} type="email" style={inputStyle} />
                <label style={labelStyle}>Role</label>
                <select value={inviteRole} onChange={e => setInviteRole(e.target.value)} style={{ ...selectStyle, width: '100%', padding: '5px 8px', marginBottom: 12 }}>
                  {assignableRoles.map(r => <option key={r.name} value={r.name}>{r.display_name}</option>)}
                </select>
                <div style={{ display: 'flex', gap: 8 }}>
                  <button onClick={handleInvite} style={btnStyle}>Invite</button>
                  <button onClick={closeInvite} style={cancelBtnStyle}>Cancel</button>
                </div>
              </>
            )}
          </div>
        </div>
      )}
    </div>
  )
}

const thStyle: React.CSSProperties = { padding: '4px 8px', fontWeight: 500, fontSize: 11, color: 'var(--text-3)', textAlign: 'left' }
const tdStyle: React.CSSProperties = { padding: '6px 8px', color: 'var(--text-2)' }
const btnStyle: React.CSSProperties = { fontSize: 11, padding: '4px 12px', background: 'var(--selected-bg)', color: 'var(--coral)', border: '1px solid var(--selected-border)', borderRadius: 6, cursor: 'pointer' }
const removeBtnStyle: React.CSSProperties = { fontSize: 11, padding: '2px 8px', background: 'transparent', color: 'var(--text-3)', border: '1px solid var(--border)', borderRadius: 4, cursor: 'pointer' }
const selectStyle: React.CSSProperties = { fontSize: 11, background: 'var(--surface2)', color: 'var(--text-2)', border: '1px solid var(--border)', borderRadius: 4, padding: '2px 4px' }
const inputStyle: React.CSSProperties = { display: 'block', width: '100%', marginBottom: 8, padding: '5px 8px', background: 'var(--surface2)', color: 'var(--text)', border: '1px solid var(--border)', borderRadius: 4, fontSize: 12, boxSizing: 'border-box' }
const labelStyle: React.CSSProperties = { display: 'block', fontSize: 11, color: 'var(--text-3)', marginBottom: 3 }
const cancelBtnStyle: React.CSSProperties = { fontSize: 11, padding: '4px 12px', background: 'transparent', color: 'var(--text-3)', border: '1px solid var(--border)', borderRadius: 6, cursor: 'pointer' }
const modalOverlay: React.CSSProperties = { position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.6)', display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 50 }
const modalBox: React.CSSProperties = { background: 'var(--surface)', border: '1px solid var(--border)', borderRadius: 8, padding: 20, width: 300, maxWidth: '90vw' }
