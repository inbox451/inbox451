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
  name: 'username',
  type: 'text' as const,
  label: 'Username',
  placeholder: 'Enter your username',
  required: true
}, {
  name: 'password',
  label: 'Password',
  type: 'password' as const,
  placeholder: 'Enter your password'
}]

const providers = [{
  label: 'Google',
  icon: 'i-simple-icons-google',
  disabled: true,
  onClick: () => {
    toast.add({ title: 'Google', description: 'Login with Google' })
  }
}, {
  label: 'GitHub',
  icon: 'i-simple-icons-github',
  disabled: true,
  onClick: () => {
    toast.add({ title: 'GitHub', description: 'Login with GitHub' })
  }
}]

const schema = z.object({
  username: z.string().min(1, 'Username is required'),
  password: z.string().min(8, 'Must be at least 8 characters')
})

type Schema = z.output<typeof schema>

async function onSubmit(payload: FormSubmitEvent<Schema>) {
  await nuxtApp.$api
    .login(payload.data)
    .then(() => {
      toast.add({ title: 'Success', description: 'Login successful', color: 'success' })
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
        title="Login"
        description="Enter your credentials to access your account."
        icon="i-lucide-user"
        :fields="fields"
        :providers="providers"
        @submit="onSubmit"
      />
    </UPageCard>
  </div>
</template>
