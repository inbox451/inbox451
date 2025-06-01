import type { ApiFetch } from './main'
import type { ApiResponse, Inbox } from '~/types'

export interface InboxesAPI {
  createInbox: (projectId: number, email: string) => Promise<{
    id: string
  }>
  getInboxes: (projectId: number) => Promise<ApiResponse<Inbox>>
  deleteInbox: (projectId: number, id: string) => Promise<void>
}

const projectPath = (projectId: number) => `/projects/${projectId}`

// So we don't have to deal with pagination for now
const limit = 100

export default (apiFetch: ApiFetch): InboxesAPI => ({
  createInbox(projectId, data) {
    return apiFetch(`${projectPath(projectId)}/inboxes`, {
      method: 'POST',
      body: {
        email: data
      }
    })
  },

  getInboxes(projectId) {
    return apiFetch(`${projectPath(projectId)}/inboxes?limit=${limit}`, {
      method: 'GET'
    })
  },

  deleteInbox(projectId, id) {
    return apiFetch(`${projectPath(projectId)}/inboxes/${id}`, {
      method: 'DELETE'
    })
  }
})
