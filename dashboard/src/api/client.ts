import type { WebhookEvent, Tunnel, User, Org, Me, TableInfo, TableResult, QueryResult } from '../types'

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
  getEvents: (tunnelID: string, limit = 50) =>
    request<WebhookEvent[]>(`/api/events?tunnel_id=${tunnelID}&limit=${limit}`),
  getTunnels: () =>
    request<Tunnel[]>('/api/tunnels'),
  replay: (eventID: string, targetURL: string) =>
    request<{ status_code: number; response_ms: number }>(`/api/events/${eventID}/replay`, {
      method: 'POST',
      body: JSON.stringify({ target_url: targetURL }),
    }),
  getMe: (apiKey: string) =>
    request<Me>('/api/me', { headers: authHeaders(apiKey) }),

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
    updateOrg: (apiKey: string, _id: string, name: string) =>
      request<Org>(`/api/admin/orgs`, { method: 'PUT', headers: authHeaders(apiKey), body: JSON.stringify({ name }) }),
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
}
