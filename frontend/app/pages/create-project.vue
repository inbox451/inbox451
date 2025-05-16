<script setup lang="ts">
import * as z from 'zod'
import type { FormSubmitEvent } from '@nuxt/ui'

const toast = useToast()
const nuxtApp = useNuxtApp()
const router = useRouter()
const userStore = useUserStore()

definePageMeta({
  layout: 'not-dashboard',
  title: 'Create Project'
})

type StateKeys = keyof typeof state
type Schema = z.output<typeof schema>

const isLoading = ref(false)
const state = reactive({
  name: ''
})

const fields = [{
  name: 'name' as StateKeys,
  type: 'text' as const,
  label: 'Name',
  placeholder: 'Enter the name of your project',
  required: true
}]

const schema = z.object({
  name: z.string().min(1, 'Project name is required')
})

async function onSubmit(payload: FormSubmitEvent<Schema>) {
  isLoading.value = true
  await nuxtApp.$api
    .createProject(payload.data.name)
    .then(async () => {
      // Refresh the projects list in the store
      await refreshProjects()

      toast.add({ title: 'Success', description: 'Project created successfully', color: 'success' })
      router.push('/')
    })
    .catch((error) => {
      toast.add({ title: 'Error', description: error.response._data.message, color: 'error' })
    }).finally(() => {
      isLoading.value = false
    })
}

onMounted(() => {
  // Check if the user has projects
  if (userStore.projects.length)
    router.push('/')
})
</script>

<template>
  <div class="flex flex-col h-screen items-center justify-center p-4">
    <UCard class="w-full max-w-md">
      <div class="flex flex-col text-center">
        <div class="mb-2">
          <UIcon
            name="i-lucide-folder-plus"
            class="size-8"
          />
        </div>
        <div class="text-xl text-pretty font-semibold text-highlighted">
          Create your first project
        </div>
        <div class="mt-1 text-base text-pretty text-muted">
          Enter the details to create your first project.
        </div>
      </div>

      <UForm
        :schema="schema"
        :state="state"
        class="space-y-4 mt-6"
        @submit="onSubmit"
      >
        <UFormField
          v-for="field in fields"
          :key="field.name"
          :label="field.label"
          :name="field.name"
          :type="field.type"
          :placeholder="field.placeholder"
          :required="field.required"
        >
          <UInput
            v-model="state[field.name]"
            :type="field.type"
            :placeholder="field.placeholder"
            class="w-full"
          />
        </UFormField>

        <UButton
          type="submit"
          class="inline-flex w-full items-center justify-center"
          :loading="isLoading"
        >
          <UIcon
            name="i-lucide-plus"
            class="me-2"
          />
          Continue
        </UButton>
      </UForm>
    </UCard>
  </div>
</template>
