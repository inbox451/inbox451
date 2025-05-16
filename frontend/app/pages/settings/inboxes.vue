<script setup lang="ts">
import type { FormSubmitEvent } from '@nuxt/ui'
import * as z from 'zod'

const nuxtApp = useNuxtApp()
const toast = useToast()
const userStore = useUserStore()

const isLoading = ref(false)
const isModalOpen = ref(false)

definePageMeta({
  title: 'Inboxes'
})

const state = reactive({
  email: ''
})

const fields = [{
  name: 'email' as keyof typeof state,
  type: 'email' as const,
  label: 'Email',
  placeholder: 'Enter the e-mail address of your inbox',
  required: true
}]

const schema = z.object({
  email: z.string().email('Invalid email address')
})

async function onSubmit(payload: FormSubmitEvent<any>) {
  if (!userStore.selectedProject?.id) return

  await nuxtApp.$api.createInbox(userStore.selectedProject.id, (payload.data as z.output<typeof schema>).email).then(async () => {
    toast.add({ title: 'Success', description: 'New inbox has been created.', color: 'success' })

    await refreshInboxes()
  })
    .catch((error) => {
      toast.add({ title: 'Error', description: error.response._data.message, color: 'error' })
    })
    .finally(() => {
      isLoading.value = false
      isModalOpen.value = false
    })
}
</script>

<template>
  <div>
    <UiCardHeader
      title="Inboxes"
      description="Create and manage your inboxes."
    >
      <UModal v-model:open="isModalOpen" title="Create Inbox">
        <UButton
          label="Create Inbox"
          color="neutral"
          class="w-fit lg:ms-auto"
        />

        <template #body>
          <UiFormSimple
            button-label="Create"
            :schema="schema"
            :state="state"
            :fields="fields"
            :is-loading="isLoading"
            @submit="onSubmit"
          />
        </template>
      </UModal>
    </UiCardHeader>

    <UCard
      variant="subtle"
      :ui="{
        body: 'p-0 sm:p-0'
      }"
    >
      <SettingsInboxesList :data="userStore.inboxes" />
    </UCard>
  </div>
</template>
