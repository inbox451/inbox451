<script setup lang="ts">
const userStore = useUserStore()
const open = ref(false)

const links = computed(() => [[{
  label: 'Inbox',
  icon: 'i-lucide-inbox',
  defaultOpen: true,
  onSelect: () => {
    open.value = false
  },
  children:
    userStore.inboxes.length
      ? userStore.inboxes.map(inbox => ({
          label: inbox.email,
          to: `/inbox/${inbox.id}`,
          onSelect: () => {
            open.value = false
          }
        }))
      : [{
          label: 'Create inbox',
          icon: 'i-lucide-plus',
          to: '/settings/inboxes',
          onSelect: () => {
            open.value = false
          }
        }]
},
{
  label: 'Settings',
  to: '/settings',
  icon: 'i-lucide-settings',
  children: [{
    label: 'General',
    to: '/settings',
    exact: true,
    onSelect: () => {
      open.value = false
    }
  },
  {
    label: 'Projects',
    to: '/settings/projects',
    onSelect: () => {
      open.value = false
    }
  },
  {
    label: 'Inboxes',
    to: '/settings/inboxes',
    onSelect: () => {
      open.value = false
    }
  },
  {
    label: 'Users',
    to: '/settings/users',
    onSelect: () => {
      open.value = false
    }
  },
  {
    label: 'Security',
    to: '/settings/security',
    onSelect: () => {
      open.value = false
    }
  }]
}],
[{
  label: 'GitHub',
  icon: 'i-lucide-github',
  to: 'https://github.com/inbox451/inbox451',
  target: '_blank'
}]])
</script>

<template>
  <UDashboardGroup unit="rem">
    <UDashboardSidebar
      id="default"
      v-model:open="open"
      collapsible
      resizable
      class="bg-elevated/25"
      :ui="{ footer: 'lg:border-t lg:border-default' }"
    >
      <template #header="{ collapsed }">
        <ProjectsMenu :collapsed="collapsed" />
      </template>

      <template #default="{ collapsed }">
        <UNavigationMenu
          :collapsed="collapsed"
          :items="links[0]"
          orientation="vertical"
        />

        <UNavigationMenu
          :collapsed="collapsed"
          :items="links[1]"
          orientation="vertical"
          class="mt-auto"
        />
      </template>

      <template #footer="{ collapsed }">
        <UserMenu :collapsed="collapsed" />
      </template>
    </UDashboardSidebar>

    <slot />
  </UDashboardGroup>
</template>
