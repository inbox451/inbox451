<script setup lang="ts">
import { format } from 'date-fns'
import type { Messages } from '~/types'

const props = defineProps<{
  mail: Messages
}>()

const emit = defineEmits(['close', 'refresh'])
const nuxtApp = useNuxtApp()
const userStore = useUserStore()
const route = useRoute()
const toast = useToast()

async function deleteMessage() {
  if (!userStore.selectedProject?.id) return

  await nuxtApp.$api.deleteMessage(userStore.selectedProject.id, String(route.params.id), props.mail.id)
  emit('refresh')
  emit('close')

  toast.add({
    title: 'Success',
    description: 'Message deleted successfully.',
    color: 'success'
  })
}

async function markAsUnread() {
  if (!userStore.selectedProject?.id) return

  if (props.mail.is_read) {
    await nuxtApp.$api.markAsRead(userStore.selectedProject.id, String(route.params.id), props.mail.id)
  } else {
    await nuxtApp.$api.markAsUnread(userStore.selectedProject.id, String(route.params.id), props.mail.id)
  }

  toast.add({
    title: 'Success',
    description: `Message marked as ${props.mail.is_read ? 'unread' : 'read'} successfully.`,
    color: 'success'
  })

  emit('refresh')
}
</script>

<template>
  <UDashboardPanel id="inbox-2">
    <UDashboardNavbar :title="mail.subject" :toggle="false">
      <template #leading>
        <UButton
          icon="i-lucide-x"
          color="neutral"
          variant="ghost"
          class="-ms-1.5"
          @click="emit('close')"
        />
      </template>

      <template #right>
        <UTooltip :text="props.mail.is_read ? 'Mark as read' : 'Mark as unread'">
          <UButton
            :icon="props.mail.is_read ? 'i-lucide-eye-off' : 'i-lucide-eye'"
            color="neutral"
            variant="ghost"
            @click="markAsUnread"
          />
        </UTooltip>

        <UTooltip text="Delete">
          <UButton
            icon="i-lucide-trash-2"
            color="neutral"
            variant="ghost"
            @click="deleteMessage"
          />
        </UTooltip>
      </template>
    </UDashboardNavbar>

    <div class="flex flex-col sm:flex-row justify-between items-center gap-1 p-4 sm:px-6 border-b border-default">
      <div class="flex items-center gap-4 sm:my-1.5">
        <UAvatar
          src="/logo.png"
          alt="NuxtLabs"
          size="3xl"
        />

        <div class="min-w-0">
          <p class="font-semibold text-highlighted">
            {{ mail.sender }}
          </p>
        </div>
      </div>

      <p class="max-sm:pl-16 text-muted text-sm ">
        {{ format(new Date(mail.created_at), 'dd MMM HH:mm') }}
      </p>
    </div>

    <div class="flex-1 p-4 sm:p-6 overflow-y-auto">
      <p class="whitespace-pre-wrap">
        {{ mail.body }}
      </p>
    </div>
  </UDashboardPanel>
</template>
