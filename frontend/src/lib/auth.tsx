import { createContext, useContext, useState, type ReactNode } from 'react'
import type { User } from './types'
import { userApi, authApi } from './api-services'

interface AuthContextValue {
  user: User | null
  isLoading: boolean
  login: (email: string, password: string) => Promise<void>
  register: (email: string, username: string, password: string) => Promise<void>
  logout: () => void
}

const AuthContext = createContext<AuthContextValue | undefined>(undefined)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [isLoading, setIsLoading] = useState(false)

  // Try to load user on mount if token exists (browser only — localStorage
  // is unavailable during SSR).
  useState(() => {
    if (typeof window === 'undefined') return
    const token = localStorage.getItem('access_token')
    if (token) {
      setIsLoading(true)
      userApi.me().then((u) => setUser(u as User)).catch(() => {
        localStorage.removeItem('access_token')
        localStorage.removeItem('refresh_token')
      }).finally(() => setIsLoading(false))
    }
  })

  const login = async (email: string, password: string) => {
    const res = await authApi.login({ email, password })
    localStorage.setItem('access_token', res.access_token)
    localStorage.setItem('refresh_token', res.refresh_token)
    setUser(res.user)
  }

  const register = async (email: string, username: string, password: string) => {
    const res = await authApi.register({ email, username, password })
    localStorage.setItem('access_token', res.access_token)
    localStorage.setItem('refresh_token', res.refresh_token)
    setUser(res.user)
  }

  const logout = () => {
    localStorage.removeItem('access_token')
    localStorage.removeItem('refresh_token')
    setUser(null)
    window.location.href = '/login'
  }

  return (
    <AuthContext.Provider value={{ user, isLoading, login, register, logout }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}
