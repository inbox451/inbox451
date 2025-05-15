<script setup lang="ts">
import type { User } from '~/types'

defineProps<{
  data: User[]
}>()

const items = [{
  label: 'Edit user',
  onSelect: () => console.log('Edit user')
}, {
  label: 'Remove user',
  color: 'error' as const,
  onSelect: () => console.log('Remove user')
}]
</script>

<template>
  <ul role="list" class="divide-y divide-default">
    <li
      v-for="(item, index) in data"
      :key="index"
      class="flex items-center justify-between gap-3 py-3 px-4 sm:px-6"
    >
      <div class="flex items-center gap-3 min-w-0">
        <UAvatar
          src="/logo.png"
          alt="NuxtLabs"
          size="md"
        />

        <div class="text-sm min-w-0">
          <p class="text-highlighted font-medium truncate">
            {{ item.name }}
          </p>
          <p class="text-muted truncate">
            {{ item.username }}
          </p>
        </div>
      </div>

      <div class="flex items-center gap-3">
        <USelect
          :model-value="item.role"
          :items="['member', 'admin']"
          color="neutral"
          :ui="{ value: 'capitalize', item: 'capitalize' }"
        />

        <UDropdownMenu :items="items" :content="{ align: 'end' }">
          <UButton
            icon="i-lucide-ellipsis-vertical"
            color="neutral"
            variant="ghost"
          />
        </UDropdownMenu>
      </div>
    </li>
  </ul>
</template>
