<script setup lang="ts">
const userStore = useUserStore()

const projects = computed(() => {
  return userStore.projects.map(project => ({
    id: project.id,
    label: project.name,
    avatar: {
      src: '/logo.png',
      alt: 'NuxtLabs'
    }

  }))
})

// TODO: Save and pick from localStorage the last selected project
const selectedProject = ref(projects.value[0])

const items = computed(() => {
  return [projects.value.map(project => ({
    ...project,
    onSelect() {
      selectedProject.value = project
    }
  })), [{
    label: 'Manage Projects',
    icon: 'i-lucide-cog',
    onSelect() {
      navigateTo('/settings/projects')
    }
  }]]
})

watch(() => selectedProject.value, async (newVal, oldVal) => {
  if (!newVal || newVal === oldVal) return

  const originalProject = userStore.projects.find(t => t.id === newVal.id)
  if (originalProject) {
    userStore.setSelectedProject(originalProject)
  }

  await refreshInboxes()
})
</script>

<template>
  <UDropdownMenu
    :items="items"
    :content="{ align: 'center', collisionPadding: 12 }"
    :ui="{ content: 'w-(--reka-dropdown-menu-trigger-width)' }"
  >
    <UButton
      v-bind="{
        ...selectedProject,
        label: selectedProject?.label,
        trailingIcon: 'i-lucide-chevrons-up-down'
      }"
      color="neutral"
      variant="ghost"
      block
      class="data-[state=open]:bg-elevated"
      :ui="{
        trailingIcon: 'text-dimmed'
      }"
    />
  </UDropdownMenu>
</template>
