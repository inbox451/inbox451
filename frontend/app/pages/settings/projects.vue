<script setup lang="ts">
import type { FormSubmitEvent } from '@nuxt/ui'
import * as z from 'zod'

definePageMeta({
  title: 'Projects'
})

const nuxtApp = useNuxtApp()
const toast = useToast()
const userStore = useUserStore()

const isLoading = ref(false)
const isModalOpen = ref(false)

const state = reactive({
  name: ''
})

const fields = [{
  name: 'name' as keyof typeof state,
  type: 'text' as const,
  label: 'Name',
  placeholder: 'Enter the name of your project',
  required: true
}]

const schema = z.object({
  name: z.string().min(1, 'Project name is required')
})

async function onSubmit(payload: FormSubmitEvent<any>) {
  await nuxtApp.$api.createProject((payload.data as z.output<typeof schema>).name).then(async () => {
    toast.add({ title: 'Success', description: 'New project has been created.', color: 'success' })

    await refreshProjects()
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
      <SettingsProjectsList :data="userStore.projects" />
    </UCard>
  </div>
</template>
