export default defineNuxtPlugin(() => {
  if (import.meta.client) {
    const userStore = useUserStore()
    userStore.getUser()
  }
})
