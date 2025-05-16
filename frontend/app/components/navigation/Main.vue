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
  <div
    class="relative hidden lg:flex flex-col min-h-svh min-w-16 shrink-0 border-r border-default bg-elevated/25 w-64"
  >
    <div class="h-(--ui-header-height) shrink-0 flex items-center gap-1.5 px-4">
      <ProjectsMenu />
    </div>

    <div class="flex flex-col gap-4 flex-1 overflow-y-auto px-4 py-2">
      <UNavigationMenu
        :items="links[0]"
        orientation="vertical"
      />

      <UNavigationMenu
        :items="links[1]"
        orientation="vertical"
        class="mt-auto"
      />
    </div>

    <div class="shrink-0 flex items-center gap-1.5 px-4 py-2 lg:border-t lg:border-default">
      <UserMenu />
    </div>
  </div>
</template>
