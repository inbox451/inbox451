<script setup lang="ts">
definePageMeta({
  title: 'Projects'
})

const nuxtApp = useNuxtApp()
const toast = useToast()
const userStore = useUserStore()

const newProjectName = ref('')
const isLoading = ref(false)
const isModalOpen = ref(false)

async function createProject() {
  await nuxtApp.$api.createProject(newProjectName.value).then(async () => {
    toast.add({ title: 'Success', description: 'New project has been created.', color: 'success' })

    await refreshProjects()
  })
    .catch((error) => {
      toast.add({ title: 'Error', description: error.response._data.message, color: 'error' })
    })
    .finally(() => {
      isModalOpen.value = false
    })

  newProjectName.value = ''
  isLoading.value = false
}
</script>

<template>
  <div>
    <UiCardHeader
      title="Projects"
      description="Create and manage your projects."
    >
      <UModal v-model:open="isModalOpen" title="Create Project">
        <UButton
          label="Create Project"
          color="neutral"
          class="w-fit lg:ms-auto"
        />

        <template #body>
          <UInput
            v-model="newProjectName"
            placeholder="Enter new project name"
            class="w-full "
          />
          <UButton
            label="Create"
            color="neutral"
            :disabled="!newProjectName"
            class="inline-flex items-center justify-center w-full mt-4"
            :loading="isLoading"
            @click="createProject"
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
      <SettingsProjectsList :data="userStore.projects" />
    </UCard>
  </div>
</template>
