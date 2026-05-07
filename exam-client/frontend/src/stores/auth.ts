import { defineStore } from 'pinia'
import type { UserInfo } from '@/api/types'

const TOKEN_KEY = 'exam-client-token'
const USER_KEY = 'exam-client-user'

function readToken(): string | null {
  try {
    return localStorage.getItem(TOKEN_KEY)
  } catch {
    return null
  }
}

function readUser(): UserInfo | null {
  try {
    const raw = localStorage.getItem(USER_KEY)
    return raw ? (JSON.parse(raw) as UserInfo) : null
  } catch {
    return null
  }
}

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: typeof localStorage !== 'undefined' ? readToken() : null as string | null,
    user: typeof localStorage !== 'undefined' ? readUser() : null as UserInfo | null,
  }),
  getters: {
    isLoggedIn: (s) => !!s.token && !!s.user,
    role: (s) => s.user?.role ?? null,
  },
  actions: {
    setSession(token: string, user: UserInfo) {
      this.token = token
      this.user = user
      try {
        localStorage.setItem(TOKEN_KEY, token)
        localStorage.setItem(USER_KEY, JSON.stringify(user))
      } catch {
        /* ignore */
      }
    },
    logout() {
      this.token = null
      this.user = null
      try {
        localStorage.removeItem(TOKEN_KEY)
        localStorage.removeItem(USER_KEY)
      } catch {
        /* ignore */
      }
    },
  },
})
