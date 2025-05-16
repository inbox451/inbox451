<script setup lang="ts">
import type { FormSubmitEvent } from '@nuxt/ui'
import type * as z from 'zod'

const props = defineProps<{
  buttonLabel?: string
  buttonIcon?: string
  schema: z.ZodObject<{
    [x: string]: z.ZodTypeAny
  }>
  fields: {
    name: string
    type: string
    label: string
    placeholder: string
    required?: boolean
  }[]
  isLoading?: boolean
}>()

const emit = defineEmits<{
  (e: 'submit', payload: FormSubmitEvent<z.infer<typeof props.schema>>): void
}>()

const state = defineModel<Partial<Record<string, string>>>('state', { default: {} })
</script>

<template>
  <UForm
    :schema="schema"
    :state="state"
    class="space-y-4 "
    @submit="emit('submit', $event)"
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
        :name="buttonIcon || 'i-lucide-arrow-right'"
        class="me-2"
      />
      {{ buttonLabel || 'Continue' }}
    </UButton>
  </UForm>
</template>
