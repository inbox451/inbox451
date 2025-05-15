<script setup lang="ts">
const nuxtApp = useNuxtApp()

definePageMeta({
  layout: 'not-dashboard'
})

onMounted(async () => {
  await nuxtApp.$api
    .logout()
    .finally(() => {
      window.location.href = '/login'
    })
})
</script>

<template>
  <div />
</template>
