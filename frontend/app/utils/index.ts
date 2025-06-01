import type { Project } from '~/types'

export async function refreshProjects() {
  const nuxtApp = useNuxtApp()
  const userStore = useUserStore()
  const toast = useToast()

  await nuxtApp.$api.getProjects().then(async (res) => {
    userStore.setProjects(res.data)

    // TODO - Improve this
    if (res.data.length > 0) {
      userStore.setSelectedProject(res.data[0] as Project)
      await refreshInboxes()
    }
  })
    .catch((error) => {
      toast.add({ title: 'Error', description: error.response._data.message, color: 'error' })
    })
}

export async function refreshInboxes() {
  const nuxtApp = useNuxtApp()
  const userStore = useUserStore()
  const toast = useToast()

  if (!userStore.selectedProject?.id) return

  await nuxtApp.$api.getInboxes(userStore.selectedProject.id).then((res) => {
    userStore.setInboxes(res.data)
  })
    .catch((error) => {
      toast.add({ title: 'Error', description: error.response._data.message, color: 'error' })
    })
}
