import type { AuthAPI } from './auth'
import authApi from './auth'
import type { ProjectsAPI } from './projects'
import projectsApi from './projects'
import type { InboxesAPI } from './inboxes'
import inboxesApi from './inboxes'
import type { MessagesAPI } from './messages'
import messagesApi from './messages'
import type { UsersAPI } from './users'
import usersApi from './users'

export type ApiFetch = (
  url: string,
  options?: {
    method?: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    body?: any
  },
// eslint-disable-next-line @typescript-eslint/no-explicit-any
) => Promise<any>

export interface Api extends AuthAPI, ProjectsAPI, InboxesAPI, MessagesAPI, UsersAPI {}

export default (apiFetch: ApiFetch): Api => ({
  ...authApi(apiFetch),
  ...projectsApi(apiFetch),
  ...inboxesApi(apiFetch),
  ...messagesApi(apiFetch),
  ...usersApi(apiFetch)
})
