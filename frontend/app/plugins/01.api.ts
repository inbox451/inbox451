import type { Api } from '~/api/main'
import mainAPI from '~/api/main'

declare module '#app' {
  interface NuxtApp {
    $api: Api
  }
}

export default defineNuxtPlugin((nuxtApp) => {
  const config = useRuntimeConfig()

  const apiFetch = $fetch.create({
    credentials: 'include',
    baseURL: config.public.apiEndpoint
  })
  nuxtApp.provide('api', mainAPI(apiFetch))
})
