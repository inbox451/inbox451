<script setup lang="ts">
const nuxtApp = useNuxtApp()
const toast = useToast()
const userStore = useUserStore()

const newInboxEmail = ref('')
const isLoading = ref(false)
const isModalOpen = ref(false)

definePageMeta({
  title: 'Inboxes'
})

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
