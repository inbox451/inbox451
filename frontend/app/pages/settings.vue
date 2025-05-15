<script setup lang="ts">
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
  <UDashboardPanel id="settings" :ui="{ body: 'lg:py-12' }">
    <template #header>
      <UDashboardNavbar title="Settings">
        <template #leading>
          <UDashboardSidebarCollapse />
        </template>
      </UDashboardNavbar>

      <UDashboardToolbar>
        <!-- NOTE: The `-mx-1` class is used to align with the `DashboardSidebarCollapse` button here. -->
        <UNavigationMenu :items="links" highlight class="-mx-1 flex-1" />
      </UDashboardToolbar>
    </template>

    <template #body>
      <div class="flex flex-col gap-4 sm:gap-6 lg:gap-12 w-full lg:max-w-2xl mx-auto">
        <NuxtPage />
      </div>
    </template>
  </UDashboardPanel>
</template>
