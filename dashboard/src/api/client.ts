import type { WebhookEvent, Tunnel } from '../types'

const BASE = ''

async function request<T>(path: string, opts?: RequestInit): Promise<T> {
  const res = await fetch(BASE + path, {
    ...opts,
    headers: { 'Content-Type': 'application/json', ...opts?.headers },
  })
  if (!res.ok) throw new Error(`${res.status} ${res.statusText}`)
  return res.json()
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
}
