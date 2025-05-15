<script setup lang="ts">
import * as z from 'zod'
import type { FormSubmitEvent } from '@nuxt/ui'

const toast = useToast()
const nuxtApp = useNuxtApp()
const router = useRouter()

definePageMeta({
  layout: 'not-dashboard'
})

const fields = [{
  name: 'name',
  type: 'text' as const,
  label: 'Name',
  placeholder: 'Enter the name of your project',
  required: true
}]

const schema = z.object({
  name: z.string().min(1, 'Project name is required')
})

type Schema = z.output<typeof schema>

async function onSubmit(payload: FormSubmitEvent<Schema>) {
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
    })
}
</script>

<template>
  <div class="flex flex-col h-screen items-center justify-center gap-4 p-4">
    <UPageCard class="w-full max-w-md">
      <UAuthForm
        :schema="schema"
        title="Create your first project"
        description="Enter the details to create your first project."
        icon="i-lucide-folder-plus"
        :fields="fields"
        @submit="onSubmit"
      />
    </UPageCard>
  </div>
</template>
