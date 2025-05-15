<script setup lang="ts">
import type { Project } from '~/types'

defineProps<{
  data: Project[]
}>()

const toast = useToast()
const nuxtApp = useNuxtApp()
const selectedProjectId = ref()

const items = [{
  label: 'Edit project',
  onSelect: () => console.log('Edit project')
}, {
  label: 'Remove project',
  color: 'error' as const,
  onSelect: async () => {
    await nuxtApp.$api.deleteProject(selectedProjectId.value).then(async () => {
      toast.add({ title: 'Success', description: 'Project has been removed.', color: 'success' })

      await refreshProjects()
    })
      .catch((error) => {
        toast.add({ title: 'Error', description: error.response._data.message, color: 'error' })
      })
  }
}]
</script>

<template>
  <ul role="list" class="divide-y divide-default">
    <li
      v-for="(item, index) in data"
      :key="index"
      class="flex items-center justify-between gap-3 py-3 px-4 sm:px-6"
    >
      <div class="flex items-center gap-3 min-w-0">
        <UAvatar
          src="/logo.png"
          alt="NuxtLabs"
          size="md"
        />

        <div class="text-sm min-w-0">
          <p class="text-highlighted font-medium truncate">
            {{ item.name }}
          </p>
          <p class="text-muted truncate">
            {{
              new Date(item.created_at).toLocaleDateString('en-US', {
                year: 'numeric',
                month: 'long',
                day: 'numeric'
              })
            }}
          </p>
        </div>
      </div>

      <div class="flex items-center gap-3">
        <UDropdownMenu :items="items" :content="{ align: 'end' }" @update:open="(open: boolean) => selectedProjectId = open ? item.id : null">
          <UButton
            icon="i-lucide-ellipsis-vertical"
            color="neutral"
            variant="ghost"
          />
        </UDropdownMenu>
      </div>
    </li>
  </ul>
</template>
