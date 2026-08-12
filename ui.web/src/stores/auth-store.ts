import { create } from 'zustand'
import { getCookie, setCookie, removeCookie } from '@/lib/cookies'

const ACCESS_TOKEN = 'thisisjustarandomstring'
const AUTH_USER = 'cabinet_auth_user'
const REJECTED_LEGACY_TOKENS = new Set([
  'mock-access-token',
  'mock-passkey-access-token',
  'cabinet-local-device-session-v1',
])

interface AuthUser {
  accountNo: string
  email: string
  role: string[]
  exp: number
}

interface AuthState {
  auth: {
    user: AuthUser | null
    setUser: (user: AuthUser | null) => void
    setRemoteUser: (user: AuthUser) => void
    remoteSession: boolean
    accessToken: string
    localSessionProfileID: string
    localSessionToken: string
    setAccessToken: (accessToken: string) => void
    setLocalSession: (profileID: string, token: string) => void
    clearLocalSession: () => void
    resetAccessToken: () => void
    reset: () => void
  }
}

export const useAuthStore = create<AuthState>()((set) => {
  const cookieState = getCookie(ACCESS_TOKEN)
  const userCookieState = getCookie(AUTH_USER)
  let initToken = ''
  let initUser: AuthUser | null = null

  if (cookieState) {
    try {
      initToken = JSON.parse(cookieState) as string
    } catch {
      initToken = cookieState
    }
    if (REJECTED_LEGACY_TOKENS.has(initToken)) {
      removeCookie(ACCESS_TOKEN)
      removeCookie(AUTH_USER)
      initToken = ''
    }
  }

  if (userCookieState) {
    try {
      initUser = JSON.parse(userCookieState) as AuthUser
    } catch {
      initUser = null
    }
  }

  return {
    auth: {
      user: initUser,
      setUser: (user) =>
        set((state) => {
          if (user) {
            setCookie(AUTH_USER, JSON.stringify(user))
          } else {
            removeCookie(AUTH_USER)
          }
          return { ...state, auth: { ...state.auth, user } }
        }),
      setRemoteUser: (user) =>
        set((state) => {
          removeCookie(AUTH_USER)
          return {
            ...state,
            auth: { ...state.auth, user, remoteSession: true },
          }
        }),
      remoteSession: false,
      accessToken: initToken,
      localSessionProfileID: '',
      localSessionToken: '',
      setAccessToken: (accessToken) =>
        set((state) => {
          setCookie(ACCESS_TOKEN, JSON.stringify(accessToken))
          return { ...state, auth: { ...state.auth, accessToken } }
        }),
      setLocalSession: (profileID, token) =>
        set((state) => ({
          ...state,
          auth: {
            ...state.auth,
            localSessionProfileID: profileID,
            localSessionToken: token,
          },
        })),
      clearLocalSession: () =>
        set((state) => ({
          ...state,
          auth: {
            ...state.auth,
            localSessionProfileID: '',
            localSessionToken: '',
          },
        })),
      resetAccessToken: () =>
        set((state) => {
          removeCookie(ACCESS_TOKEN)
          return { ...state, auth: { ...state.auth, accessToken: '' } }
        }),
      reset: () =>
        set((state) => {
          removeCookie(ACCESS_TOKEN)
          removeCookie(AUTH_USER)
          return {
            ...state,
            auth: {
              ...state.auth,
              user: null,
              accessToken: '',
              localSessionProfileID: '',
              localSessionToken: '',
              remoteSession: false,
            },
          }
        }),
    },
  }
})
