<script setup lang="ts">
import * as z from 'zod'
import type { FormSubmitEvent } from '@nuxt/ui'

definePageMeta({
  layout: 'not-dashboard',
  title: 'Login'
})

const toast = useToast()
const nuxtApp = useNuxtApp()
const router = useRouter()
const userStore = useUserStore()

const isLoading = ref(false)
const state = reactive({
  username: '',
  password: ''
})

const schema = z.object({
  username: z.string().min(1, 'Username is required'),
  password: z.string().min(8, 'Must be at least 8 characters')
})

const fields = [{
  name: 'username' as keyof typeof state,
  type: 'text' as const,
  label: 'Username',
  placeholder: 'Enter your username',
  required: true
}, {
  name: 'password' as keyof typeof state,
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

async function onSubmit(payload: FormSubmitEvent<any>) {
  isLoading.value = true
  await nuxtApp.$api
    .login(payload.data as z.output<typeof schema>)
    .then(async () => {
      await userStore.getUser()

      toast.add({ title: 'Success', description: 'Login successful', color: 'success' })
      router.push('/')
    })
    .catch((error) => {
      toast.add({ title: 'Error', description: error.response._data.message, color: 'error' })
    }).finally(() => {
      isLoading.value = false
    })
}
</script>

<template>
  <div class="flex flex-col h-screen items-center justify-center p-4">
    <UCard class="w-full max-w-md">
      <div class="flex flex-col text-center">
        <div class="mb-2">
          <UIcon
            name="i-lucide-user"
            class="size-8"
          />
        </div>
        <div class="text-xl text-pretty font-semibold text-highlighted">
          Login
        </div>
        <div class="mt-1 text-base text-pretty text-muted">
          Enter your credentials to access your account.
        </div>
      </div>

      <!-- OAuth Login -->
      <UButton
        v-for="provider in providers"
        :key="provider.label"
        :label="provider.label"
        :icon="provider.icon"
        :disabled="provider.disabled"
        class="w-full mt-4 inline-flex items-center justify-center"
        variant="subtle"
        color="neutral"
        size="md"
        @click="provider.onClick"
      />

      <USeparator label="or" class="my-6" />

      <!-- Simple Login -->
      <UiFormSimple
        :schema="schema"
        :state="state"
        :fields="fields"
        :is-loading="isLoading"
        @submit="onSubmit"
      />
    </UCard>
  </div>
</template>
