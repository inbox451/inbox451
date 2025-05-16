<script setup lang="ts">
definePageMeta({
  title: 'Settings'
})

const links = [[{
  label: 'General',
  icon: 'i-lucide-user',
  to: '/settings',
  exact: true
}, {
  label: 'Projects',
  icon: 'i-lucide-folder',
  to: '/settings/projects'
},
{
  label: 'Inboxes',
  icon: 'i-lucide-inbox',
  to: '/settings/inboxes'
}, {
  label: 'Users',
  icon: 'i-lucide-users',
  to: '/settings/users'
}, {
  label: 'Security',
  icon: 'i-lucide-shield',
  to: '/settings/security'
}]]
</script>

<template>
  <div class="dashboard-panel flex-1">
    <NavigationHeader />
    <NavigationTabs :links="links" />

    <div class="flex flex-col gap-4 sm:gap-6 lg:gap-12 w-full overflow-y-auto p-4 sm:p-6 lg:py-12 lg:max-w-2xl mx-auto">
      <NuxtPage />
    </div>
  </div>
</template>
