import { useState } from 'react'

interface Props {
  onLogin: (apiKey: string) => void
}

export function LoginForm({ onLogin }: Props) {
  const [email, setEmail] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      const res = await fetch('/api/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email }),
      })
      if (!res.ok) throw new Error('Invalid email or unauthorized')
      const data = await res.json() as { api_key: string }
      onLogin(data.api_key)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Login failed')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="flex items-center justify-center h-screen bg-zinc-950">
      <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-8 w-80">
        <div className="flex items-center gap-2 mb-6">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#10b981" strokeWidth="2.5">
            <path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z" />
          </svg>
          <span className="text-zinc-50 text-sm font-bold">PomeloHook Admin</span>
        </div>
        <form onSubmit={handleSubmit} className="flex flex-col gap-3">
          <input
            type="email"
            placeholder="Email"
            value={email}
            onChange={e => setEmail(e.target.value)}
            required
            className="bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-xs text-zinc-200 placeholder-zinc-600 outline-none focus:border-zinc-500 font-mono"
          />
          {error && <p className="text-red-400 text-xs">{error}</p>}
          <button
            type="submit"
            disabled={loading}
            className="bg-emerald-700 hover:bg-emerald-600 text-emerald-50 text-xs py-2 rounded font-medium disabled:opacity-50"
          >
            {loading ? 'Signing in…' : 'Sign in'}
          </button>
        </form>
      </div>
    </div>
  )
}
