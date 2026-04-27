import { useState, useEffect } from 'react'

const STORAGE_KEY = 'pomelo_api_key'

export interface AuthState {
  apiKey: string
  isServerMode: boolean
  loading: boolean
  login: (key: string) => void
  logout: () => void
}

export function useAuth(): AuthState {
  const [loading, setLoading] = useState(true)
  const [isServerMode, setIsServerMode] = useState(false)
  const [apiKey, setApiKey] = useState('')

  useEffect(() => {
    fetch('/api/me')
      .then(res => {
        if (res.ok) {
          setIsServerMode(false)
        } else {
          setIsServerMode(true)
          const saved = sessionStorage.getItem(STORAGE_KEY)
          if (saved) setApiKey(saved)
        }
      })
      .catch(() => {
        setIsServerMode(true)
        const saved = sessionStorage.getItem(STORAGE_KEY)
        if (saved) setApiKey(saved)
      })
      .finally(() => setLoading(false))
  }, [])

  function login(key: string) {
    sessionStorage.setItem(STORAGE_KEY, key)
    setApiKey(key)
  }

  function logout() {
    sessionStorage.removeItem(STORAGE_KEY)
    setApiKey('')
  }

  return { apiKey, isServerMode, loading, login, logout }
}
