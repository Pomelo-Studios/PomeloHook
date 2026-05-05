import type { WebhookEvent, Tunnel, User, Org, Me, OrgRole, OrgMember, TableInfo, TableResult, QueryResult } from '../types'

const BASE = ''

async function request<T>(path: string, opts?: RequestInit): Promise<T> {
  const res = await fetch(BASE + path, {
    ...opts,
    headers: { 'Content-Type': 'application/json', ...opts?.headers },
  })
  if (!res.ok) throw new Error(`${res.status} ${res.statusText}`)
  if (res.status === 204) return undefined as T
  return res.json()
}

function authHeaders(apiKey: string): Record<string, string> {
  return apiKey ? { Authorization: `Bearer ${apiKey}` } : {}
}

export const api = {
  getEvents: (tunnelID: string, limit = 50, apiKey = '') =>
    request<WebhookEvent[]>(`/api/events?tunnel_id=${tunnelID}&limit=${limit}`,
      { headers: apiKey ? authHeaders(apiKey) : {} }),
  getTunnels: () =>
    request<Tunnel[]>('/api/tunnels'),
  replay: (eventID: string, targetURL: string, apiKey = '') =>
    request<{ status_code: number; response_ms: number }>(`/api/events/${eventID}/replay`, {
      method: 'POST',
      headers: apiKey ? authHeaders(apiKey) : {},
      body: JSON.stringify({ target_url: targetURL }),
    }),
  getMe: (apiKey: string) =>
    request<Me>('/api/me', { headers: authHeaders(apiKey) }),
  updateMe: (apiKey: string, name: string, email: string) =>
    request<{ id: string; email: string; name: string; role: string }>(
      '/api/me',
      { method: 'PUT', headers: authHeaders(apiKey), body: JSON.stringify({ name, email }) }
    ),
  changePassword: (apiKey: string, currentPassword: string, newPassword: string) =>
    request<void>(
      '/api/me/password',
      { method: 'POST', headers: authHeaders(apiKey), body: JSON.stringify({ current_password: currentPassword, new_password: newPassword }) }
    ),
  login: (email: string, password: string) =>
    request<{ api_key: string }>('/api/auth/login', { method: 'POST', body: JSON.stringify({ email, password }) }),

  admin: {
    listUsers: (apiKey: string) =>
      request<User[]>('/api/admin/users', { headers: authHeaders(apiKey) }),
    createUser: (apiKey: string, body: { email: string; name: string; role: string }) =>
      request<User>('/api/admin/users', { method: 'POST', headers: authHeaders(apiKey), body: JSON.stringify(body) }),
    updateUser: (apiKey: string, id: string, body: { email: string; name: string; role: string }) =>
      request<User>(`/api/admin/users/${id}`, { method: 'PUT', headers: authHeaders(apiKey), body: JSON.stringify(body) }),
    deleteUser: (apiKey: string, id: string) =>
      request<void>(`/api/admin/users/${id}`, { method: 'DELETE', headers: authHeaders(apiKey) }),
    rotateKey: (apiKey: string, id: string) =>
      request<{ api_key: string }>(`/api/admin/users/${id}/rotate-key`, { method: 'POST', headers: authHeaders(apiKey) }),
    getOrg: (apiKey: string) =>
      request<Org>('/api/admin/orgs', { headers: authHeaders(apiKey) }),
    updateOrg: (apiKey: string, name: string) =>
      request<Org>('/api/admin/orgs', { method: 'PUT', headers: authHeaders(apiKey), body: JSON.stringify({ name }) }),
    listTunnels: (apiKey: string) =>
      request<Tunnel[]>('/api/admin/tunnels', { headers: authHeaders(apiKey) }),
    deleteTunnel: (apiKey: string, id: string) =>
      request<void>(`/api/admin/tunnels/${id}`, { method: 'DELETE', headers: authHeaders(apiKey) }),
    disconnectTunnel: (apiKey: string, id: string) =>
      request<void>(`/api/admin/tunnels/${id}/disconnect`, { method: 'POST', headers: authHeaders(apiKey) }),
    listTables: (apiKey: string) =>
      request<TableInfo[]>('/api/admin/db/tables', { headers: authHeaders(apiKey) }),
    getTableRows: (apiKey: string, name: string, limit = 200, offset = 0) =>
      request<TableResult>(`/api/admin/db/tables/${name}?limit=${limit}&offset=${offset}`, { headers: authHeaders(apiKey) }),
    runQuery: (apiKey: string, sql: string) =>
      request<QueryResult>('/api/admin/db/query', { method: 'POST', headers: authHeaders(apiKey), body: JSON.stringify({ sql }) }),
  },

  org: {
    getUserTunnels: (apiKey: string) =>
      request<Tunnel[]>('/api/tunnels', { headers: authHeaders(apiKey) }),
    getTunnels: (apiKey: string) =>
      request<Tunnel[]>('/api/org/tunnels', { headers: authHeaders(apiKey) }),
    createPersonalTunnel: (apiKey: string, name = '') =>
      request<Tunnel>('/api/tunnels', {
        method: 'POST',
        headers: authHeaders(apiKey),
        body: JSON.stringify({ type: 'personal', name }),
      }),
    createOrgTunnel: (apiKey: string, name = '') =>
      request<Tunnel>('/api/tunnels', {
        method: 'POST',
        headers: authHeaders(apiKey),
        body: JSON.stringify({ type: 'org', name }),
      }),
    deleteOrgTunnel: (apiKey: string, id: string) =>
      request<void>(`/api/tunnels/${id}`, { method: 'DELETE', headers: authHeaders(apiKey) }),
    updateTunnel: (apiKey: string, id: string, displayName: string) =>
      request<Tunnel>(`/api/tunnels/${id}`, {
        method: 'PUT',
        headers: authHeaders(apiKey),
        body: JSON.stringify({ display_name: displayName }),
      }),
    listMembers: (apiKey: string) =>
      request<OrgMember[]>('/api/org/members', { headers: authHeaders(apiKey) }),
    inviteMember: (apiKey: string, body: { email: string; name: string; role: string }) =>
      request<{ id: string; email: string; name: string; role: string; api_key: string }>(
        '/api/org/members/invite',
        { method: 'POST', headers: authHeaders(apiKey), body: JSON.stringify(body) }
      ),
    removeMember: (apiKey: string, id: string) =>
      request<void>(`/api/org/members/${id}`, { method: 'DELETE', headers: authHeaders(apiKey) }),
    changeMemberRole: (apiKey: string, id: string, role: string) =>
      request<{ id: string; role: string }>(`/api/org/members/${id}/role`, {
        method: 'PUT',
        headers: authHeaders(apiKey),
        body: JSON.stringify({ role }),
      }),
    listRoles: (apiKey: string) =>
      request<OrgRole[]>('/api/org/roles', { headers: authHeaders(apiKey) }),
    createRole: (apiKey: string, body: { name: string; display_name: string; permissions: string[] }) =>
      request<OrgRole>('/api/org/roles', { method: 'POST', headers: authHeaders(apiKey), body: JSON.stringify(body) }),
    updateRole: (apiKey: string, name: string, body: { display_name: string; permissions: string[] }) =>
      request<OrgRole>(`/api/org/roles/${name}`, { method: 'PUT', headers: authHeaders(apiKey), body: JSON.stringify(body) }),
    deleteRole: (apiKey: string, name: string) =>
      request<void>(`/api/org/roles/${name}`, { method: 'DELETE', headers: authHeaders(apiKey) }),
    getSettings: (apiKey: string) =>
      request<Org>('/api/org/settings', { headers: authHeaders(apiKey) }),
    updateSettings: (apiKey: string, name: string) =>
      request<Org>('/api/org/settings', { method: 'PUT', headers: authHeaders(apiKey), body: JSON.stringify({ name }) }),
  },
}
