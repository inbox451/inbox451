<script setup lang="ts">
const nuxtApp = useNuxtApp()
const toast = useToast()
const userStore = useUserStore()

const newInboxEmail = ref('')
const isLoading = ref(false)
const isModalOpen = ref(false)

async function createInbox() {
  if (!userStore.selectedProject?.id) return

  await nuxtApp.$api.createInbox(userStore.selectedProject.id, newInboxEmail.value).then(async () => {
    toast.add({ title: 'Success', description: 'New inbox has been created.', color: 'success' })

    await refreshInboxes()
  })
    .catch((error) => {
      toast.add({ title: 'Error', description: error.response._data.message, color: 'error' })
    })
    .finally(() => {
      isModalOpen.value = false
    })

  newInboxEmail.value = ''
  isLoading.value = false
}
</script>

<template>
  <div>
    <UPageCard
      title="Inboxes"
      description="Create and manage your inboxes."
      variant="naked"
      orientation="horizontal"
      class="mb-4"
    >
      <UModal v-model:open="isModalOpen" title="Create Inbox">
        <UButton
          label="Create Inbox"
          color="neutral"
          class="w-fit lg:ms-auto"
        />

        <template #body>
          <UInput
            v-model="newInboxEmail"
            placeholder="Enter new email address"
            class="w-full "
          />
          <UButton
            label="Create"
            color="neutral"
            :disabled="!newInboxEmail"
            class="inline-flex items-center justify-center w-full mt-4"
            :loading="isLoading"
            @click="createInbox"
          />
        </template>
      </UModal>
    </UPageCard>

    <UPageCard variant="subtle" :ui="{ container: 'p-0 sm:p-0 gap-y-0', wrapper: 'items-stretch', header: 'p-4 mb-0 border-b border-default' }">
      <SettingsInboxesList :data="userStore.inboxes" />
    </UPageCard>
  </div>
</template>
