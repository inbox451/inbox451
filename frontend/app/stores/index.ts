import { defineStore } from 'pinia'
import type { Inbox, Project, User } from '../types'

interface State {
  isLoggedIn: boolean | undefined
  user: User | undefined
  projects: Project[]
  selectedProject?: Project
  inboxes: Inbox[]
}

export const useUserStore = defineStore('user', {
  state: (): State => ({
    isLoggedIn: undefined,
    user: undefined,
    projects: [],
    selectedProject: undefined,
    inboxes: []
  }),

  actions: {
    setIsloggedIn(isLoggedIn: boolean) {
      this.isLoggedIn = isLoggedIn
    },

    setUser(data: User) {
      this.user = data
    },

    setProjects(data: Project[]) {
      this.projects = data
    },

    setSelectedProject(data: Project) {
      this.selectedProject = data
    },

    setInboxes(data: Inbox[]) {
      this.inboxes = data
    },

    async getUser() {
      const nuxtApp = useNuxtApp()
      const router = useRouter()

      await nuxtApp.$api
        .getCurrentUser()
        .then(async (response) => {
          this.setUser(response)
          this.setIsloggedIn(true)

          // Get projects
          await refreshProjects()

          // If no projects, redirect to create project page to create the first project
          if (!this.projects.length) {
            router.push('/create-project')
          }

          // Check if is in the login page
          if (router.currentRoute.value.path === '/login') {
            router.push('/')
          }
        })
        .catch(() => {
          router.push('/login')

          this.setIsloggedIn(false)
        })
    }
  }
})
