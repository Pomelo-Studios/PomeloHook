export function UsersPanel({ apiKey }: { apiKey: string }) {
  return <div className="p-4 text-zinc-600 text-xs">Users — {apiKey ? 'authed' : 'cli mode'}</div>
}
