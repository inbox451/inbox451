import type { ApiFetch } from './main'
import type { ApiResponse, User } from '~/types'

export interface UsersAPI {

  getCurrentUser: () => Promise<User>
  getUsers: () => Promise<ApiResponse<User>>
}

// So we don't have to deal with pagination for now
const limit = 100

export default (apiFetch: ApiFetch): UsersAPI => ({
  getCurrentUser() {
    return apiFetch(`/users/me`, {
      method: 'GET'
    })
  },

  getUsers() {
    return apiFetch(`/users?limit=${limit}`, {
      method: 'GET'
    })
  }
})
