import type { ApiFetch } from './main'
import type { ApiResponse, Messages } from '~/types'

export interface MessagesAPI {
  getMessages: (projectId: number, inboxId: string) => Promise<ApiResponse<Messages>>
  deleteMessage: (projectId: number, inboxId: string, id: number) => Promise<void>
  markAsRead: (projectId: number, inboxId: string, id: number) => Promise<void>
  markAsUnread: (projectId: number, inboxId: string, id: number) => Promise<void>
}

const inboxPath = (projectId: number, inboxId: string) => `/projects/${projectId}/inboxes/${inboxId}`

// So we don't have to deal with pagination for now
const limit = 100

export default (apiFetch: ApiFetch): MessagesAPI => ({
  getMessages(projectId, inboxId) {
    return apiFetch(`${inboxPath(projectId, inboxId)}/messages?limit=${limit}`, {
      method: 'GET'
    })
  },

  deleteMessage(projectId, inboxId, id) {
    return apiFetch(`${inboxPath(projectId, inboxId)}/messages/${id}`, {
      method: 'DELETE'
    })
  },

  markAsRead(projectId, inboxId, id) {
    return apiFetch(`${inboxPath(projectId, inboxId)}/messages/${id}/read`, {
      method: 'PUT'
    })
  },

  markAsUnread(projectId, inboxId, id) {
    return apiFetch(`${inboxPath(projectId, inboxId)}/messages/${id}/unread`, {
      method: 'PUT'
    })
  }
})
