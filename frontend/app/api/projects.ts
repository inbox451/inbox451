import type { ApiFetch } from './main'
import type { ApiResponse, Project } from '~/types'

export interface ProjectsAPI {
  createProject: (name: string) => Promise<{
    id: string
  }>
  getProjects: () => Promise<ApiResponse<Project>>
  deleteProject: (id: string) => Promise<void>
}

// So we don't have to deal with pagination for now
const limit = 100

export default (apiFetch: ApiFetch): ProjectsAPI => ({
  createProject(data) {
    return apiFetch(`/projects`, {
      method: 'POST',
      body: {
        name: data
      }
    })
  },

  getProjects() {
    return apiFetch(`/projects?limit=${limit}`, {
      method: 'GET'
    })
  },

  deleteProject(id) {
    return apiFetch(`/projects/${id}`, {
      method: 'DELETE'
    })
  }
})
