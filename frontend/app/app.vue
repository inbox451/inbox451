<script setup lang="ts">
const colorMode = useColorMode()
const userStore = useUserStore()

const color = computed(() => colorMode.value === 'dark' ? '#1b1718' : 'white')

useHead({
  meta: [
    { charset: 'utf-8' },
    { name: 'viewport', content: 'width=device-width, initial-scale=1' },
    { key: 'theme-color', name: 'theme-color', content: color }
  ],
  link: [
    { rel: 'icon', href: '/favicon.ico' }
  ],
  htmlAttrs: {
    lang: 'en'
  }
})

const title = 'inbox451'
const description = 'A simple email server that allows you to create inboxes and rules to filter emails, written in Go.'

useSeoMeta({
  title,
  description,
  ogTitle: title,
  ogDescription: description,
  // ogImage: '',
  // twitterImage: '',
  twitterCard: 'summary_large_image'
})
</script>

<template>
  <UApp>
    <NuxtLoadingIndicator />

    <div v-if="userStore.isLoggedIn === undefined">
      <UContainer class="flex flex-col items-center justify-center h-screen">
        <USkeleton class="aspect-square w-24 animate-spin" />
      </UContainer>
    </div>

    <NuxtLayout v-else>
      <NuxtPage />
    </NuxtLayout>
  </UApp>
</template>
