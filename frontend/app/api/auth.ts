import type { ApiFetch } from './main'

export interface AuthAPI {
  login: (data: { username: string, password: string }) => Promise<{
    token: string
  }>
  logout: () => Promise<void>
}

export default (apiFetch: ApiFetch): AuthAPI => ({
  login(data) {
    return apiFetch(`/auth/login`, {
      method: 'POST',
      body: data
    })
  },

  logout() {
    return apiFetch(`/auth/logout`, {
      method: 'POST'
    })
  }
})
