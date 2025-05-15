<script setup lang="ts">
import type { Inbox } from '~/types'

defineProps<{
  data: Inbox[]
}>()

const toast = useToast()
const nuxtApp = useNuxtApp()
const userStore = useUserStore()
const selectedInboxId = ref()

const items = [{
  label: 'Edit inbox',
  onSelect: () => console.log('Edit inbox')
}, {
  label: 'Remove inbox',
  color: 'error' as const,
  onSelect: async () => {
    if (!userStore.selectedProject?.id) return

    await nuxtApp.$api.deleteInbox(userStore.selectedProject.id, selectedInboxId.value).then(async () => {
      toast.add({ title: 'Success', description: 'Inbox has been removed.', color: 'success' })

      await refreshInboxes()
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
            {{ item.email }}
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
        <UDropdownMenu :items="items" :content="{ align: 'end' }" @update:open="(open: boolean) => selectedInboxId = open ? item.id : null">
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
