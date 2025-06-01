<script setup lang="ts">
import type { NuxtError } from '#app'

defineProps<{
  error: NuxtError
}>()

useSeoMeta({
  title: 'Page not found',
  description: 'We are sorry but this page could not be found.'
})

useHead({
  htmlAttrs: {
    lang: 'en'
  }
})
</script>

<template>
  <UApp>
    <main class="min-h-[calc(100vh-var(--ui-header-height))] flex flex-col items-center justify-center text-center">
      <p class="text-base font-semibold text-primary">
        {{ error.statusCode }}
      </p>
      <h1 class="mt-2 text-4xl sm:text-5xl font-bold text-highlighted text-balance">
        {{ error.message }}
      </h1>

      <div class="mt-8 flex items-center justify-center gap-6">
        <UButton>Back to Home</UButton>
      </div>
    </main>
  </UApp>
</template>
